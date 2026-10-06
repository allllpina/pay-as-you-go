ALTER TABLE business_schema.usage_logs
ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'pending';
