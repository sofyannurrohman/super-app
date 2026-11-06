DROP INDEX IF EXISTS idx_marketplace_tokens_mp_branch;
CREATE UNIQUE INDEX IF NOT EXISTS idx_marketplace_shop ON marketplace_tokens(marketplace, shop_id);
