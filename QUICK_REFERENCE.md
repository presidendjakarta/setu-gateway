# 🚀 Setu Gateway - Quick Reference

## Cara Menjalankan (Tanpa Docker)

### Windows - Paling Mudah
```bash
# 1. Setup (sekali saja)
setup-windows.bat

# 2. Start gateway
start.bat
```

### Linux/Mac - Paling Mudah
```bash
# 1. Setup (sekali saja)
chmod +x setup.sh start.sh
./setup.sh

# 2. Start gateway
./start.sh
```

### Menggunakan Make (Semua Platform)
```bash
# Setup
make setup

# Build & Run
make run

# Atau development mode (auto reload)
make dev
```

### Manual
```bash
# 1. Install PostgreSQL
# Download dari: https://www.postgresql.org/download/

# 2. Create database
psql -U postgres -c "CREATE DATABASE setu_gateway;"

# 3. Run migrations
psql -U postgres -d setu_gateway -f migrations/001_initial.up.sql

# 4. Build
go build -o setu-gateway ./cmd/gateway

# 5. Run
./setu-gateway
```

---

## Cara Menjalankan (Dengan Docker)

```bash
# Start semua services
docker-compose up -d

# Lihat logs
docker-compose logs -f gateway

# Stop
docker-compose down
```

---

## Testing

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","timestamp":"...","version":"1.0.0"}
```

---

## Ports

| Service | Port | URL |
|---------|------|-----|
| Gateway | 8080 | http://localhost:8080 |
| Admin | 8081 | http://localhost:8081 |
| Metrics | 9090 | http://localhost:9090 |

---

## File Penting

| File | Fungsi |
|------|--------|
| `configs/gateway.yaml` | Konfigurasi gateway |
| `migrations/001_initial.up.sql` | Database schema |
| `setup-windows.bat` | Setup otomatis Windows |
| `setup.sh` | Setup otomatis Linux/Mac |
| `start.bat` | Start gateway Windows |
| `start.sh` | Start gateway Linux/Mac |
| `Makefile` | Commands untuk development |

---

## Make Commands

```bash
make help          # Lihat semua commands
make setup         # Setup dependencies & database
make build         # Build binary
make run           # Build & run gateway
make test          # Run tests
make dev           # Development mode (auto reload)
make clean         # Clean build artifacts
make migrate       # Run migrations
make docker-up     # Start Docker services
make docker-down   # Stop Docker services
make check-deps    # Cek dependencies
```

---

## Troubleshooting

### PostgreSQL tidak bisa connect
```bash
# Cek PostgreSQL running
# Windows
Get-Service postgresql*

# Linux
systemctl status postgresql

# Mac
brew services list | grep postgresql

# Start PostgreSQL
# Windows
Start-Service postgresql-x64-16

# Linux
sudo systemctl start postgresql

# Mac
brew services start postgresql
```

### Port sudah dipakai
```bash
# Cek port 8080
# Windows
netstat -ano | findstr :8080

# Linux/Mac
lsof -i :8080

# Kill process
# Windows
taskkill /PID <PID> /F

# Linux/Mac
kill -9 <PID>
```

### Build error
```bash
# Download dependencies
go mod tidy

# Clean cache
go clean -cache

# Rebuild
go build -o setu-gateway ./cmd/gateway
```

---

## Add Sample Route

```sql
-- Connect to database
psql -U postgres -d setu_gateway

-- Create upstream
INSERT INTO upstreams (id, name, algorithm, enabled) 
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'test-service', 'round_robin', true);

-- Add target (httpbin.org)
INSERT INTO targets (id, upstream_id, host, port, weight, enabled, healthy)
VALUES (gen_random_uuid(), '550e8400-e29b-41d4-a716-446655440000', 'httpbin.org', 80, 1, true, true);

-- Create route
INSERT INTO routes (id, name, path, path_type, methods, upstream_id, strip_path, enabled, priority)
VALUES (gen_random_uuid(), 'test-route', '/api', 'prefix', ARRAY['GET', 'POST'], '550e8400-e29b-41d4-a716-446655440000', true, true, 10);
```

Test route:
```bash
curl http://localhost:8080/api/get
```

---

## Configuration

Edit `configs/gateway.yaml`:

```yaml
server:
  port: 8080  # Ganti port

database:
  postgres:
    password: your_password  # Ganti password

logging:
  level: debug  # debug, info, warn, error
  format: console  # console atau json
```

---

## Next Steps

1. ✅ Gateway sudah running
2. 📝 Tambah routes di database
3. 🔒 Setup authentication (JWT, API Key)
4. 🚦 Setup rate limiting
5. 📊 Setup monitoring (Prometheus, Grafana)
6. 🚀 Deploy to production

---

## Dokumentasi Lengkap

- **README.md** - Dokumentasi utama
- **QUICKSTART.md** - Quick start guide
- **RUN_WITHOUT_DOCKER.md** - Panduan lengkap tanpa Docker
- **PROJECT_SUMMARY.md** - Summary project
- **ROADMAP.md** - Development roadmap
- **IMPLEMENTATION_PROGRESS.md** - Progress tracker

---

## Support

- GitHub: https://github.com/presidendjakarta/setu-gateway
- Issues: https://github.com/presidendjakarta/setu-gateway/issues
