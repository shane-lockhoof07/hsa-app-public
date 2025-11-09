import base64
import io
import json
import os
import re
import tempfile
from typing import Optional

import httpx
import pdf2image
import pillow_heif
import pytesseract
from fastapi import FastAPI, File, HTTPException, UploadFile
from PIL import Image

app = FastAPI()
pillow_heif.register_heif_opener()

# Configuration
USE_CLAUDE = os.getenv("USE_CLAUDE_API", "true").lower() == "true"
CLAUDE_API_KEY = os.getenv("CLAUDE_API_KEY", "")
CLAUDE_MODEL = os.getenv("CLAUDE_MODEL", "claude-3-5-haiku-20241022")


async def extract_with_claude(image_data: bytes, mime_type: str) -> dict:
    """Extract receipt data using Claude API with vision - assumes full HSA qualification"""

    if not CLAUDE_API_KEY:
        raise HTTPException(status_code=500, detail="Claude API key not configured")

    # Encode image to base64
    image_b64 = base64.standard_b64encode(image_data).decode("utf-8")

    # Prepare Claude API request
    headers = {
        "x-api-key": CLAUDE_API_KEY,
        "anthropic-version": "2023-06-01",
        "content-type": "application/json",
    }

    prompt = """Analyze this receipt image and extract the following information in JSON format:
{
  "vendor": "store name",
  "amount": 0.00,
  "date": "MM/DD/YYYY",
  "items": ["item1", "item2"],
  "raw_text": "full receipt text"
}

IMPORTANT INSTRUCTIONS:
- Extract the TOTAL amount from the receipt (after tax, all discounts applied)
- This should be the final amount paid
- Date should be in MM/DD/YYYY format
- List all visible items from the receipt in the items array
- Provide the complete text from the receipt in raw_text
- Do NOT try to determine if items are HSA-qualified - the user will do that manually

Example response:
{
  "vendor": "Walgreens",
  "amount": 24.99,
  "date": "11/07/2025",
  "items": ["Band-Aids", "Ibuprofen", "Paper towels"],
  "raw_text": "WALGREENS\\nStore #1234\\n..."
}"""

    payload = {
        "model": CLAUDE_MODEL,
        "max_tokens": 1024,
        "messages": [
            {
                "role": "user",
                "content": [
                    {
                        "type": "image",
                        "source": {
                            "type": "base64",
                            "media_type": mime_type,
                            "data": image_b64,
                        },
                    },
                    {"type": "text", "text": prompt},
                ],
            }
        ],
    }

    async with httpx.AsyncClient(timeout=30.0) as client:
        response = await client.post(
            "https://api.anthropic.com/v1/messages",
            headers=headers,
            json=payload,
        )

        print(f"Claude API Status Code: {response.status_code}")
        print(f"Claude API Response: {response.text[:500]}")

        if response.status_code != 200:
            raise HTTPException(
                status_code=response.status_code,
                detail=f"Claude API error: {response.text}",
            )

        result = response.json()
        text_content = result["content"][0]["text"]

        # Parse JSON from response
        json_match = re.search(r"```json\n(.*?)\n```", text_content, re.DOTALL)
        if json_match:
            json_str = json_match.group(1)
        else:
            json_match = re.search(r"\{.*\}", text_content, re.DOTALL)
            json_str = json_match.group(0) if json_match else text_content

        try:
            parsed_data = json.loads(json_str)

            # Always assume fully HSA-qualified - user will adjust if needed
            parsed_data["hsa_status"] = "Yes"
            parsed_data["hsa_qualified"] = True

            return parsed_data
        except json.JSONDecodeError:
            # Fallback parsing
            return {
                "vendor": extract_field(text_content, "vendor"),
                "amount": float(extract_field(text_content, "amount", "0")),
                "date": extract_field(text_content, "date"),
                "hsa_status": "Yes",
                "hsa_qualified": True,
                "raw_text": text_content,
            }


