-- Drop old unique constraint on (marketplace, shop_id)
DROP INDEX IF EXISTS idx_marketplace_shop;
DROP INDEX IF EXISTS idx_marketplace_tokens_mp_branch;

-- Create unique constraint on (marketplace, branch_id)
CREATE UNIQUE INDEX IF NOT EXISTS idx_marketplace_tokens_mp_branch ON marketplace_tokens(marketplace, COALESCE(branch_id, '00000000-0000-0000-0000-000000000000'::uuid));
