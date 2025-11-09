package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hsa-app/config"
	"hsa-app/internal"
)

type OCRResponse struct {
	Vendor       string  `json:"vendor"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	HSAQualified bool    `json:"hsa_qualified"`
	HSAStatus    string  `json:"hsa_status"` // "Yes", "No", or "Partially"
	RawText      string  `json:"raw_text"`
}

func main() {
	cfg := config.Load()

	log.Println("Starting HSA Receipt Management System")
	log.Printf("Configuration loaded:")
	log.Printf("  - Database: %s", maskConnectionString(cfg.DatabaseURL))
	log.Printf("  - OCR Service: %s", cfg.OCRServiceURL)
	log.Printf("  - HSA Directory: %s", cfg.HSADir)
	log.Printf("  - Port: %s", cfg.Port)
	log.Printf("  - Claude Model: %s", cfg.ClaudeModel)

	db, err := internal.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Run migrations on startup
	if err := db.RunMigrations("./migrations"); err != nil {
		log.Printf("Warning: Failed to run migrations: %v", err)
	}

	// Create base HSA directory and subdirectories
	if err := os.MkdirAll(cfg.HSADir, 0755); err != nil {
		log.Printf("Warning: Could not create HSA directory: %v", err)
	}
	
	// Create unused and used directories
	unusedDir := filepath.Join(cfg.HSADir, "unused")
	usedDir := filepath.Join(cfg.HSADir, "used")
	if err := os.MkdirAll(unusedDir, 0755); err != nil {
		log.Printf("Warning: Could not create unused directory: %v", err)
	}
	if err := os.MkdirAll(usedDir, 0755); err != nil {
		log.Printf("Warning: Could not create used directory: %v", err)
	}

	server := &internal.Server{
		DB:            db,
		OCRServiceURL: cfg.OCRServiceURL,
		ReceiptDir:    cfg.HSADir,
	}

	http.HandleFunc("/api/health", HealthHandler)
	http.HandleFunc("/api/receipts/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(w, r, server)
	})
	http.HandleFunc("/api/receipts/deduct", server.DeductHandler)
	http.HandleFunc("/api/receipts", func(w http.ResponseWriter, r *http.Request) {
		ReceiptsHandler(w, r, server)
	})
	http.HandleFunc("/api/receipts/", func(w http.ResponseWriter, r *http.Request) {
		ReceiptByIDHandler(w, r, server)
	})
	http.HandleFunc("/receipts/file/", func(w http.ResponseWriter, r *http.Request) {
		ServeReceiptFile(w, r, server)
	})

	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, enableCORS(http.DefaultServeMux)))
}

// getReceiptFilePath constructs the file path based on receipt status and year
func getReceiptFilePath(baseDir string, year int, filename string, used bool) string {
	var subdir string
	if used {
		subdir = filepath.Join("used", fmt.Sprintf("%d", year))
	} else {
		subdir = filepath.Join("unused", fmt.Sprintf("%d", year))
	}

	fullDir := filepath.Join(baseDir, subdir)
	return filepath.Join(fullDir, filename)
}

// ensureDirectoryExists creates directory structure if it doesn't exist
func ensureDirectoryExists(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

// moveReceiptFile moves a receipt file between unused and used directories
func moveReceiptFile(oldPath, baseDir string, year int, filename string, toUsed bool) (string, error) {
	newPath := getReceiptFilePath(baseDir, year, filename, toUsed)

	// If the paths are the same, no need to move
	if oldPath == newPath {
		return newPath, nil
	}

	// Ensure the destination directory exists
	if err := ensureDirectoryExists(newPath); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Try to move the file using os.Rename first (fastest)
	if err := os.Rename(oldPath, newPath); err != nil {
		// If rename fails (possibly cross-device), do copy + delete
		log.Printf("os.Rename failed, using copy+delete: %v", err)
		
		// Copy the file
		src, err := os.Open(oldPath)
		if err != nil {
			return "", fmt.Errorf("failed to open source file: %v", err)
		}
		defer src.Close()

		dst, err := os.Create(newPath)
		if err != nil {
			return "", fmt.Errorf("failed to create destination file: %v", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return "", fmt.Errorf("failed to copy file: %v", err)
		}

		// Delete the original file after successful copy
		if err := os.Remove(oldPath); err != nil {
			log.Printf("Warning: Failed to delete original file after copy: %v", err)
			// Don't fail the operation, the file was at least copied
		}
	}

	return newPath, nil
}
func maskConnectionString(connStr string) string {
	if idx := strings.Index(connStr, "@"); idx > 0 {
		return "postgres://****:****" + connStr[idx:]
	}
	return connStr
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "hsa-receipt-api",
	})
}

func UploadHandler(w http.ResponseWriter, r *http.Request, s *internal.Server) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	hash := sha256.Sum256(fileData)
	imageHash := hex.EncodeToString(hash[:])

	if existingReceipt, err := s.DB.GetReceiptByImageHash(imageHash); err == nil && existingReceipt != nil {
		log.Printf("Duplicate receipt detected (by hash): %s", imageHash)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "duplicate",
			"message": "This receipt has already been uploaded",
			"receipt": existingReceipt,
		})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)

	// Determine receipt year and construct path in unused directory
	receiptYear := time.Now().Year()
	savePath := getReceiptFilePath(s.ReceiptDir, receiptYear, filename, false)

	// Ensure directory exists
	if err := ensureDirectoryExists(savePath); err != nil {
		log.Printf("Failed to create directory: %v", err)
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	outFile, err := os.Create(savePath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	if _, err := outFile.Write(fileData); err != nil {
		log.Printf("Failed to write file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("File saved to: %s", savePath)

	ocrResult, err := callOCRService(savePath, s.OCRServiceURL)
	if err != nil {
		log.Printf("OCR processing failed: %v", err)
		http.Error(w, fmt.Sprintf("OCR processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("OCR Result: Vendor=%s, Amount=%.2f, Date=%s, HSAStatus=%s",
		ocrResult.Vendor, ocrResult.Amount, ocrResult.Date, ocrResult.HSAStatus)

	var receiptDate time.Time
	if ocrResult.Date != "" {
		receiptDate, err = time.Parse("01/02/2006", ocrResult.Date)
		if err != nil {
			log.Printf("Failed to parse date '%s': %v", ocrResult.Date, err)
			receiptDate = time.Now()
		}
	} else {
		receiptDate = time.Now()
	}

	// Normalize HSA status
	hsaStatus := ocrResult.HSAStatus
	if hsaStatus == "" {
		if ocrResult.HSAQualified {
			hsaStatus = internal.HSAStatusYes
		} else {
			hsaStatus = internal.HSAStatusNo
		}
	}

	// Check for duplicates by vendor/amount/date
	if existingReceipt, err := s.DB.GetDuplicateReceipt(ocrResult.Vendor, ocrResult.Amount, receiptDate); err == nil && existingReceipt != nil {
		log.Printf("Duplicate receipt detected (by data): vendor=%s, amount=%.2f, date=%s",
			ocrResult.Vendor, ocrResult.Amount, receiptDate.Format("2006-01-02"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "duplicate",
			"message": "A similar receipt already exists (same vendor, amount, and date)",
			"receipt": existingReceipt,
		})
		return
	}

	receipt := &internal.Receipt{
		UserID:       internal.HOUSEHOLD_USER,
		Vendor:       ocrResult.Vendor,
		TotalAmount:  ocrResult.Amount,
		Date:         receiptDate,
		HSAQualified: hsaStatus == internal.HSAStatusYes || hsaStatus == internal.HSAStatusPartially,
		HSAStatus:    hsaStatus,
		ImagePath:    savePath,
		ImageHash:    imageHash,
		RawText:      ocrResult.RawText,
		Used:         false,
	}

	err = s.DB.CreateReceipt(receipt)
	if err != nil {
		log.Printf("Failed to save receipt to database: %v", err)
		http.Error(w, "Failed to save receipt", http.StatusInternalServerError)
		return
	}

	log.Printf("Receipt saved to database with ID: %d", receipt.ID)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"id":            receipt.ID,
		"vendor":        receipt.Vendor,
		"amount":        receipt.TotalAmount,
		"date":          receipt.Date.Format("2006-01-02"),
		"hsa_qualified": receipt.HSAQualified,
		"hsa_status":    receipt.HSAStatus,
		"message":       "Receipt uploaded successfully",
	}
	json.NewEncoder(w).Encode(response)
}

func callOCRService(imagePath string, ocrServiceURL string) (*OCRResponse, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	writer.Close()

	url := fmt.Sprintf("%s/parse", ocrServiceURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OCR service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OCR service returned status %d: %s",
			resp.StatusCode, string(bodyBytes))
	}

	var ocrResult OCRResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResult); err != nil {
		return nil, fmt.Errorf("failed to parse OCR response: %v", err)
	}

	return &ocrResult, nil
}

func ReceiptsHandler(w http.ResponseWriter, r *http.Request, s *internal.Server) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	receipts, err := s.DB.GetAllReceipts(internal.HOUSEHOLD_USER)
	if err != nil {
		log.Printf("Failed to get receipts: %v", err)
		http.Error(w, "Failed to retrieve receipts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipts)
}

func ReceiptByIDHandler(w http.ResponseWriter, r *http.Request, s *internal.Server) {
	path := strings.TrimPrefix(r.URL.Path, "/api/receipts/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid receipt ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		GetReceiptHandler(w, r, s, id)
	case http.MethodPut:
		UpdateReceiptHandler(w, r, s, id)
	case http.MethodDelete:
		DeleteReceiptHandler(w, r, s, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func GetReceiptHandler(w http.ResponseWriter, r *http.Request, s *internal.Server, id int) {
	receipt, err := s.DB.GetReceiptByID(id)
	if err != nil {
		log.Printf("Failed to get receipt: %v", err)
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

func UpdateReceiptHandler(w http.ResponseWriter, r *http.Request, s *internal.Server, id int) {
	receipt, err := s.DB.GetReceiptByID(id)
	if err != nil {
		log.Printf("Failed to get receipt: %v", err)
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Track if we're changing the used status
	wasUsed := receipt.Used
	willBeUsed := receipt.Used

	log.Printf("UpdateReceipt ID=%d: Initial state - wasUsed=%v, ImagePath=%s", id, wasUsed, receipt.ImagePath)

	if vendor, ok := updates["vendor"].(string); ok {
		receipt.Vendor = vendor
	}
	if amount, ok := updates["total_amount"].(float64); ok {
		receipt.TotalAmount = amount
	}
	if dateStr, ok := updates["date"].(string); ok {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			receipt.Date = date
		}
	}
	if hsaQual, ok := updates["hsa_qualified"].(bool); ok {
		receipt.HSAQualified = hsaQual
	}
	if hsaStatus, ok := updates["hsa_status"].(string); ok {
		receipt.HSAStatus = hsaStatus
		receipt.HSAQualified = hsaStatus == internal.HSAStatusYes || hsaStatus == internal.HSAStatusPartially
	}
	if used, ok := updates["used"].(bool); ok {
		log.Printf("UpdateReceipt ID=%d: Updating 'used' from %v to %v", id, receipt.Used, used)
		receipt.Used = used
		willBeUsed = used
		if used {
			now := time.Now()
			receipt.UsedDate = &now
		} else {
			receipt.UsedDate = nil
		}
	}
	if useReason, ok := updates["use_reason"].(string); ok {
		receipt.UseReason = &useReason
	}

	// Move file if changing used status (either direction)
	log.Printf("UpdateReceipt ID=%d: Checking file move - wasUsed=%v, willBeUsed=%v, ImagePath=%s", 
		id, wasUsed, willBeUsed, receipt.ImagePath)
	
	if wasUsed != willBeUsed && receipt.ImagePath != "" {
		log.Printf("UpdateReceipt ID=%d: File move condition met!", id)
		// Determine the year for the file path
		fileYear := time.Now().Year()
		if willBeUsed && receipt.UsedDate != nil {
			fileYear = receipt.UsedDate.Year()
		} else if !willBeUsed {
			// When moving back to unused, use the receipt date year
			fileYear = receipt.Date.Year()
		}

		filename := filepath.Base(receipt.ImagePath)
		newPath, err := moveReceiptFile(receipt.ImagePath, s.ReceiptDir, fileYear, filename, willBeUsed)
		if err != nil {
			log.Printf("Warning: Failed to move receipt file: %v", err)
			// Don't fail the request, just log the warning
		} else {
			receipt.ImagePath = newPath
			if willBeUsed {
				log.Printf("Moved receipt file to used: %s", newPath)
			} else {
				log.Printf("Moved receipt file back to unused: %s", newPath)
			}
		}
	}

	if err := s.DB.UpdateReceipt(receipt); err != nil {
		log.Printf("Failed to update receipt: %v", err)
		http.Error(w, "Failed to update receipt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(receipt)
}

func DeleteReceiptHandler(w http.ResponseWriter, r *http.Request, s *internal.Server, id int) {
	receipt, err := s.DB.GetReceiptByID(id)
	if err != nil {
		log.Printf("Failed to get receipt: %v", err)
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	if err := s.DB.DeleteReceipt(id); err != nil {
		log.Printf("Failed to delete receipt: %v", err)
		http.Error(w, "Failed to delete receipt", http.StatusInternalServerError)
		return
	}

	if receipt.ImagePath != "" {
		if err := os.Remove(receipt.ImagePath); err != nil {
			log.Printf("Warning: Failed to delete file %s: %v", receipt.ImagePath, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Receipt deleted successfully"})
}

func ServeReceiptFile(w http.ResponseWriter, r *http.Request, s *internal.Server) {
	path := strings.TrimPrefix(r.URL.Path, "/receipts/file/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid receipt ID", http.StatusBadRequest)
		return
	}

	receipt, err := s.DB.GetReceiptByID(id)
	if err != nil {
		log.Printf("Failed to get receipt: %v", err)
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	if _, err := os.Stat(receipt.ImagePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(receipt.ImagePath))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	case ".heic":
		contentType = "image/heic"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, receipt.ImagePath)
}