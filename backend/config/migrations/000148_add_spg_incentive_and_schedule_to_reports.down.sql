ALTER TABLE mt_spg_daily_reports
DROP COLUMN IF EXISTS schedule_id,
DROP COLUMN IF EXISTS target_amount,
DROP COLUMN IF EXISTS base_fee,
DROP COLUMN IF EXISTS incentive_bonus,
DROP COLUMN IF EXISTS total_spg_payout,
DROP COLUMN IF EXISTS is_target_achieved;
