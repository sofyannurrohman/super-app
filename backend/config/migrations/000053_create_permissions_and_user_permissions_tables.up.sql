CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL, -- e.g., 'employee:read', 'payroll:write'
    menu_group VARCHAR(100) NOT NULL, -- e.g., 'Employee', 'Payroll'
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    is_granted BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, permission_id)
);

-- Seed Default Master Permissions
INSERT INTO permissions (name, menu_group, description) VALUES
('employee:read', 'Employee', 'Membaca data karyawan'),
('employee:write', 'Employee', 'Membuat, mengubah, dan menghapus data karyawan'),
('payroll:read', 'Payroll', 'Melihat data slip gaji dan laporan penggajian'),
('payroll:write', 'Payroll', 'Membuat, mengubah, dan menghapus proses penggajian'),
('attendance:read', 'Attendance', 'Melihat rekap absensi karyawan'),
('attendance:write', 'Attendance', 'Mengubah atau memverifikasi absensi manual'),
('leave:read', 'Leave', 'Melihat daftar pengajuan cuti/izin'),
('leave:write', 'Leave', 'Menyetujui atau menolak pengajuan cuti/izin'),
('overtime:read', 'Overtime', 'Melihat daftar pengajuan lembur'),
('overtime:write', 'Overtime', 'Menyetujui atau menolak pengajuan lembur'),
('reimbursement:read', 'Reimbursement', 'Melihat daftar pengajuan reimbursement'),
('reimbursement:write', 'Reimbursement', 'Menyetujui atau memproses pembayaran reimbursement'),
('sales:read', 'Sales', 'Melihat data transaksi sales dan toko'),
('sales:write', 'Sales', 'Membuat order penjualan atau manipulasi data sales'),
('delivery:read', 'Delivery', 'Melihat dashboard tugas pengiriman surat jalan'),
('delivery:write', 'Delivery', 'Membuat, menugaskan, dan memperbarui surat jalan pengiriman'),
('warehouse:read', 'Warehouse', 'Melihat stok gudang dan mutasi produk'),
('warehouse:write', 'Warehouse', 'Mengupdate stok gudang dan memproses penerimaan barang'),
('announcement:read', 'Announcement', 'Melihat pengumuman perusahaan'),
('announcement:write', 'Announcement', 'Membuat, mengubah, dan mempublikasikan pengumuman');
