-- 1. Create kas_income_sources table
CREATE TABLE IF NOT EXISTS kas_income_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    account_code VARCHAR(50),
    account_name VARCHAR(200),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create kas_expense_types table
CREATE TABLE IF NOT EXISTS kas_expense_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Create kas_expense_accounts table
CREATE TABLE IF NOT EXISTS kas_expense_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    expense_type_id UUID NOT NULL REFERENCES kas_expense_types(id) ON DELETE CASCADE,
    account_code VARCHAR(50) NOT NULL UNIQUE,
    account_name VARCHAR(200) NOT NULL,
    accurate_account_id VARCHAR(100),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kas_exp_acc_type ON kas_expense_accounts(expense_type_id);
CREATE INDEX IF NOT EXISTS idx_kas_exp_acc_code ON kas_expense_accounts(account_code);

-- 4. Alter kas_transactions to add foreign keys
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS income_source_id UUID REFERENCES kas_income_sources(id) ON DELETE SET NULL;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS expense_type_id UUID REFERENCES kas_expense_types(id) ON DELETE SET NULL;
ALTER TABLE kas_transactions ADD COLUMN IF NOT EXISTS expense_account_id UUID REFERENCES kas_expense_accounts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_kas_tx_income_source ON kas_transactions(income_source_id);
CREATE INDEX IF NOT EXISTS idx_kas_tx_expense_type ON kas_transactions(expense_type_id);
CREATE INDEX IF NOT EXISTS idx_kas_tx_expense_account ON kas_transactions(expense_account_id);

-- 5. Seed initial income sources
INSERT INTO kas_income_sources (code, name, sort_order) VALUES
('PRODUCT_SALE', 'Penjualan Produk Pabrik (Faktur / Telemarketing)', 1),
('BRANCH_RECEIVABLE_PAYMENT', 'Pelunasan Talangan Piutang Cabang', 2),
('ADVANCE_REFUND', 'Pengembalian Sisa Kasbon / Uang Muka Dinas', 3),
('BRANCH_DEPOSIT', 'Setoran Kas Besar Cabang', 4),
('CLAIM_PENALTY', 'Pembayaran Claim / Denda Penalty', 5),
('ROSOK_SALE', 'Penjualan Rosok (Jurigen, Drum, dll)', 6),
('AMPAS_SALE', 'Penjualan / Pendapatan Ampas', 7),
('INTEREST_INCOME', 'Pendapatan Bunga Bank', 8),
('BANK_DISBURSEMENT', 'Pencairan Bank', 9),
('TRANSFER_FROM_KAS', 'Transfer dari KAS Lain', 10),
('OTHER', 'Lainnya', 99)
ON CONFLICT (code) DO NOTHING;

-- 6. Seed initial expense types
INSERT INTO kas_expense_types (id, code, name, description, sort_order) VALUES
('a0000001-0000-0000-0000-000000000001', 'OPERASIONAL_PABRIK', 'Biaya Operasional Pabrik & Umum', 'Biaya utilitas, operasional pabrik dan perlengkapan', 1),
('a0000001-0000-0000-0000-000000000002', 'ARMADA_LOGISTIK', 'Biaya Armada & Logistik', 'BBM, tol, pemeliharaan armada dan ekspedisi', 2),
('a0000001-0000-0000-0000-000000000003', 'TENAGA_KERJA_PRODUKSI', 'Biaya Tenaga Kerja & Produksi', 'Upah harian, borongan, lembur, dan teknisi', 3),
('a0000001-0000-0000-0000-000000000004', 'PERJALANAN_DINAS_TALANGAN', 'Perjalanan Dinas & Talangan', 'Kasbon dinas, perjalanan audit, dan piutang cabang', 4),
('a0000001-0000-0000-0000-000000000005', 'PEMASARAN_CSR', 'Biaya Pemasaran & CSR', 'Promosi, sample, banner MMT, dan CSR', 5),
('a0000001-0000-0000-0000-000000000006', 'PAJAK_LEGALITAS', 'Biaya Pajak & Legalitas', 'Pajak PPN, PPh, PBB, dan perijinan', 6),
('a0000001-0000-0000-0000-000000000007', 'PEMELIHARAAN_INVENTARIS', 'Pemeliharaan & Inventaris', 'Servis mesin, bangunan, inventaris', 7),
('a0000001-0000-0000-0000-000000000008', 'LAIN_LAIN_TRANSFER', 'Beban Lain & Transfer', 'Biaya administrasi bank, transfer kas, beban lain', 8)
ON CONFLICT (code) DO NOTHING;

