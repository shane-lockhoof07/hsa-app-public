package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Database struct {
	conn *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Database{conn: db}, nil
}

func (db *Database) GetEligibleReceipts(userID string) ([]Receipt, error) {
	query := `
        SELECT id, user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
               image_path, image_hash, raw_text, used, used_date, use_reason, created_at
        FROM receipts
        WHERE user_id = $1 AND used = false AND (hsa_status = 'Yes' OR hsa_status = 'Partially')
        ORDER BY date DESC
    `

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []Receipt
	for rows.Next() {
		var r Receipt
		err := rows.Scan(&r.ID, &r.UserID, &r.Vendor, &r.TotalAmount,
			&r.Date, &r.HSAQualified, &r.HSAStatus, &r.ImagePath, &r.ImageHash, &r.RawText,
			&r.Used, &r.UsedDate, &r.UseReason, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, r)
	}

	return receipts, nil
}

func (db *Database) MarkUsed(receipts []Receipt) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE receipts SET used = true, used_date = NOW() WHERE id = $1")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range receipts {
		if _, err := stmt.Exec(r.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *Database) CreateReceipt(receipt *Receipt) error {
	query := `
        INSERT INTO receipts (user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
                             image_path, image_hash, raw_text, used, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
        RETURNING id, created_at
    `

	err := db.conn.QueryRow(
		query,
		receipt.UserID,
		receipt.Vendor,
		receipt.TotalAmount,
		receipt.Date,
		receipt.HSAQualified,
		receipt.HSAStatus,
		receipt.ImagePath,
		receipt.ImageHash,
		receipt.RawText,
		receipt.Used,
	).Scan(&receipt.ID, &receipt.CreatedAt)

	return err
}

func (db *Database) GetAllReceipts(userID string) ([]Receipt, error) {
	query := `
        SELECT id, user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
               image_path, image_hash, raw_text, used, used_date, use_reason, created_at
        FROM receipts
        WHERE user_id = $1
        ORDER BY date DESC
    `

	rows, err := db.conn.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []Receipt
	for rows.Next() {
		var r Receipt
		err := rows.Scan(&r.ID, &r.UserID, &r.Vendor, &r.TotalAmount,
			&r.Date, &r.HSAQualified, &r.HSAStatus, &r.ImagePath, &r.ImageHash, &r.RawText,
			&r.Used, &r.UsedDate, &r.UseReason, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, r)
	}

	return receipts, nil
}

func (db *Database) GetReceiptByID(id int) (*Receipt, error) {
	query := `
        SELECT id, user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
               image_path, image_hash, raw_text, used, used_date, use_reason, created_at
        FROM receipts
        WHERE id = $1
    `

	var r Receipt
	err := db.conn.QueryRow(query, id).Scan(
		&r.ID, &r.UserID, &r.Vendor, &r.TotalAmount,
		&r.Date, &r.HSAQualified, &r.HSAStatus, &r.ImagePath, &r.ImageHash, &r.RawText,
		&r.Used, &r.UsedDate, &r.UseReason, &r.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Add this logging
	log.Printf("GetReceiptByID(%d): Read from DB - used=%v, image_path=%s", id, r.Used, r.ImagePath)

	return &r, nil
}

func (db *Database) GetReceiptByImageHash(hash string) (*Receipt, error) {
	query := `
        SELECT id, user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
               image_path, image_hash, raw_text, used, used_date, use_reason, created_at
        FROM receipts
        WHERE image_hash = $1
    `

	var r Receipt
	err := db.conn.QueryRow(query, hash).Scan(
		&r.ID, &r.UserID, &r.Vendor, &r.TotalAmount,
		&r.Date, &r.HSAQualified, &r.HSAStatus, &r.ImagePath, &r.ImageHash, &r.RawText,
		&r.Used, &r.UsedDate, &r.UseReason, &r.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (db *Database) GetDuplicateReceipt(vendor string, amount float64, date time.Time) (*Receipt, error) {
	query := `
        SELECT id, user_id, vendor, total_amount, date, hsa_qualified, hsa_status,
               image_path, image_hash, raw_text, used, used_date, use_reason, created_at
        FROM receipts
        WHERE vendor = $1 
          AND ABS(total_amount - $2) < 0.01
          AND date = $3
        LIMIT 1
    `

	var r Receipt
	err := db.conn.QueryRow(query, vendor, amount, date).Scan(
		&r.ID, &r.UserID, &r.Vendor, &r.TotalAmount,
		&r.Date, &r.HSAQualified, &r.HSAStatus, &r.ImagePath, &r.ImageHash, &r.RawText,
		&r.Used, &r.UsedDate, &r.UseReason, &r.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (db *Database) UpdateReceipt(receipt *Receipt) error {
	query := `
        UPDATE receipts
        SET vendor = $1, total_amount = $2, date = $3, hsa_qualified = $4, hsa_status = $5,
            used = $6, used_date = $7, use_reason = $8, image_path = $9
        WHERE id = $10
    `

	_, err := db.conn.Exec(
		query,
		receipt.Vendor,
		receipt.TotalAmount,
		receipt.Date,
		receipt.HSAQualified,
		receipt.HSAStatus,
		receipt.Used,
		receipt.UsedDate,
		receipt.UseReason,
		receipt.ImagePath,
		receipt.ID,
	)

	return err
}

func (db *Database) DeleteReceipt(id int) error {
	query := "DELETE FROM receipts WHERE id = $1"
	_, err := db.conn.Exec(query, id)
	return err
}

// RunMigrations executes all migration files in order
func (db *Database) RunMigrations(migrationsDir string) error {
	log.Println("Running database migrations...")

	migrations := []string{
		"001_init.sql",
	}

	for _, migration := range migrations {
		migrationPath := fmt.Sprintf("%s/%s", migrationsDir, migration)
		log.Printf("Applying migration: %s", migration)

		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", migration, err)
		}

		if _, err := db.conn.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %v", migration, err)
		}

		log.Printf("Successfully applied migration: %s", migration)
	}

	log.Println("All migrations completed successfully")
	return nil
}