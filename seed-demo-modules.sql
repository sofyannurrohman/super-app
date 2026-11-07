-- Fictional, idempotent records for the isolated trial database only.
BEGIN;
DO $seed$
DECLARE
  c uuid;
  b uuid;
  e uuid;
  f uuid := md5('trial-demo-factory-1')::uuid;
  p1 uuid := md5('trial-demo-product-kecap')::uuid;
  p2 uuid := md5('trial-demo-product-saus')::uuid;
  p3 uuid := md5('trial-demo-product-tepung')::uuid;
  s uuid := md5('trial-demo-supplier')::uuid;
  t uuid := md5('trial-demo-store')::uuid;
  pr uuid := md5('trial-demo-purchase-request')::uuid;
BEGIN
  SELECT id INTO STRICT c FROM companies WHERE code = 'DEMO';
  SELECT id INTO STRICT b FROM branches WHERE company_id = c AND name = 'Kantor Pusat Demo';
  SELECT id INTO STRICT e FROM employees WHERE company_id = c AND employee_id_number = 'EMP-001';

  INSERT INTO factories (id, company_id, branch_id, name, location, description)
  VALUES (f, c, b, 'Pabrik Demo', 'Jakarta', 'Data fiktif untuk demonstrasi') ON CONFLICT (id) DO NOTHING;
  INSERT INTO products (id, company_id, name, sku, unit, pcs_per_unit, category, cost_price, factory_price, selling_price)
  VALUES (p1,c,'Kecap Manis Demo 250 ml','DEMO-KCP-250','PCS',1,'Kecap',9000,11000,15000),
         (p2,c,'Saus Sambal Demo 250 ml','DEMO-SSB-250','PCS',1,'Saus',8500,10000,14000),
         (p3,c,'Tepung Serbaguna Demo 1 kg','DEMO-TPG-1KG','PCS',1,'Tepung',7000,8500,12000)
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO factory_stocks (id, factory_id, product_id, batch_no, quantity)
  VALUES (md5('trial-demo-fstock-1')::uuid,f,p1,'DEMO-2026-01',600),
         (md5('trial-demo-fstock-2')::uuid,f,p2,'DEMO-2026-01',450),
         (md5('trial-demo-fstock-3')::uuid,f,p3,'DEMO-2026-01',300)
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO warehouse_stocks (id, branch_id, product_id, stock_type, batch_no, quantity)
  VALUES (md5('trial-demo-wstock-1')::uuid,b,p1,'GOOD','DEMO-2026-01',120),
         (md5('trial-demo-wstock-2')::uuid,b,p2,'GOOD','DEMO-2026-01',95),
         (md5('trial-demo-wstock-3')::uuid,b,p3,'GOOD','DEMO-2026-01',80)
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO suppliers (id, company_id, name, contact_name, city, is_active)
  VALUES (s,c,'PT Pemasok Demo','Rina Demo','Jakarta',true) ON CONFLICT (id) DO NOTHING;
  INSERT INTO raw_material_purchase_requests (id, company_id, factory_id, pr_number, request_date, status, notes, created_by_name)
  VALUES (pr,c,f,'PR-DEMO-001',CURRENT_DATE,'PENDING','Permintaan fiktif untuk demonstrasi','Admin Demo')
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO raw_material_purchase_request_items (id, purchase_request_id, product_id, item_name, requested_qty, unit)
  VALUES (md5('trial-demo-pr-item')::uuid,pr,p3,'Tepung Serbaguna Demo 1 kg',100,'KG')
  ON CONFLICT (id) DO NOTHING;

  INSERT INTO stores (id, company_id, name, owner_name, address, is_active)
  VALUES (t,c,'Toko Maju Demo','Sari Demo','Jakarta',true) ON CONFLICT (id) DO NOTHING;
  INSERT INTO sales_transactions (id, company_id, store_id, employee_id, total_amount, store_category, period_month, period_year, transaction_date, status)
  VALUES (md5('trial-demo-sale-1')::uuid,c,t,e,1500000,'TOKO_LAMA',EXTRACT(MONTH FROM CURRENT_DATE)::int,EXTRACT(YEAR FROM CURRENT_DATE)::int,CURRENT_DATE,'VERIFIED'),
         (md5('trial-demo-sale-2')::uuid,c,t,e,900000,'TOKO_LAMA',EXTRACT(MONTH FROM CURRENT_DATE)::int,EXTRACT(YEAR FROM CURRENT_DATE)::int,CURRENT_DATE - 1,'VERIFIED')
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO sales_kpis (id, employee_id, period_month, period_year, target_omzet, target_new_stores)
  VALUES (md5('trial-demo-kpi')::uuid,e,EXTRACT(MONTH FROM CURRENT_DATE)::int,EXTRACT(YEAR FROM CURRENT_DATE)::int,5000000,5)
  ON CONFLICT (id) DO NOTHING;

  INSERT INTO banner_orders (id, company_id, employee_id, store_id, campaign_name, store_name, location, banner_type, size, status, qty)
  VALUES (md5('trial-demo-banner')::uuid,c,e,t,'Promo Akhir Pekan Demo','Toko Maju Demo','Jakarta','Spanduk',2.5,'PENDING',2)
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO marketplace_products (branch_id, marketplace, product_id, product_name, sku_id, seller_sku, price_amount, price_currency, stock, status, category)
  SELECT b,'shopee',p1::text,'Kecap Manis Demo 250 ml','DEMO-SKU-KCP','DEMO-KCP-250',15000,'IDR',35,'ACTIVATE','Makanan'
  WHERE NOT EXISTS (SELECT 1 FROM marketplace_products WHERE marketplace='shopee' AND sku_id='DEMO-SKU-KCP' AND branch_id=b);
  INSERT INTO marketplace_stocks (id, branch_id, product_id, product_name, sku, quantity, min_stock, price, is_active, created_by)
  VALUES (md5('trial-demo-marketplace-stock')::uuid,b,p1,'Kecap Manis Demo 250 ml','DEMO-KCP-250',35,10,15000,true,'Admin Demo')
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO company_bank_accounts (id, company_id, branch_id, bank_name, account_number, account_holder, is_active, manual_balance, notes)
  VALUES (md5('trial-demo-bank-account')::uuid,c,b,'Bank Demo','0000000001','Perusahaan Demo',true,25000000,'Rekening fiktif; tidak dapat menerima pembayaran')
  ON CONFLICT (id) DO NOTHING;
  INSERT INTO vehicles (id, company_id, branch_id, name, license_plate, model, type, status, year, capacity)
  VALUES (md5('trial-demo-vehicle')::uuid,c,b,'Mobil Pengiriman Demo','B 0001 DEM','Van Demo','Van','AVAILABLE',2024,1000)
  ON CONFLICT (id) DO NOTHING;
END $seed$;
COMMIT;