-- 7. Seed initial expense accounts mapped to types
INSERT INTO kas_expense_accounts (expense_type_id, account_code, account_name, sort_order) VALUES
-- Operasional Pabrik
('a0000001-0000-0000-0000-000000000001', 'BIAYA_PERALATAN_KEBERSIHAN', 'Biaya Peralatan & Perlengkapan Kebersihan', 1),
('a0000001-0000-0000-0000-000000000001', 'PERLENGKAPAN_KANTOR', 'Biaya Perlengkapan Kantor', 2),
('a0000001-0000-0000-0000-000000000001', 'PERLENGKAPAN_PABRIK', 'Biaya Perlengkapan dan Peralatan Pabrik', 3),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_SOLAR', 'Biaya Solar', 4),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_GAS_LPG', 'Biaya Gas LPG', 5),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_PDAM_PABRIK', 'Biaya PDAM Pabrik', 6),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_PLN_PABRIK', 'Biaya PLN Pabrik', 7),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_PLN_MESS', 'Biaya PLN Mess', 8),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_TELPON_PABRIK', 'Biaya Telpon Pabrik', 9),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_PULSA_HP', 'Biaya Pulsa HP', 10),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_INTERNET', 'Biaya Internet', 11),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_FOTOCOPY', 'Biaya Fotocopy', 12),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_MATERAI', 'Biaya Materai', 13),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_KONSUMSI', 'Biaya Konsumsi', 14),
('a0000001-0000-0000-0000-000000000001', 'ENTERTAINT', 'Biaya Entertaint', 15),
('a0000001-0000-0000-0000-000000000001', 'BIAYA_SEWA_TANAH_LIMBAH_SAMBIREJO', 'Biaya Sewa Tanah Limbah Sambirejo', 16),

-- Armada & Logistik
('a0000001-0000-0000-0000-000000000002', 'BIAYA_BBM_KENDARAAN', 'BBM & Tol Armada Kendaraan', 1),
('a0000001-0000-0000-0000-000000000002', 'BIAYA_PEMELIHARAAN_KENDARAAN', 'Pemeliharaan & Servis Armada Kendaraan', 2),
('a0000001-0000-0000-0000-000000000002', 'EXPEDISI', 'Biaya Ekspedisi (Raja Transport)', 3),
('a0000001-0000-0000-0000-000000000002', 'EKSPEDISI_OTHER', 'Ekspedisi JNT/POS/JNE, dsb', 4),
('a0000001-0000-0000-0000-000000000002', 'BIAYA_RETRIBUSI_PARKIR', 'Biaya Retribusi Parkir', 5),

-- Tenaga Kerja & Produksi
('a0000001-0000-0000-0000-000000000003', 'DAILY_WORKER', 'Gaji Harian Mingguan', 1),
('a0000001-0000-0000-0000-000000000003', 'KULI', 'Gaji Kuli Muat / Bongkar', 2),
('a0000001-0000-0000-0000-000000000003', 'GAJI_TUKANG_PABRIK', 'Biaya Gaji Tukang Pabrik', 3),
('a0000001-0000-0000-0000-000000000003', 'GAJI_BORONGAN_PACKING_KECAP', 'Biaya Gaji Borongan Packing Kecap', 4),
('a0000001-0000-0000-0000-000000000003', 'GAJI_BORONGAN_PACKING_SAOS', 'Biaya Gaji Borongan Packing Saos', 5),
('a0000001-0000-0000-0000-000000000003', 'GAJI_BORONGAN_OVEN', 'Biaya Gaji Borongan Oven', 6),
('a0000001-0000-0000-0000-000000000003', 'GAJI_PRODUKSI_KECAP', 'Biaya Gaji Produksi Kecap', 7),
('a0000001-0000-0000-0000-000000000003', 'GAJI_PRODUKSI_SAOS', 'Biaya Gaji Produksi Saos', 8),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BORONGAN_MUAT_KECAP', 'Biaya Borongan Muat Kecap', 9),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BORONGAN_MUAT_SAOS', 'Biaya Borongan Muat Saos', 10),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BORONGAN_MUAT_BIHUN', 'Biaya Borongan Muat Bihun', 11),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BORONGAN_MUAT_CUKA', 'Biaya Borongan Muat Cuka', 12),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BORONGAN_TRONTON', 'Biaya Borongan dan Muat Tronton', 13),
('a0000001-0000-0000-0000-000000000003', 'PEMBERSIHAN_PRODUKSI', 'Biaya Pembersihan Tempat Produksi', 14),
('a0000001-0000-0000-0000-000000000003', 'DROP_GULA', 'Biaya Drop Gula', 15),
('a0000001-0000-0000-0000-000000000003', 'UANG_MAKAN_TEKNISI', 'Biaya Uang Makan Teknisi', 16),
('a0000001-0000-0000-0000-000000000003', 'LEMBUR', 'Biaya Lembur', 17),
('a0000001-0000-0000-0000-000000000003', 'LEMBUR_TEKNISI', 'Biaya Lembur Teknisi', 18),
('a0000001-0000-0000-0000-000000000003', 'PERLENGKAPAN_TEKNISI', 'Biaya Perlengkapan dan Peralatan Teknisi', 19),
('a0000001-0000-0000-0000-000000000003', 'CUCI_BOTOL_LINGKUNGAN', 'Biaya Cuci Botol Lingkungan', 20),
('a0000001-0000-0000-0000-000000000003', 'CUCI_BOTOL_GUNUNGKEBO', 'Biaya Cuci Botol Gunungkebo', 21),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_KEHADIRAN_KARYAWAN', 'Biaya Kehadiran Karyawan', 22),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_PERAWATAN_MEDIS', 'Biaya Perawatan Medis', 23),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_KOMISI_BONUS', 'Biaya Komisi & Bonus Karyawan', 24),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_THR', 'Biaya THR', 25),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BPJS_KETENAGAKERJAAN', 'Biaya BPJS Ketenagakerjaan', 26),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_BPJS_KESEHATAN', 'Biaya BPJS Kesehatan', 27),
('a0000001-0000-0000-0000-000000000003', 'BIAYA_ASURANSI', 'Biaya Asuransi', 28),
('a0000001-0000-0000-0000-000000000003', 'BEBAN_LAIN_PRODUKSI', 'Beban Lain-Lain Produksi', 29),

