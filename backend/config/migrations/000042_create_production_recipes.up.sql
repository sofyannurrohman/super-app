-- Create production_recipes table
CREATE TABLE IF NOT EXISTS production_recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    finished_product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE production_recipes DROP CONSTRAINT IF EXISTS production_recipes_finished_product_id_key CASCADE;
ALTER TABLE production_recipes DROP CONSTRAINT IF EXISTS uni_production_recipes_finished_product_id CASCADE;
DROP INDEX IF EXISTS uni_production_recipes_finished_product_id CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS uni_production_recipes_finished_product_id ON production_recipes(finished_product_id);

-- Create production_recipe_items table
CREATE TABLE IF NOT EXISTS production_recipe_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id UUID NOT NULL REFERENCES production_recipes(id) ON DELETE CASCADE,
    raw_product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity DECIMAL(10,4) NOT NULL
);
