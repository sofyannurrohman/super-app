DROP INDEX IF EXISTS idx_reimbursements_start_km;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS fuel_liters;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS vehicle_plate_number;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS end_km_photo_url;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS start_km_photo_url;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS total_km;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS end_km;
ALTER TABLE reimbursements DROP COLUMN IF EXISTS start_km;
