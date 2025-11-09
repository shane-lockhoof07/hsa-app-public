-- Combined migration: Initial schema with all features
-- This replaces 001_init.sql, 002_add_image_hash.sql, and 003_add_hsa_status.sql

-- Create receipts table with all columns
CREATE TABLE IF NOT EXISTS receipts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    vendor VARCHAR(255),
    total_amount DECIMAL(10,2) NOT NULL,
    date DATE,
    hsa_qualified BOOLEAN DEFAULT true,
    hsa_status VARCHAR(20) DEFAULT 'Yes',  -- 'Yes', 'No', or 'Partially'
    image_path TEXT,
    image_hash VARCHAR(64),
    raw_text TEXT,
    used BOOLEAN DEFAULT false,
    used_date TIMESTAMP,
    use_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create all indexes (idempotent)
DO $$ 
BEGIN
    -- Index on user_id for filtering by user
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_receipts_user_id'
    ) THEN
        CREATE INDEX idx_receipts_user_id ON receipts(user_id);
    END IF;

    -- Index on used status for filtering unused receipts
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_receipts_used'
    ) THEN
        CREATE INDEX idx_receipts_used ON receipts(used);
    END IF;

    -- Index on image_hash for duplicate detection
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_receipts_image_hash'
    ) THEN
        CREATE INDEX idx_receipts_image_hash ON receipts(image_hash);
    END IF;

    -- Composite index for duplicate detection by vendor/amount/date
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_receipts_composite'
    ) THEN
        CREATE INDEX idx_receipts_composite ON receipts(vendor, total_amount, date);
    END IF;

    -- Index on HSA status for filtering qualified receipts
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_receipts_hsa_status'
    ) THEN
        CREATE INDEX idx_receipts_hsa_status ON receipts(hsa_status);
    END IF;
END $$;

-- Update any existing records that might have been created without hsa_status
-- (This handles migration from older schemas)
UPDATE receipts 
SET hsa_status = CASE 
    WHEN hsa_qualified = true THEN 'Yes'
    WHEN hsa_qualified = false THEN 'No'
    ELSE 'Yes'
END
WHERE hsa_status IS NULL;

-- Notes:
-- * hsa_qualified column is kept for backward compatibility
-- * total_amount represents only the HSA-qualified portion
-- * When hsa_status = 'Partially', total_amount contains just the qualified portion
-- * image_hash is used for duplicate detection via file content comparison