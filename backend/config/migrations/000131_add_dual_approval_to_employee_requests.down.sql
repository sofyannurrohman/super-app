ALTER TABLE employee_requests
DROP COLUMN IF EXISTS requester_job_title,
DROP COLUMN IF EXISTS hr_approver_id,
DROP COLUMN IF EXISTS hr_approval_date,
DROP COLUMN IF EXISTS hr_approval_notes,
DROP COLUMN IF EXISTS hr_status,
DROP COLUMN IF EXISTS director_approver_id,
DROP COLUMN IF EXISTS director_approval_date,
DROP COLUMN IF EXISTS director_approval_notes,
DROP COLUMN IF EXISTS director_status,
DROP COLUMN IF EXISTS rejected_by_role;
