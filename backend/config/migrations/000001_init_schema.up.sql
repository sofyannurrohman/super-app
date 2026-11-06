CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Master Perusahaan (Multi-tenant support)
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    tax_number VARCHAR(50), -- NPWP Perusahaan
    address TEXT,
    logo_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Cabang Kantor
CREATE TABLE branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- e.g., "Head Office", "Cabang Bandung"
    address TEXT,
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta', -- Penting untuk absensi
    latitude DECIMAL(10, 8), -- Pusat lokasi kantor untuk Geofencing
    longitude DECIMAL(11, 8),
    radius_meter INT DEFAULT 100, -- Batas toleransi jarak absen
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Departemen
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES departments(id) ON DELETE SET NULL, -- Hierarchy (Sub-department)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Level Jabatan
CREATE TABLE job_positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL, -- e.g., "Senior Software Engineer"
    level INT DEFAULT 1, -- e.g., 1=Staff, 5=Manager, 9=C-Level
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users (Login Credential)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'employee', -- 'superadmin', 'hr_admin', 'employee'
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    job_position_id UUID REFERENCES job_positions(id) ON DELETE SET NULL,
    
    -- Personal Data
    employee_id_number VARCHAR(50) NOT NULL, -- NIK Karyawan (Nomor Induk)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    gender VARCHAR(10), -- 'MALE', 'FEMALE'
    birth_date DATE,
    birth_place VARCHAR(100),
    marital_status VARCHAR(20), -- 'SINGLE', 'MARRIED', 'WIDOW'
    address_ktp TEXT,
    address_residential TEXT,
    face_reference_url TEXT,
    face_embedding JSONB,
    is_face_registered BOOLEAN,
    
    -- Government IDs (Indonesian Specific)
    identity_number VARCHAR(50), -- No KTP
    npwp_number VARCHAR(50), -- NPWP
    bpjs_ketenagakerjaan_number VARCHAR(50),
    bpjs_kesehatan_number VARCHAR(50),
    ptkp_status VARCHAR(10), -- K/0, TK/0, K/1, dll (Penting untuk PPh21)
    
    -- Employment Status
    join_date DATE NOT NULL,
    end_date DATE, -- Nullable if active
    employment_status VARCHAR(50), -- 'PERMANENT', 'CONTRACT', 'PROBATION'
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL, -- Self Join untuk atasan langsung
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Definisi Shift (Jadwal Kerja)
CREATE TABLE shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id),
    name VARCHAR(50) NOT NULL, -- e.g., "Office Hour", "Shift Malam"
    start_time TIME NOT NULL, -- 09:00
    end_time TIME NOT NULL, -- 18:00
    break_start TIME,
    break_end TIME,
    is_flexible BOOLEAN DEFAULT FALSE, -- Jika true, jam kerja bebas asal total jam terpenuhi
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Jadwal Karyawan (Roster)
CREATE TABLE employee_shifts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    shift_id UUID REFERENCES shifts(id),
    date DATE NOT NULL, -- Tanggal shift berlaku
    is_off_day BOOLEAN DEFAULT FALSE, -- Libur/Off
    UNIQUE(employee_id, date)
);

