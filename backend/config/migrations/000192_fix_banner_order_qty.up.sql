UPDATE banner_orders SET qty = 1 WHERE qty IS NULL OR qty <= 0;
