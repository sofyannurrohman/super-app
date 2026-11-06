ALTER TABLE product_transfers 
ADD COLUMN IF NOT EXISTS fulfillment_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS is_stock_borrowed BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS replenishment_transfer_id UUID REFERENCES product_transfers(id) ON DELETE SET NULL;

