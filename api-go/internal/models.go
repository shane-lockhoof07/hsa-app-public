package internal

import (
	"time"
)

// HSA qualification status constants
const (
	HSAStatusYes       = "Yes"
	HSAStatusNo        = "No"
	HSAStatusPartially = "Partially"
)

// Receipt represents a stored receipt with all metadata
type Receipt struct {
	ID           int        `json:"id"`
	UserID       string     `json:"user_id"`
	Vendor       string     `json:"vendor"`
	TotalAmount  float64    `json:"total_amount"` 
	Date         time.Time  `json:"date"`
	HSAQualified bool       `json:"hsa_qualified"`
	HSAStatus    string     `json:"hsa_status"`
	ImagePath    string     `json:"image_path"`
	ImageHash    string     `json:"image_hash"`
	RawText      string     `json:"raw_text"`
	Used         bool       `json:"used"`
	UsedDate     *time.Time `json:"used_date"`
	UseReason    *string    `json:"use_reason"`
	CreatedAt    time.Time  `json:"created_at"`
}

