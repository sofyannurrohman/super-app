-- Migration: Fix shift time column types from TIMESTAMPTZ to TIME
-- Version: 000041

ALTER TABLE shifts 
    ALTER COLUMN start_time TYPE TIME USING start_time::TIME,
    ALTER COLUMN end_time TYPE TIME USING end_time::TIME,
    ALTER COLUMN break_start TYPE TIME USING break_start::TIME,
    ALTER COLUMN break_end TYPE TIME USING break_end::TIME;
