-- Sample data for testing Setu API Gateway metrics
-- Run this after migrations to populate the database with test data

-- Insert sample upstreams
INSERT INTO upstreams (id, name, description, algorithm, enabled) VALUES
  ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'user-service', 'User Management Service', 'round_robin', true),
  ('b2c3d4e5-f6a7-8901-bcde-f12345678901', 'product-service', 'Product Catalog Service', 'round_robin', true),
  ('c3d4e5f6-a7b8-9012-cdef-123456789012', 'order-service', 'Order Processing Service', 'round_robin', true),
  ('d4e5f6a7-b8c9-0123-defa-234567890123', 'mock-httpbin', 'HTTP Bin for Testing', 'round_robin', true)
ON CONFLICT (name) DO NOTHING;

-- Insert sample targets for each upstream
INSERT INTO targets (id, upstream_id, host, port, weight, enabled, healthy) VALUES
  -- User service targets
  ('e5f6a7b8-c9d0-1234-efab-345678901234', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'localhost', 8001, 1, true, true),
  ('f6a7b8c9-d0e1-2345-fabc-456789012345', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'localhost', 8002, 1, true, true),
  
  -- Product service targets
  ('a7b8c9d0-e1f2-3456-abcd-567890123456', 'b2c3d4e5-f6a7-8901-bcde-f12345678901', 'localhost', 8003, 1, true, true),
  
  -- Order service targets
  ('b8c9d0e1-f2a3-4567-bcde-678901234567', 'c3d4e5f6-a7b8-9012-cdef-123456789012', 'localhost', 8004, 1, true, true),
  ('c9d0e1f2-a3b4-5678-cdef-789012345678', 'c3d4e5f6-a7b8-9012-cdef-123456789012', 'localhost', 8005, 1, true, true),
  
  -- Mock HTTPBin target (for testing)
  ('d0e1f2a3-b4c5-6789-defa-890123456789', 'd4e5f6a7-b8c9-0123-defa-234567890123', 'httpbin.org', 443, 1, true, true)
ON CONFLICT DO NOTHING;

-- Insert sample routes
INSERT INTO routes (id, name, description, path, path_type, methods, enabled, priority, upstream_id, timeout_interval) VALUES
  -- User service routes
  ('11111111-1111-1111-1111-111111111111', 'Get Users', 'Retrieve all users', '/api/users', 'prefix', ARRAY['GET'], true, 10, 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', '30 seconds'),
  ('22222222-2222-2222-2222-222222222222', 'Get User by ID', 'Retrieve a specific user', '/api/users', 'prefix', ARRAY['GET'], true, 20, 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', '30 seconds'),
  ('33333333-3333-3333-3333-333333333333', 'Create User', 'Create a new user', '/api/users', 'prefix', ARRAY['POST'], true, 20, 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', '30 seconds'),
  
  -- Product service routes
  ('44444444-4444-4444-4444-444444444444', 'Get Products', 'Retrieve all products', '/api/products', 'prefix', ARRAY['GET'], true, 10, 'b2c3d4e5-f6a7-8901-bcde-f12345678901', '30 seconds'),
  ('55555555-5555-5555-5555-555555555555', 'Get Product by ID', 'Retrieve a specific product', '/api/products', 'prefix', ARRAY['GET'], true, 20, 'b2c3d4e5-f6a7-8901-bcde-f12345678901', '30 seconds'),
  
  -- Order service routes
  ('66666666-6666-6666-6666-666666666666', 'Get Orders', 'Retrieve all orders', '/api/orders', 'prefix', ARRAY['GET'], true, 10, 'c3d4e5f6-a7b8-9012-cdef-123456789012', '30 seconds'),
  ('77777777-7777-7777-7777-777777777777', 'Create Order', 'Create a new order', '/api/orders', 'prefix', ARRAY['POST'], true, 20, 'c3d4e5f6-a7b8-9012-cdef-123456789012', '60 seconds'),
  
  -- Health check route (for gateway self-monitoring)
  ('88888888-8888-8888-8888-888888888888', 'Gateway Health', 'Gateway health check endpoint', '/health', 'exact', ARRAY['GET'], true, 100, 'd4e5f6a7-b8c9-0123-defa-234567890123', '5 seconds'),
  
  -- Test route with HTTPBin
  ('99999999-9999-9999-9999-999999999999', 'HTTPBin Get', 'Test route to httpbin.org', '/test/httpbin', 'prefix', ARRAY['GET'], true, 10, 'd4e5f6a7-b8c9-0123-defa-234567890123', '30 seconds')
ON CONFLICT DO NOTHING;

-- Insert sample rate limits
INSERT INTO rate_limits (id, route_id, requests_per_second, burst_size, key_type, enabled) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 100, 150, 'ip', true),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444', 200, 300, 'ip', true),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', '66666666-6666-6666-6666-666666666666', 50, 75, 'ip', true)
ON CONFLICT DO NOTHING;

-- Verify data insertion
SELECT 'Upstreams:' as info;
SELECT id, name, algorithm, enabled FROM upstreams;

SELECT 'Targets:' as info;
SELECT id, host, port, weight, enabled, healthy FROM targets;

SELECT 'Routes:' as info;
SELECT id, name, path, methods, enabled, priority FROM routes ORDER BY priority DESC;

SELECT 'Rate Limits:' as info;
SELECT id, route_id, requests_per_second, burst_size, key_type FROM rate_limits;
