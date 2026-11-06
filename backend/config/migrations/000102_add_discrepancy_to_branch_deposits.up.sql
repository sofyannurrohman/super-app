ALTER TABLE branch_deposits
ADD COLUMN system_amount DECIMAL(20,2) DEFAULT 0,
ADD COLUMN discrepancy_notes TEXT;
