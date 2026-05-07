# Setu API Gateway

Production-grade API Gateway built with Golang, designed for high-performance, scalability, and enterprise use cases.

## Architecture Overview

Setu Gateway follows Clean Architecture principles with a modular, extensible design:

```
┌─────────────────────────────────────────────────┐
│                  Client Request                  │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│              Gateway Server (HTTP/2)             │
│  - TLS Support                                   │
│  - Graceful Shutdown                             │
│  - Connection Management                         │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│              Middleware Pipeline                 │
│  - Recovery                                      │
│  - Request ID                                    │
│  - CORS                                          │
│  - Rate Limiting                                 │
│  - Authentication                                │
│  - Logging                                       │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│              Router (Radix Tree)                 │
│  - Path Matching (Exact/Prefix/Regex/Wildcard)  │
│  - Method-based Routing                          │
│  - Priority Resolution                           │
│  - High-performance O(k) lookup                  │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│              Load Balancer                       │
│  - Round Robin                                   │
│  - Weighted Round Robin                          │
│  - Least Connection                              │
│  - Health-aware Selection                        │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│           Reverse Proxy (HTTP/1.1, HTTP/2)       │
│  - Connection Pooling                            │
│  - Request/Response Transformation               │
│  - WebSocket Support                             │
│  - gRPC Support                                  │
│  - Timeout Management                            │
└──────────────────┬──────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────┐
│                 Upstream Services                │
└─────────────────────────────────────────────────┘
```

## Features

### Core Features
- ✅ **Reverse Proxy** - HTTP/1.1 and HTTP/2 support
- ✅ **Dynamic Routing** - Radix tree-based high-performance routing
- ✅ **Multiple Path Types** - Exact, Prefix, Regex, Wildcard
- ✅ **Request/Response Transformation** - Path, Header, Method, Query rewriting
- ✅ **Connection Pooling** - Optimized HTTP transport with keep-alive
- ✅ **Graceful Shutdown** - Zero-downtime deployments

### Authentication
- 🔄 JWT (RS256/ES256/HS256)
- 🔄 API Key
- 🔄 OAuth2 Introspection
- 🔄 Basic Auth
- 🔄 mTLS
- 🔄 HMAC Signature
- 🔄 External Auth Service
- 🔄 Auth Chaining & Caching

### Traffic Management
- ✅ **Load Balancing** - Round Robin, Weighted RR, Least Connection
- 🔄 **Circuit Breaker** - Resilience pattern
- 🔄 **Retry Mechanism** - Exponential backoff with jitter
- 🔄 **Rate Limiting** - Token bucket, sliding window
- 🔄 **Timeout Management** - Per-route timeouts

### Service Discovery
- 🔄 Static Configuration
- 🔄 Consul
- 🔄 Etcd
- 🔄 Kubernetes

### Observability
- ✅ **Structured Logging** - Zap logger with JSON format
- 🔄 **Distributed Tracing** - OpenTelemetry integration
- 🔄 **Prometheus Metrics** - Request metrics, performance monitoring
- 🔄 **Health Checks** - Liveness, readiness, upstream health
- 🔄 **Access Logs** - Detailed request logging

### Plugin System
- 🔄 Middleware Plugins
- 🔄 Auth Plugins
- 🔄 Transform Plugins
- 🔄 WASM Plugin Support (wazero runtime)
- 🔄 Dynamic Plugin Loading

### Admin Dashboard
- 🔄 Route Management
- 🔄 Upstream Management
- 🔄 Plugin Management
- AuthProvider Management
- 🔄 Metrics Dashboard
- 🔄 Live Request Logs
- 🔄 Config Editor

## Project Structure

```
setu-gateway/
├── cmd/
│   └── gateway/
│       └── main.go                 # Entry point
├── internal/
│   ├── gateway/
│   │   ├── gateway.go              # Main gateway handler
│   │   └── interfaces.go           # Core interfaces
│   ├── router/
│   │   └── router.go               # Radix tree router
│   ├── proxy/
│   │   └── proxy.go                # Reverse proxy
│   ├── loadbalancer/
│   │   └── roundrobin.go           # Load balancers
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── database/
│   │   └── postgres.go             # PostgreSQL connection
│   ├── repository/
│   │   ├── interfaces.go           # Repository interfaces
│   │   └── postgres/
│   │       └── route_repo.go       # Route repository
│   ├── logger/
│   │   └── logger.go               # Zap logger wrapper
│   ├── di/
│   │   └── container.go            # Dependency injection
│   ├── auth/                       # Authentication (TODO)
│   ├── ratelimit/                  # Rate limiting (TODO)
│   ├── circuitbreaker/             # Circuit breaker (TODO)
│   ├── discovery/                  # Service discovery (TODO)
│   ├── plugin/                     # Plugin system (TODO)
│   ├── middleware/                 # Middleware pipeline (TODO)
│   ├── metrics/                    # Prometheus metrics (TODO)
│   ├── tracing/                    # Distributed tracing (TODO)
│   ├── health/                     # Health checks (TODO)
│   ├── admin/                      # Admin API (TODO)
│   └── streaming/                  # Live streaming (TODO)
├── pkg/
│   └── types/
│       ├── types.go                # Domain types
│       └── errors.go               # Error definitions
├── configs/
│   ├── gateway.yaml                # Gateway configuration
│   └── prometheus.yml              # Prometheus config
├── migrations/
│   └── 001_initial.up.sql          # Database schema
├── web/
│   └── admin/                      # Next.js admin dashboard (TODO)
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL 16+
- Redis 7+ (optional, for caching and rate limiting)
- Docker & Docker Compose (optional, for containerized deployment)

### Option 1: Without Docker (Recommended for Development)

**Windows:**
```bash
# Automated setup
setup-windows.bat

