-- Migration: 001_initial
-- Description: Initial schema for Setu API Gateway

-- Routes table
CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    path VARCHAR(1024) NOT NULL,
    path_type VARCHAR(50) NOT NULL DEFAULT 'prefix',
    methods TEXT[] NOT NULL DEFAULT '{}',
    strip_path BOOLEAN NOT NULL DEFAULT false,
    preserve_host BOOLEAN NOT NULL DEFAULT false,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    upstream_id UUID NOT NULL,
    auth_chain TEXT[] NOT NULL DEFAULT '{}',
    plugins TEXT[] NOT NULL DEFAULT '{}',
    rate_limit_id UUID,
    transform_config JSONB,
    timeout_interval INTERVAL DEFAULT '30 seconds',
    retry_enabled BOOLEAN NOT NULL DEFAULT true,
    circuit_breaker BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_routes_path ON routes(path);
CREATE INDEX idx_routes_enabled ON routes(enabled);
CREATE INDEX idx_routes_priority ON routes(priority DESC);

-- Upstreams table
CREATE TABLE IF NOT EXISTS upstreams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'round_robin',
    health_check_config JSONB,
    sticky_session BOOLEAN NOT NULL DEFAULT false,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Targets table
CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    upstream_id UUID NOT NULL REFERENCES upstreams(id) ON DELETE CASCADE,
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT true,
    healthy BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_targets_upstream_id ON targets(upstream_id);
CREATE INDEX idx_targets_enabled ON targets(enabled);

-- Auth providers table
CREATE TABLE IF NOT EXISTS auth_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    timeout_interval INTERVAL DEFAULT '5 seconds',
    cache_enabled BOOLEAN NOT NULL DEFAULT false,
    cache_ttl INTERVAL DEFAULT '5 minutes',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Route auth mapping table
CREATE TABLE IF NOT EXISTS route_auth (
    route_id UUID NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    auth_provider_id UUID NOT NULL REFERENCES auth_providers(id) ON DELETE CASCADE,
    priority INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (route_id, auth_provider_id)
);

CREATE INDEX idx_route_auth_route_id ON route_auth(route_id);

-- Plugins table
CREATE TABLE IF NOT EXISTS plugins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    route_id UUID REFERENCES routes(id) ON DELETE CASCADE,
    config JSONB NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plugins_route_id ON plugins(route_id);
CREATE INDEX idx_plugins_enabled ON plugins(enabled);

-- Rate limits table
CREATE TABLE IF NOT EXISTS rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id UUID REFERENCES routes(id) ON DELETE SET NULL,
    requests_per_second INTEGER NOT NULL,
    burst_size INTEGER NOT NULL,
    key_type VARCHAR(50) NOT NULL DEFAULT 'ip',
    key_header VARCHAR(255),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rate_limits_route_id ON rate_limits(route_id);

-- Access logs table (partitioned by month for performance)
CREATE TABLE IF NOT EXISTS access_logs (
    id BIGSERIAL,
    request_id UUID NOT NULL,
    route_id UUID,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(2048) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER NOT NULL,
    client_ip INET NOT NULL,
    user_agent TEXT,
    auth_provider VARCHAR(50),
    upstream_host VARCHAR(255),
    upstream_status INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Create index for access logs
CREATE INDEX idx_access_logs_timestamp ON access_logs(timestamp DESC);
CREATE INDEX idx_access_logs_route_id ON access_logs(route_id);
CREATE INDEX idx_access_logs_status_code ON access_logs(status_code);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_routes_updated_at BEFORE UPDATE ON routes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_upstreams_updated_at BEFORE UPDATE ON upstreams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_targets_updated_at BEFORE UPDATE ON targets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_auth_providers_updated_at BEFORE UPDATE ON auth_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_plugins_updated_at BEFORE UPDATE ON plugins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rate_limits_updated_at BEFORE UPDATE ON rate_limits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