-- Log Absensi (Live Attendance)
CREATE TABLE attendance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    shift_id UUID REFERENCES shifts(id), -- Shift yang sedang dijalani
    
    clock_in_time TIMESTAMPTZ,
    clock_out_time TIMESTAMPTZ,
    
    -- Validasi Lokasi Clock In
    clock_in_lat DECIMAL(10, 8),
    clock_in_long DECIMAL(11, 8),
    clock_in_photo_url TEXT, -- Bukti foto selfie
    clock_in_device_id VARCHAR(100),
    clock_in_location_status VARCHAR(20), -- 'INSIDE_RADIUS', 'OUTSIDE_RADIUS'
    
    -- Validasi Lokasi Clock Out
    clock_out_lat DECIMAL(10, 8),
    clock_out_long DECIMAL(11, 8),
    clock_out_photo_url TEXT,
    
    status VARCHAR(20), -- 'PRESENT', 'LATE', 'EARLY_LEAVE', 'ABSENT'
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Jenis Cuti
CREATE TABLE leave_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id),
    name VARCHAR(50) NOT NULL, -- e.g., "Cuti Tahunan", "Sakit", "Unpaid Leave"
    is_paid BOOLEAN DEFAULT TRUE,
    default_quota INT DEFAULT 12, -- Kuota default per tahun
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Saldo Cuti Karyawan
CREATE TABLE leave_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID REFERENCES leave_types(id),
    year INT NOT NULL, -- e.g., 2024
    balance_total INT DEFAULT 0,
    balance_used INT DEFAULT 0,
    UNIQUE(employee_id, leave_type_id, year)
);

-- Request Cuti
CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    leave_type_id UUID REFERENCES leave_types(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    attachment_url TEXT, -- Surat dokter, dll
    status VARCHAR(20) DEFAULT 'PENDING', -- 'APPROVED', 'REJECTED'
    approved_by UUID REFERENCES employees(id) ON DELETE SET NULL, -- Manager who approved
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Komponen Gaji (Master)
CREATE TABLE payroll_components (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id),
    name VARCHAR(100) NOT NULL, -- e.g., "Gaji Pokok", "Tunjangan Makan", "Potongan Terlambat"
    type VARCHAR(20) NOT NULL, -- 'EARNING', 'DEDUCTION', 'BENEFIT' (BPJS perusahaan)
    is_taxable BOOLEAN DEFAULT TRUE, -- Apakah kena PPh21?
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Setting Gaji Karyawan (Berapa gaji si A saat ini?)
CREATE TABLE employee_salary_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    component_id UUID REFERENCES payroll_components(id),
    amount DECIMAL(15, 2) NOT NULL, -- Nominal
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payroll Run (Batch Penggajian Bulanan)
CREATE TABLE payroll_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id),
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    payment_schedule DATE, -- Tanggal gajian
    status VARCHAR(20) DEFAULT 'DRAFT', -- 'PROCESSED', 'COMPLETED'
    total_payout DECIMAL(20, 2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payslip Header (Slip Gaji per Karyawan)
CREATE TABLE payslips (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payroll_run_id UUID REFERENCES payroll_runs(id),
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,
    
    -- Snapshot data penting saat gaji dibuat (jika karyawan pindah jabatan bulan depan, history tetap aman)
    snapshot_job_title VARCHAR(100),
    snapshot_ptkp_status VARCHAR(10),
    
    basic_salary DECIMAL(15, 2),
    total_allowance DECIMAL(15, 2),
    total_deduction DECIMAL(15, 2),
    take_home_pay DECIMAL(15, 2),
    
    pph21_amount DECIMAL(15, 2), -- Pajak
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payslip Detail (Rincian komponen di slip gaji)
CREATE TABLE payslip_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payslip_id UUID REFERENCES payslips(id),
    component_name VARCHAR(100) NOT NULL, -- Disimpan text-nya agar immutable
    amount DECIMAL(15, 2) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'EARNING', 'DEDUCTION'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Performance & KPI
CREATE TABLE sales_kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    target_omzet DECIMAL(15, 2) NOT NULL,
    achieved_omzet DECIMAL(15, 2) DEFAULT 0,
    estimated_bonus DECIMAL(15, 2) DEFAULT 0,
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE employee_kpis (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    attendance_score DECIMAL(5, 2) DEFAULT 0,
    productivity_score DECIMAL(5, 2) DEFAULT 0,
    final_score DECIMAL(5, 2) DEFAULT 0,
    period_month INT NOT NULL,
    period_year INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE managerial_appraisals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    manager_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    review_notes TEXT,
    rating DECIMAL(3, 2) NOT NULL,
    review_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);