# Start gateway
start.bat
```

**Linux/Mac:**
```bash
# Automated setup
chmod +x setup.sh start.sh
./setup.sh

# Start gateway
./start.sh
```

**Manual Setup:**
```bash
# 1. Install PostgreSQL and create database
psql -U postgres -c "CREATE DATABASE setu_gateway;"

# 2. Run migrations
psql -U postgres -d setu_gateway -f migrations/001_initial.up.sql

# 3. Build
go build -o setu-gateway ./cmd/gateway

# 4. Run
./setu-gateway
```

📖 **Complete Guide**: See [RUN_WITHOUT_DOCKER.md](RUN_WITHOUT_DOCKER.md)

### Option 2: Using Docker Compose

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f gateway

# Stop services
docker-compose down
```

### Manual Setup

1. **Start PostgreSQL:**
```bash
docker run -d --name setu-db \
  -e POSTGRES_DB=setu_gateway \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```

2. **Run migrations:**
```bash
psql -h localhost -U postgres -d setu_gateway -f migrations/001_initial.up.sql
```

3. **Build and run:**
```bash
# Build
go build -o setu-gateway ./cmd/gateway

# Run
./setu-gateway
```

### Configuration

Edit `configs/gateway.yaml`:

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

Set custom config path:
```bash
export SETU_CONFIG=/path/to/your/config.yaml
```

## API Usage

### Adding Routes

Routes are stored in PostgreSQL and loaded on startup. Example route insertion:

```sql
INSERT INTO routes (
  id, name, path, path_type, methods, upstream_id, enabled
) VALUES (
  gen_random_uuid(),
  'api-route',
  '/api',
  'prefix',
  ARRAY['GET', 'POST'],
  '<upstream-uuid>',
  true
);
```

### Upstream Configuration

```sql
-- Create upstream
INSERT INTO upstreams (id, name, algorithm, enabled) 
VALUES (gen_random_uuid(), 'my-service', 'round_robin', true);

-- Add targets
INSERT INTO targets (upstream_id, host, port, weight, enabled, healthy)
VALUES 
  ('<upstream-uuid>', '10.0.0.1', 8080, 1, true, true),
  ('<upstream-uuid>', '10.0.0.2', 8080, 1, true, true);
```

## Architecture Decisions

### Why Radix Tree for Routing?
- O(k) lookup time where k is the path length
- Memory efficient with prefix compression
- Fast prefix matching for REST APIs
- Better performance than regex-based routing

### Why Manual DI?
- Zero dependencies
- Better performance than reflection-based DI
- Explicit dependency graph
- Easier to test and mock

### Why pgx over database/sql?
- Better performance
- Native PostgreSQL features support
- Connection pooling built-in
- Lower memory allocation

### Why Zap Logger?
- Structured logging
- High performance
- Low allocation
- JSON output for log aggregation

## Performance Optimizations

1. **Connection Pooling** - Reuse HTTP and database connections
2. **Buffer Pooling** - sync.Pool for request buffers
3. **Lock-free Reads** - RWMutex for concurrent config access
4. **Radix Tree** - Fast O(k) route matching
5. **HTTP/2** - Multiplexed connections
6. **Streaming** - FlushInterval=-1 for response streaming
7. **Context Propagation** - Proper cancellation and timeouts

## Concurrency Model

- **Reader-Writer Locks** - Config and route access
- **Atomic Operations** - Counters and hot-reload swaps
- **Channel-based Communication** - Events and signals
- **Context Cancellation** - Request lifecycle management
- **Thread-safe Data Structures** - Mutex-protected maps

## Security Features

- TLS/HTTPS support
- mTLS for upstream communication
- Multiple authentication providers
- Rate limiting per client/route
- IP filtering
- Request size limits
- Timeout enforcement
- Secure config handling

## Monitoring

### Health Checks
```bash
curl http://localhost:8080/health
```

### Metrics (Prometheus)
```bash
curl http://localhost:9090/metrics
```

### Logs
Structured JSON logs with:
- Request ID
- Route information
- Response time
- Upstream details
- Error details

## Roadmap

### Phase 1 - Core Infrastructure ✅
- [x] Project structure
- [x] Configuration management
- [x] Database connection
- [x] Repository layer
- [x] Router implementation
- [x] Reverse proxy
- [x] Load balancer
- [x] Main entry point
- [x] Docker setup

### Phase 2 - Security & Traffic Management
- [ ] Authentication providers
- [ ] Rate limiting
- [ ] Circuit breaker
- [ ] Retry mechanism

### Phase 3 - Advanced Features
- [ ] Service discovery
- [ ] Plugin system
- [ ] WASM support
- [ ] WebSocket proxy
- [ ] gRPC proxy

### Phase 4 - Observability
- [ ] Prometheus metrics
- [ ] Distributed tracing
- [ ] Advanced health checks
- [ ] Live metrics streaming

### Phase 5 - Admin Dashboard
- [ ] Admin API
- [ ] Next.js frontend
- [ ] Real-time monitoring
- [ ] Configuration UI

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - See LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: https://github.com/presidendjakarta/setu-gateway/issues
- Documentation: https://github.com/presidendjakarta/setu-gateway/wiki

## Acknowledgments

Inspired by:
- Kong
- Traefik
- Envoy
- Caddy
- NGINX

Built with:
- Go standard library
- Chi router patterns
- pgx for PostgreSQL
- Zap for logging