-- Perjalanan Dinas & Talangan
('a0000001-0000-0000-0000-000000000004', 'PERJALANAN_DINAS', 'Biaya Perjalanan Dinas Umum', 1),
('a0000001-0000-0000-0000-000000000004', 'BIAYA_BBM_PERJALANAN_DINAS', 'Biaya BBM Perjalanan Dinas', 2),
('a0000001-0000-0000-0000-000000000004', 'BIAYA_PERJALANAN_DINAS_AUDIT', 'Biaya Perjalanan Dinas Audit', 3),
('a0000001-0000-0000-0000-000000000004', 'KASBON_DINAS', 'Kasbon / Uang Muka Perjalanan Dinas (Piutang Karyawan)', 4),
('a0000001-0000-0000-0000-000000000004', 'BIAYA_OPERASIONAL_CABANG', 'Talangan Biaya Operasional Cabang (Piutang Cabang)', 5),
('a0000001-0000-0000-0000-000000000004', 'BIAYA_REIMBURSE', 'Biaya Reimburse', 6),

-- Pemasaran & CSR
('a0000001-0000-0000-0000-000000000005', 'PROMOSI', 'Biaya Promosi', 1),
('a0000001-0000-0000-0000-000000000005', 'BIAYA_SAMPLE', 'Biaya Sample', 2),
('a0000001-0000-0000-0000-000000000005', 'BIAYA_BANNER_MMT', 'Biaya Banner MMT', 3),
('a0000001-0000-0000-0000-000000000005', 'CSR', 'Sumbangan / CSR', 4),

-- Pajak & Legalitas
('a0000001-0000-0000-0000-000000000006', 'PAJAK_PPN', 'Biaya Pajak PPN', 1),
('a0000001-0000-0000-0000-000000000006', 'PAJAK_PPH', 'Biaya Pajak PPH', 2),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPH_FINAL', 'Biaya PPh Final', 3),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPH_21', 'Biaya PPh Pasal 21', 4),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPH_22', 'Biaya PPh Pasal 22', 5),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPH_23', 'Biaya PPh Pasal 23', 6),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPH_25', 'Biaya PPh Pasal 25', 7),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PPN_KELUARAN', 'Biaya PPN Keluaran', 8),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PBB', 'Biaya PBB', 9),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PAJAK_AIR_TANAH', 'Biaya Pajak Air Tanah', 10),
('a0000001-0000-0000-0000-000000000006', 'BIAYA_PERIJINAN', 'Biaya Perijinan', 11),

-- Pemeliharaan & Inventaris
('a0000001-0000-0000-0000-000000000007', 'INVENTARIS', 'Inventaris', 1),
('a0000001-0000-0000-0000-000000000007', 'BIAYA_PEMELIHARAAN_INVENTARIS', 'Biaya Pemeliharaan Inventaris', 2),
('a0000001-0000-0000-0000-000000000007', 'BIAYA_PEMELIHARAAN_MESIN', 'Biaya Pemeliharaan Mesin', 3),
('a0000001-0000-0000-0000-000000000007', 'BIAYA_PEMELIHARAAN_BANGUNAN_PABRIK', 'Biaya Pemeliharaan Bangunan Pabrik', 4),

-- Lain-lain & Transfer
('a0000001-0000-0000-0000-000000000008', 'SUPPLIER_PAYMENT', 'Pembayaran Supplier', 1),
('a0000001-0000-0000-0000-000000000008', 'ADMIN_BANK', 'Biaya Admin Bank', 2),
('a0000001-0000-0000-0000-000000000008', 'BIAYA_BUNGA_BANK', 'Biaya Bunga Bank', 3),
('a0000001-0000-0000-0000-000000000008', 'TRANSFER_TO_KAS', 'Transfer ke KAS Lain', 4),
('a0000001-0000-0000-0000-000000000008', 'BEBAN_LAIN', 'Beban Lain-Lain', 5),
('a0000001-0000-0000-0000-000000000008', 'OTHER', 'Lainnya', 99)
ON CONFLICT (account_code) DO NOTHING;
