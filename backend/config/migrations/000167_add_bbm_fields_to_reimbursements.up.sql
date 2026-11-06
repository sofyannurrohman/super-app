-- Add BBM and Odometer fields to reimbursements table
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS start_km DECIMAL(10,2);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS end_km DECIMAL(10,2);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS total_km DECIMAL(10,2);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS start_km_photo_url TEXT;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS end_km_photo_url TEXT;
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS vehicle_plate_number VARCHAR(50);
ALTER TABLE reimbursements ADD COLUMN IF NOT EXISTS fuel_liters DECIMAL(10,2);

CREATE INDEX IF NOT EXISTS idx_reimbursements_start_km ON reimbursements(start_km);
