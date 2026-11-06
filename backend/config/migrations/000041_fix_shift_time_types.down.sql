-- Migration: Revert shift time column types from TIME to TIMESTAMPTZ
-- Version: 000041

ALTER TABLE shifts 
    ALTER COLUMN start_time TYPE TIMESTAMPTZ USING ('0001-01-01 ' || start_time::text)::TIMESTAMPTZ,
    ALTER COLUMN end_time TYPE TIMESTAMPTZ USING ('0001-01-01 ' || end_time::text)::TIMESTAMPTZ,
    ALTER COLUMN break_start TYPE TIMESTAMPTZ USING ('0001-01-01 ' || break_start::text)::TIMESTAMPTZ,
    ALTER COLUMN break_end TYPE TIMESTAMPTZ USING ('0001-01-01 ' || break_end::text)::TIMESTAMPTZ;
