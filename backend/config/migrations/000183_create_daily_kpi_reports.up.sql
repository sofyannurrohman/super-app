CREATE TABLE IF NOT EXISTS daily_kpi_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_date DATE NOT NULL,
    reporter_employee_id UUID NOT NULL REFERENCES employees(id),
    reporter_role VARCHAR(50) NOT NULL DEFAULT 'SPV',
    branch_id UUID REFERENCES branches(id),
    area_name VARCHAR(150),
    parent_report_id UUID,
    status VARCHAR(30) NOT NULL DEFAULT 'DRAFT',
    submitted_at TIMESTAMPTZ,
    reviewed_at TIMESTAMPTZ,
    reviewer_employee_id UUID REFERENCES employees(id),
    reviewer_feedback TEXT,
    summary_target_call INTEGER DEFAULT 0,
    summary_realization_call INTEGER DEFAULT 0,
    summary_achievement_call_percent DECIMAL(5,2) DEFAULT 0,
    summary_target_ec INTEGER DEFAULT 0,
    summary_realization_ec INTEGER DEFAULT 0,
    summary_achievement_ec_percent DECIMAL(5,2) DEFAULT 0,
    summary_target_sales DECIMAL(15,2) DEFAULT 0,
    summary_realization_sales DECIMAL(15,2) DEFAULT 0,
    summary_achievement_sales_percent DECIMAL(5,2) DEFAULT 0,
    tomorrow_priority TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_daily_kpi_reports_report_date ON daily_kpi_reports(report_date);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_reports_reporter_employee_id ON daily_kpi_reports(reporter_employee_id);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_reports_branch_id ON daily_kpi_reports(branch_id);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_reports_parent_report_id ON daily_kpi_reports(parent_report_id);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_reports_reviewer_employee_id ON daily_kpi_reports(reviewer_employee_id);

CREATE TABLE IF NOT EXISTS daily_kpi_report_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    daily_kpi_report_id UUID NOT NULL REFERENCES daily_kpi_reports(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    employee_name VARCHAR(150) NOT NULL,
    area_route VARCHAR(150),
    target_call INTEGER DEFAULT 0,
    realization_call INTEGER DEFAULT 0,
    achievement_call_percent DECIMAL(5,2) DEFAULT 0,
    target_ec INTEGER DEFAULT 0,
    realization_ec INTEGER DEFAULT 0,
    achievement_ec_percent DECIMAL(5,2) DEFAULT 0,
    target_sales DECIMAL(15,2) DEFAULT 0,
    realization_sales DECIMAL(15,2) DEFAULT 0,
    achievement_sales_percent DECIMAL(5,2) DEFAULT 0,
    diff_sales DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_daily_kpi_report_items_daily_kpi_report_id ON daily_kpi_report_items(daily_kpi_report_id);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_report_items_employee_id ON daily_kpi_report_items(employee_id);

CREATE TABLE IF NOT EXISTS daily_kpi_report_action_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    daily_kpi_report_id UUID NOT NULL REFERENCES daily_kpi_reports(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id),
    employee_name VARCHAR(150) NOT NULL,
    reason_unachieved TEXT,
    field_obstacle TEXT,
    action_follow_up TEXT,
    pic_name VARCHAR(100),
    due_date DATE,
    status VARCHAR(30) DEFAULT 'OPEN',
    support_needed TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_daily_kpi_report_action_plans_daily_kpi_report_id ON daily_kpi_report_action_plans(daily_kpi_report_id);
CREATE INDEX IF NOT EXISTS idx_daily_kpi_report_action_plans_employee_id ON daily_kpi_report_action_plans(employee_id);
