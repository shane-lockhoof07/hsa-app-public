package internal

import (
    "encoding/json"
    "log"
    "net/http"
)

type Server struct {
    DB            *Database
    OCRServiceURL string
    ReceiptDir    string
}

const HOUSEHOLD_USER = "household"

func (s *Server) ListReceiptsHandler(w http.ResponseWriter, r *http.Request) {
    receipts, err := s.DB.GetAllReceipts(HOUSEHOLD_USER)
    if err != nil {
        log.Printf("Failed to get receipts: %v", err)
        http.Error(w, "Failed to get receipts", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(receipts)
}

func (s *Server) DeductHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID string  `json:"user_id"`
        Amount float64 `json:"amount"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if req.UserID == "" {
        req.UserID = HOUSEHOLD_USER
    }
    
    receipts, err := s.DB.GetEligibleReceipts(req.UserID)
    if err != nil {
        log.Printf("Failed to get eligible receipts: %v", err)
        http.Error(w, "Failed to get receipts", http.StatusInternalServerError)
        return
    }
    
    amounts := make([]float64, len(receipts))
    for i, r := range receipts {
        amounts[i] = r.TotalAmount
    }
    
    indices := SubsetSum(amounts, req.Amount)
    
    selected := make([]Receipt, len(indices))
    for i, idx := range indices {
        selected[i] = receipts[idx]
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(selected)
}