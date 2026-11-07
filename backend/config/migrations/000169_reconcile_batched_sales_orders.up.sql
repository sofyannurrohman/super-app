-- Migration: Reconcile sales orders that are already assigned to active delivery batches
-- Version: 000169

-- The batch association was previously added by GORM on long-lived installs.
ALTER TABLE sales_orders ADD COLUMN IF NOT EXISTS delivery_batch_id UUID REFERENCES delivery_batches(id) ON DELETE SET NULL;

-- 1. Update sales orders that are in COMPLETED delivery batches
UPDATE sales_orders so
SET delivery_batch_id = di.delivery_batch_id,
    delivery_order_no = db.delivery_order_no,
    status = CASE 
        WHEN so.status IN ('CONFIRMED', 'WAITING_STOCK', 'WAITING_WAREHOUSE', 'DRAFT') THEN 'DELIVERED'
        ELSE so.status 
    END
FROM delivery_items di
JOIN delivery_batches db ON db.id = di.delivery_batch_id
WHERE di.sales_order_id = so.id
  AND db.status = 'COMPLETED'
  AND (so.delivery_batch_id IS NULL OR so.status IN ('CONFIRMED', 'WAITING_STOCK'));

-- 2. Update sales orders that are in ongoing/active delivery batches (WAITING_APPROVAL, SUPERVISOR_APPROVED, PICKING, ON_DELIVERY)
UPDATE sales_orders so
SET delivery_batch_id = di.delivery_batch_id,
    delivery_order_no = db.delivery_order_no,
    status = 'IN_DELIVERY'
FROM delivery_items di
JOIN delivery_batches db ON db.id = di.delivery_batch_id
WHERE di.sales_order_id = so.id
  AND db.status NOT IN ('COMPLETED', 'CANCELLED', 'REJECTED')
  AND (so.delivery_batch_id IS NULL OR so.status IN ('CONFIRMED', 'WAITING_STOCK'));
