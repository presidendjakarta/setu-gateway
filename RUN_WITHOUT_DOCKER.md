# Running Setu Gateway Without Docker

This guide shows you how to run Setu API Gateway locally without Docker.

---

## Prerequisites

### Required
1. **Go 1.22+** - [Download](https://golang.org/dl/)
2. **PostgreSQL 16+** - [Download](https://www.postgresql.org/download/)

### Optional
3. **Redis 7+** - [Download](https://redis.io/download/) (for distributed rate limiting)

---

## Quick Setup (Automated)

### Windows
```bash
# Run setup script
setup-windows.bat

# Start gateway
start.bat
```

### Linux/Mac
```bash
# Make scripts executable
chmod +x setup.sh start.sh

# Run setup script
./setup.sh

# Start gateway
./start.sh
```

---

## Manual Setup

### Step 1: Install PostgreSQL

#### Windows
1. Download PostgreSQL from https://www.postgresql.org/download/windows/
2. Run installer and remember your password
3. Add PostgreSQL to PATH: `C:\Program Files\PostgreSQL\16\bin`

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### macOS
```bash
brew install postgresql
brew services start postgresql
```

### Step 2: Create Database

```bash
# Login to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE setu_gateway;

# Exit
\q
```

**Or use one command:**
```bash
# Linux/Mac
PGPASSWORD=postgres psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;"

# Windows (PowerShell)
$env:PGPASSWORD="postgres"
psql -h localhost -U postgres -c "CREATE DATABASE setu_gateway;"
```

### Step 3: Run Migrations

```bash
# Linux/Mac
PGPASSWORD=postgres psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql

# Windows
set PGPASSWORD=postgres
psql -h localhost -U postgres -d setu_gateway -f migrations\001_initial.up.sql
```

### Step 4: (Optional) Install Redis

#### Windows
Download from: https://github.com/microsoftarchive/redis/releases

#### Ubuntu/Debian
```bash
sudo apt install redis-server
sudo systemctl start redis
sudo systemctl enable redis
```

#### macOS
```bash
brew install redis
brew services start redis
```

**Test Redis:**
```bash
redis-cli ping
# Should return: PONG
```

### Step 5: Build the Gateway

```bash
# Build binary
go build -o setu-gateway ./cmd/gateway

# Windows
go build -o setu-gateway.exe ./cmd/gateway
```

### Step 6: Run the Gateway

```bash
# Linux/Mac
./setu-gateway

# Windows
setu-gateway.exe
```

---

## Testing

### Health Check
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "timestamp": "2026-05-07T10:00:00Z",
  "version": "1.0.0"
}
```

### Add Test Data

```sql
-- Connect to database
psql -h localhost -U postgres -d setu_gateway

-- Create upstream
INSERT INTO upstreams (id, name, algorithm, enabled) 
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'test-service', 'round_robin', true);

-- Add target (example: httpbin.org)
INSERT INTO targets (id, upstream_id, host, port, weight, enabled, healthy)
VALUES (gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440000', 'httpbin.org', 80, 1, true, true);

-- Create route
INSERT INTO routes (
  id, name, path, path_type, methods, 
  upstream_id, strip_path, enabled, priority
) VALUES (
  gen_random_uuid(),
  'test-route',
  '/api',
  'prefix',
  ARRAY['GET', 'POST'],
  '550e8400-e29b-41d4-a716-446655440000',
  true,
  true,
  10
);

-- Exit
\q
```

### Test Route

```bash
# Restart gateway to load new routes
./setu-gateway

# Test in another terminal
curl http://localhost:8080/api/get
```

This will proxy the request to `http://httpbin.org/get`.

---

## Configuration

### Edit Config File

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
    password: your_password  # Change this!
    
cache:
  enabled: true
  type: memory  # Change to 'redis' if Redis is installed
```

### Use Custom Config Path

```bash
# Linux/Mac
export SETU_CONFIG=/path/to/config.yaml
./setu-gateway

# Windows
set SETU_CONFIG=C:\path\to\config.yaml
setu-gateway.exe
```

---

## Ports

| Service | Port | Description |
|---------|------|-------------|
| Gateway | 8080 | Main API gateway |
| Admin | 8081 | Admin API |
| Metrics | 9090 | Prometheus metrics |

---

## Troubleshooting

### PostgreSQL Connection Failed

**Error:** `failed to ping database`

**Solution:**
```bash
# Check if PostgreSQL is running
# Linux
systemctl status postgresql

# Windows
Get-Service postgresql*

# Start PostgreSQL
# Linux
sudo systemctl start postgresql

# Windows (Admin)
Start-Service postgresql-x64-16
```

### Database Does Not Exist

**Error:** `database "setu_gateway" does not exist`

**Solution:**
```bash
psql -U postgres -c "CREATE DATABASE setu_gateway;"
```

### Permission Denied

**Error:** `permission denied for database`

**Solution:**
```bash
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE setu_gateway TO postgres;"
```

### Port Already in Use

**Error:** `address already in use`

**Solution:**
```bash
# Find process using port 8080
# Linux
lsof -i :8080
sudo kill -9 <PID>

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Mac
lsof -ti:8080 | xargs kill -9
```

### Go Build Failed

**Error:** `package not found`

**Solution:**
```bash
# Download dependencies
go mod tidy

# Clean build cache
go clean -cache

# Rebuild
go build -o setu-gateway ./cmd/gateway
```

---

## Development Mode

### Watch for File Changes

Install air for live reload:
```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml`:
```toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/setu-gateway ./cmd/gateway"
bin = "./tmp/setu-gateway"
full_bin = "./tmp/setu-gateway"
include_ext = ["go", "yaml"]
exclude_dir = ["tmp", "vendor"]
```

Run with live reload:
```bash
air
```

### Debug Mode

```bash
# Enable debug logging
# Edit configs/gateway.yaml
logging:
  level: debug
  format: console  # Easier to read

# Run with debug
./setu-gateway
```

---

## Production Deployment (Without Docker)

### Systemd Service (Linux)

Create `/etc/systemd/system/setu-gateway.service`:

```ini
[Unit]
Description=Setu API Gateway
After=network.target postgresql.service

[Service]
Type=simple
User=setu
Group=setu
WorkingDirectory=/opt/setu-gateway
ExecStart=/opt/setu-gateway/setu-gateway
Restart=on-failure
RestartSec=5

# Security
NoNewPrivileges=true
ProtectSystem=strict
ReadWritePaths=/opt/setu-gateway

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable setu-gateway
sudo systemctl start setu-gateway

# Check status
sudo systemctl status setu-gateway

# View logs
sudo journalctl -u setu-gateway -f
```

### Windows Service

Use NSSM (Non-Sucking Service Manager):

```bash
# Download NSSM from https://nssm.cc/download

# Install service
nssm install SetuGateway C:\path\to\setu-gateway.exe

# Start service
nssm start SetuGateway

# Check status
nssm status SetuGateway
```

---

## Performance Tuning

### PostgreSQL Optimization

```sql
-- Increase connection limit
ALTER SYSTEM SET max_connections = 200;

-- Increase shared buffers
ALTER SYSTEM SET shared_buffers = '1GB';

-- Reload config
SELECT pg_reload_conf();
```

### Gateway Optimization

Edit `configs/gateway.yaml`:

```yaml
proxy:
  transport:
    max_idle_conns: 1000
    max_idle_conns_per_host: 100
    idle_conn_timeout: 90s
    
database:
  postgres:
    max_open_conns: 50
    max_idle_conns: 10
```

### OS Limits (Linux)

```bash
# Increase file descriptor limit
ulimit -n 65536

# Make permanent
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf
```

---

## Monitoring

### Check Logs

```bash
# Gateway logs are in stdout
# Redirect to file
./setu-gateway > gateway.log 2>&1

# View logs
tail -f gateway.log
```

### Check Database

```bash
# Connection count
psql -U postgres -d setu_gateway -c "SELECT count(*) FROM pg_stat_activity;"

# Table sizes
psql -U postgres -d setu_gateway -c "
  SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
  FROM pg_tables
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
"
```

### Health Checks

```bash
# Gateway health
curl http://localhost:8080/health

# Admin health
curl http://localhost:8081/health
```

---

## Uninstall

### Remove Binary
```bash
# Linux/Mac
rm setu-gateway

# Windows
del setu-gateway.exe
```

### Drop Database
```bash
psql -U postgres -c "DROP DATABASE setu_gateway;"
```

### Remove Redis Data (if installed)
```bash
redis-cli FLUSHALL
```

---

## Next Steps

1. **Add Routes** - Insert routes into PostgreSQL
2. **Configure Auth** - Setup JWT or API Key authentication
3. **Add Rate Limiting** - Protect your APIs
4. **Setup Monitoring** - Configure Prometheus & Grafana
5. **Deploy to Production** - Use systemd or Windows service

---

## Support

- Documentation: README.md
- Quick Start: QUICKSTART.md
- Issues: https://github.com/presidendjakarta/setu-gateway/issues
