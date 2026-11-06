ALTER TABLE sales_kpis 
ADD COLUMN achieved_omzet_lama DECIMAL(15, 2) DEFAULT 0,
ADD COLUMN achieved_omzet_baru DECIMAL(15, 2) DEFAULT 0,
ADD COLUMN achieved_new_stores INT DEFAULT 0;
