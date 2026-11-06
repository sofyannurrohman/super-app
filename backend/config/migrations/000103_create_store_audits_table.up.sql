CREATE TYPE store_verification_status AS ENUM ('UNVERIFIED', 'VERIFIED_REAL', 'FICTITIOUS');

ALTER TABLE stores ADD COLUMN verification_status store_verification_status DEFAULT 'UNVERIFIED';

CREATE TABLE store_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id UUID NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    auditor_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    status store_verification_status NOT NULL,
    notes TEXT,
    photo_proof_url TEXT,
    audit_date TIMESTAMP DEFAULT now(),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
