-- Bersihkan flag palsu yang terpasang akibat bug self-match pada kunjungan sales tanggal 15 Okt 2026 dan seterusnya
UPDATE sales_visits
SET notes = TRIM(REPLACE(notes, '[SISTEM: Kunjungan berulang < 2 hari tanpa order sebelumnya]', ''))
WHERE notes LIKE '%[SISTEM: Kunjungan berulang < 2 hari tanpa order sebelumnya]%'
  AND check_in_time >= '2026-10-15 00:00:00';
