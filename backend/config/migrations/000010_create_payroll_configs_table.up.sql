CREATE TABLE payroll_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id UUID REFERENCES companies(id) UNIQUE,
    jht_company_percentage DECIMAL(5, 2) DEFAULT 3.70,
    jht_employee_percentage DECIMAL(5, 2) DEFAULT 2.00,
    jp_company_percentage DECIMAL(5, 2) DEFAULT 2.00,
    jp_employee_percentage DECIMAL(5, 2) DEFAULT 1.00,
    jp_max_wage_base DECIMAL(15, 2) DEFAULT 10042300,
    jkk_company_percentage DECIMAL(5, 2) DEFAULT 0.24,
    jkm_company_percentage DECIMAL(5, 2) DEFAULT 0.30,
    bpjs_kes_company_percentage DECIMAL(5, 2) DEFAULT 4.00,
    bpjs_kes_employee_percentage DECIMAL(5, 2) DEFAULT 1.00,
    bpjs_kes_max_wage_base DECIMAL(15, 2) DEFAULT 12000000,
    ptkp_base_tk0 DECIMAL(15, 2) DEFAULT 54000000,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
