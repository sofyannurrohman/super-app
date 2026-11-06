ALTER TABLE epkwt_requests 
DROP COLUMN IF EXISTS contract_sequence,
DROP COLUMN IF EXISTS document_url,
DROP COLUMN IF EXISTS notes;
