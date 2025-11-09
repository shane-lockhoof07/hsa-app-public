# HSA Receipt Management API

A Go-based REST API for managing Health Savings Account (HSA) receipts with intelligent OCR processing and duplicate detection.

## Features

- **Receipt Upload & OCR Processing**: Upload receipt images (JPG, PNG, PDF, HEIC) with automatic text extraction and HSA qualification detection
- **Duplicate Detection**: Prevents duplicate receipts using both image hash comparison and vendor/amount/date matching
- **Smart Receipt Organization**: Automatically organizes receipts into year-based folders with unused/used classification
- **Subset Sum Algorithm**: Find optimal receipt combinations to match target HSA reimbursement amounts
- **Full CRUD Operations**: Create, read, update, and delete receipts with complete metadata
- **File Management**: Automatic file movement between unused/used directories based on receipt status

## Architecture

```
api-go/
├── config/
│   └── config.go          # Configuration management
├── internal/
│   ├── db.go              # Database operations & migrations
│   ├── handlers.go        # HTTP request handlers
│   ├── hash.go            # Image hashing utilities
│   ├── models.go          # Data models & constants
│   └── subset_sum.go      # Receipt combination algorithm
├── migrations/
│   └── 001_init.sql       # Database schema
├── Dockerfile             # Container definition
├── go.mod                 # Go dependencies
├── go.sum                 # Dependency checksums
├── main.go                # Application entry point
└── README.md              # This file
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- OCR service (separate service required)

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:password@localhost:5432/hsa_app?sslmode=disable` |
| `OCR_SERVICE_URL` | URL of the OCR processing service | `http://localhost:8001` |
| `HSA_DIR` | Base directory for receipt storage | `/data/hsa` |
| `PORT` | Server port | `8080` |
| `CLAUDE_API_KEY` | API key for Claude AI (if using) | `""` |
| `CLAUDE_MODEL` | Claude model identifier | `claude-3-5-haiku-20241022` |

## Installation

### Using Docker (Recommended)

```bash
# Build the image
docker build -t hsa-api .

# Run the container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/hsa_app" \
  -e OCR_SERVICE_URL="http://ocr-service:8001" \
  -v /path/to/receipts:/data/hsa \
  hsa-api
```

### Local Development

```bash
# Install dependencies
go mod download

# Run migrations (automatic on startup)
# Ensure PostgreSQL is running and DATABASE_URL is set

# Run the application
go run main.go
```

## API Endpoints

### Health Check
```
GET /api/health
```
Returns service status.

### Upload Receipt
```
POST /api/receipts/upload
Content-Type: multipart/form-data

file: <image file>
```
Uploads a receipt image, processes it via OCR, and stores metadata.

**Response:**
```json
{
  "id": 123,
  "vendor": "CVS Pharmacy",
  "amount": 45.67,
  "date": "2025-01-15",
  "hsa_qualified": true,
  "hsa_status": "Yes",
  "message": "Receipt uploaded successfully"
}
```

### List All Receipts
```
GET /api/receipts
```
Returns all receipts for the household user.

### Get Receipt by ID
```
GET /api/receipts/{id}
```
Returns a specific receipt by ID.

### Update Receipt
```
PUT /api/receipts/{id}
Content-Type: application/json

{
  "vendor": "Updated Vendor",
  "total_amount": 50.00,
  "date": "2025-01-15",
  "hsa_status": "Yes",
  "used": true,
  "use_reason": "Q1 2025 reimbursement"
}
```
Updates receipt metadata and automatically moves files between unused/used directories.

### Delete Receipt
```
DELETE /api/receipts/{id}
```
Deletes a receipt and its associated file.

### Find Receipt Combinations
```
POST /api/receipts/deduct
Content-Type: application/json

{
  "user_id": "household",
  "amount": 150.00
}
```
Returns optimal combination of unused receipts that sum closest to the target amount.

### Serve Receipt File
```
GET /receipts/file/{id}
```
Serves the actual receipt image file.

## Receipt Storage Structure

Receipts are automatically organized in the following structure:

```
/data/hsa/
├── unused/
│   ├── 2024/
│   │   └── 1234567890_receipt.jpg
│   └── 2025/
│       └── 1234567891_receipt.pdf
└── used/
    ├── 2024/
    │   └── 1234567892_receipt.jpg
    └── 2025/
        └── 1234567893_receipt.png
```

When a receipt's `used` status changes, it's automatically moved to the appropriate directory.

## HSA Status Values

- **`Yes`**: Fully HSA-qualified
- **`No`**: Not HSA-qualified
- **`Partially`**: Partially HSA-qualified (total_amount reflects only qualified portion)

## Database Schema

The application uses a single `receipts` table with the following key fields:

- `id`: Primary key
- `user_id`: User identifier (defaults to "household")
- `vendor`: Merchant name
- `total_amount`: Receipt amount (HSA-qualified portion only)
- `date`: Receipt date
- `hsa_status`: Qualification status (Yes/No/Partially)
- `image_path`: File system path to receipt image
- `image_hash`: SHA-256 hash for duplicate detection
- `used`: Whether receipt has been used for reimbursement
- `used_date`: When receipt was marked as used
- `use_reason`: Optional reason for using receipt

## Duplicate Detection

The API prevents duplicates using two methods:

1. **Image Hash**: SHA-256 hash of file contents
2. **Data Matching**: Vendor name + amount + date combination

If a duplicate is detected, the API returns a `409 Conflict` status with details of the existing receipt.

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o hsa-api main.go
```

## Security Considerations

- Always use environment variables for sensitive configuration
- Never commit API keys or passwords to version control
- Use HTTPS in production
- Implement proper authentication/authorization (not included in this version)
- Ensure database connections use SSL in production (`sslmode=require`)