def extract_field(text: str, field: str, default: str = "") -> str:
    """Extract a field value from text"""
    pattern = rf'"{field}":\s*"([^"]*)"'
    match = re.search(pattern, text)
    return match.group(1) if match else default


def extract_with_tesseract(image: Image.Image) -> dict:
    """Extract receipt data using Tesseract OCR (free/local) - basic fallback"""
    text = pytesseract.image_to_string(image)

    vendor = re.search(r"(Walgreens|CVS|Costco|Walmart|Target|Kroger|Safeway|Rite Aid)", text, re.I)
    amount = re.search(r"\$?\s*([0-9]+[,.]?[0-9]*\.[0-9]{2})", text)
    date = re.search(r"(\d{1,2}[-/]\d{1,2}[-/]\d{2,4})", text)

    parsed_amount = 0.0
    if amount:
        amount_str = amount.group(1).replace(",", "")
        try:
            parsed_amount = float(amount_str)
        except ValueError:
            parsed_amount = 0.0

    parsed_date = None
    if date:
        date_str = date.group(1)
        parts = re.split(r"[-/]", date_str)
        if len(parts) == 3:
            month, day, year = parts
            if len(year) == 2:
                year = "20" + year
            parsed_date = f"{month.zfill(2)}/{day.zfill(2)}/{year}"

    return {
        "vendor": vendor.group(1) if vendor else "Unknown",
        "amount": parsed_amount,
        "date": parsed_date,
        "hsa_status": "Yes",  # Always assume qualified
        "hsa_qualified": True,
        "raw_text": text,
    }


@app.post("/parse")
async def parse_receipt(file: UploadFile = File(...)):
    """Parse receipt using Claude API or Tesseract based on configuration"""

    content = await file.read()
    mime_type = file.content_type or "image/jpeg"

    # Determine correct MIME type
    if file.filename:
        ext = file.filename.lower()
        if ext.endswith(".png"):
            mime_type = "image/png"
        elif ext.endswith((".jpg", ".jpeg")):
            mime_type = "image/jpeg"
        elif ext.endswith(".gif"):
            mime_type = "image/gif"
        elif ext.endswith(".webp"):
            mime_type = "image/webp"
        elif ext.endswith(".pdf"):
            mime_type = "application/pdf"
        elif ext.endswith((".heic", ".heif")):
            mime_type = "image/heic"

    # Handle HEIC conversion
    if mime_type == "image/heic" or (file.filename and file.filename.lower().endswith((".heic", ".heif"))):
        try:
            image = Image.open(io.BytesIO(content))
            buffer = io.BytesIO()
            image.save(buffer, format="JPEG")
            content = buffer.getvalue()
            mime_type = "image/jpeg"
        except Exception as e:
            raise HTTPException(status_code=400, detail=f"Failed to convert HEIC: {str(e)}")

    # Handle PDF
    if mime_type == "application/pdf":
        try:
            with tempfile.NamedTemporaryFile(suffix=".pdf", delete=False) as tmp:
                tmp.write(content)
                tmp_path = tmp.name

            images = pdf2image.convert_from_path(tmp_path)
            os.unlink(tmp_path)

            if not images:
                raise HTTPException(status_code=400, detail="No pages found in PDF")

            # Convert first page to bytes
            buffer = io.BytesIO()
            images[0].save(buffer, format="JPEG")
            content = buffer.getvalue()
            mime_type = "image/jpeg"
        except Exception as e:
            raise HTTPException(status_code=400, detail=f"Failed to process PDF: {str(e)}")

    # Extract data
    if USE_CLAUDE:
        result = await extract_with_claude(content, mime_type)
    else:
        image = Image.open(io.BytesIO(content))
        result = extract_with_tesseract(image)

    return result


@app.get("/health")
async def health():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "ocr_method": "claude" if USE_CLAUDE else "tesseract",
        "model": CLAUDE_MODEL if USE_CLAUDE else "tesseract",
    }
