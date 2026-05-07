# Quick Start Guide - Setu API Gateway

## Prerequisites

- Docker & Docker Compose installed
- OR Go 1.22+ and PostgreSQL 16+ installed locally

## Option 1: Docker Compose (Recommended)

### 1. Start All Services

```bash
cd x:\laragon\go-apps\setu-gateway
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Setu Gateway (port 8080, 8081, 9090)
- Prometheus (port 9091)
- Jaeger UI (port 16686)

### 2. Verify Services

```bash
# Check gateway health
curl http://localhost:8080/health

# Check admin server
curl http://localhost:8081/health

# View gateway logs
docker-compose logs -f gateway
```

### 3. Stop Services

```bash
docker-compose down
```

## Option 2: Manual Setup

### 1. Start PostgreSQL

```bash
docker run -d --name setu-db \
  -e POSTGRES_DB=setu_gateway \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```

### 2. Run Database Migrations

```bash
psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql
```

### 3. Build the Gateway

```bash
cd x:\laragon\go-apps\setu-gateway
go build -o setu-gateway ./cmd/gateway
```

### 4. Run the Gateway

```bash
# Linux/Mac
./setu-gateway

# Windows
setu-gateway.exe
```

### 5. Test the Gateway

```bash
# Health check
curl http://localhost:8080/health

# Should return:
# {"status":"ok","timestamp":"2026-05-07T...","version":"1.0.0"}
```

## Configuration

### Edit Configuration File

Open `configs/gateway.yaml` and modify as needed:

```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  postgres:
    host: localhost
    port: 5432
    name: setu_gateway
    user: postgres
    password: postgres
```

### Use Custom Config Path

```bash
export SETU_CONFIG=/path/to/custom/config.yaml
./setu-gateway
```

## Adding Your First Route

### 1. Create an Upstream

```sql
-- Insert upstream
INSERT INTO upstreams (id, name, algorithm, enabled) 
VALUES (
  '550e8400-e29b-41d4-a716-446655440000',
  'my-backend-service',
  'round_robin',
  true
);

-- Add targets
INSERT INTO targets (id, upstream_id, host, port, weight, enabled, healthy)
VALUES 
  (gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440000', '10.0.0.1', 8080, 1, true, true),
  (gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440000', '10.0.0.2', 8080, 1, true, true);
```

### 2. Create a Route

```sql
INSERT INTO routes (
  id, name, description, path, path_type, methods, 
  upstream_id, strip_path, enabled, priority
) VALUES (
  gen_random_uuid(),
  'api-route',
  'Route to backend API',
  '/api',
  'prefix',
  ARRAY['GET', 'POST', 'PUT', 'DELETE'],
  '550e8400-e29b-41d4-a716-446655440000',
  true,
  true,
  10
);
```

### 3. Restart Gateway

```bash
# The gateway loads routes on startup
# Restart to pick up new routes
./setu-gateway
```

### 4. Test the Route

```bash
# This will be proxied to your upstream
curl http://localhost:8080/api/users
```

## Ports Reference

| Service | Port | Description |
|---------|------|-------------|
| Gateway | 8080 | Main API gateway |
| Admin | 8081 | Admin API |
| Metrics | 9090 | Prometheus metrics |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache/Rate limiting |
| Prometheus | 9091 | Metrics UI |
| Jaeger | 16686 | Tracing UI |

## Troubleshooting

### Gateway Won't Start

**Issue**: Database connection error

**Solution**:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
psql -h localhost -U postgres -d setu_gateway -c "SELECT 1"
```

### Routes Not Working

**Issue**: 404 Route not found

**Solution**:
```sql
-- Check if routes exist
SELECT id, name, path, enabled FROM routes;

-- Check if upstream has targets
SELECT t.host, t.port, t.healthy 
FROM targets t 
JOIN upstreams u ON t.upstream_id = u.id 
WHERE u.name = 'my-backend-service';
```

### High Memory Usage

**Solution**:
```yaml
# configs/gateway.yaml
proxy:
  transport:
    max_idle_conns: 100  # Reduce from 1000
    max_idle_conns_per_host: 10  # Reduce from 100
```

## Next Steps

1. **Add Authentication** - See documentation for JWT/API Key setup
2. **Configure Rate Limiting** - Protect your APIs
3. **Set Up Monitoring** - Configure Prometheus and Grafana
4. **Deploy to Production** - Use Kubernetes manifests (coming soon)
5. **Build Admin Dashboard** - Run Next.js frontend (coming soon)

## Useful Commands

```bash
# View all containers
docker-compose ps

# View logs
docker-compose logs -f gateway
docker-compose logs -f postgres

# Rebuild gateway
docker-compose build gateway
docker-compose up -d gateway

# Reset database
docker-compose down -v
docker-compose up -d postgres
psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql

# Check gateway stats
curl http://localhost:8080/health | jq
```

## Support

- Documentation: README.md
- Progress: IMPLEMENTATION_PROGRESS.md
- Issues: https://github.com/presidendjakarta/setu-gateway/issues
