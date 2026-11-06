-- Migration Down: Remove sales_tester_requests and delivery_batch_samples tables
-- Version: 000060

DROP TABLE IF EXISTS delivery_batch_samples;
DROP TABLE IF EXISTS sales_tester_requests;
