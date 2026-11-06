-- Restore admin_designer role for users whose employee records are designated as designers/editors
UPDATE users
SET role = 'admin_designer'
WHERE role = 'employee'
  AND id IN (
      SELECT e.user_id
      FROM employees e
      LEFT JOIN job_positions jp ON jp.id = e.job_position_id
      WHERE LOWER(COALESCE(jp.title, '')) LIKE '%desain%'
         OR LOWER(COALESCE(jp.title, '')) LIKE '%editor%'
         OR e.id IN (SELECT DISTINCT editor_id FROM banner_orders WHERE editor_id IS NOT NULL)
  );
