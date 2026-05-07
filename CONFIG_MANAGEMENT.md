# Configuration Management Guide

## 🎯 **Important: Config is NOT Compiled!**

**Config file YAML is read at RUNTIME, not at compile time!**

```bash
# Compile (once)
go build -o setu-gateway.exe ./cmd/gateway
# Binary is created WITHOUT config inside

# Run (can change config anytime)
./setu-gateway.exe                    # Uses configs/gateway.yaml
SETU_CONFIG=prod.yaml ./setu-gateway.exe  # Uses prod.yaml
```

**You can change database config without recompiling!** ✅

---

## 📋 **Ways to Change Database Config**

### **Method 1: Environment Variables (Recommended)**

You can override ANY config value using environment variables:

#### **Windows PowerShell:**
```powershell
# Set database config
$env:SETU_DB_HOST="172.16.100.250"
$env:SETU_DB_PORT="5432"
$env:SETU_DB_NAME="setu_gateway"
$env:SETU_DB_USER="setu_gateway"
$env:SETU_DB_PASSWORD="setu_gateway1324!!"

# Run gateway (uses env vars instead of YAML)
.\setu-gateway.exe
```

#### **Linux/Mac:**
```bash
# Set database config
export SETU_DB_HOST="172.16.100.250"
export SETU_DB_PORT="5432"
export SETU_DB_NAME="setu_gateway"
export SETU_DB_USER="setu_gateway"
export SETU_DB_PASSWORD="setu_gateway1324!!"

# Run gateway
./setu-gateway
```

#### **All Available Environment Variables:**

| Variable | Description | Example |
|----------|-------------|---------|
| `SETU_DB_HOST` | Database host | `172.16.100.250` |
| `SETU_DB_PORT` | Database port | `5432` |
| `SETU_DB_NAME` | Database name | `setu_gateway` |
| `SETU_DB_USER` | Database user | `setu_gateway` |
| `SETU_DB_PASSWORD` | Database password | `secret` |
| `SETU_DB_SSL_MODE` | SSL mode | `disable`, `require` |
| `SETU_SERVER_HOST` | Server host | `0.0.0.0` |
| `SETU_SERVER_PORT` | Server port | `8080` |
| `SETU_ADMIN_API_KEY` | Admin API key | `my-secret-key` |
| `SETU_LOG_LEVEL` | Log level | `debug`, `info`, `warn`, `error` |
| `SETU_LOG_FORMAT` | Log format | `json`, `console` |
| `SETU_REDIS_HOST` | Redis host | `localhost` |
| `SETU_REDIS_PORT` | Redis port | `6379` |
| `SETU_REDIS_PASSWORD` | Redis password | `secret` |

**Priority:** Environment variables > YAML file

---

### **Method 2: Multiple Config Files**

Create different config files for different environments:

#### **Development Config:**
`configs/development.yaml`
```yaml
database:
  postgres:
    host: localhost
    port: 5432
    name: setu_gateway_dev
    user: postgres
    password: postgres
    ssl_mode: disable

logging:
  level: debug
  format: console
```

#### **Production Config:**
`configs/production.yaml`
```yaml
database:
  postgres:
    host: 172.16.100.250
    port: 5432
    name: setu_gateway_prod
    user: setu_gateway
    password: super_secret_password
    ssl_mode: require

logging:
  level: warn
  format: json
```

#### **Run with specific config:**

**Windows:**
```powershell
$env:SETU_CONFIG="configs\production.yaml"
.\setu-gateway.exe
```

**Linux/Mac:**
```bash
SETU_CONFIG=configs/production.yaml ./setu-gateway
```

---

### **Method 3: Edit YAML File Directly**

Simply edit `configs/gateway.yaml`:

```yaml
database:
  postgres:
    host: 172.16.100.250      # Change this
    port: 5432
    name: setu_gateway         # Change this
    user: setu_gateway         # Change this
    password: your_password    # Change this
```

**Restart gateway to apply changes.**

---

## 🔄 **Hot-Reload Config**

Gateway supports **config hot-reload**! When you edit `configs/gateway.yaml`, changes are applied automatically without restart.

```bash
# Start gateway
./setu-gateway.exe

# Edit config in another terminal
# (configs/gateway.yaml is watched automatically)

# Changes applied automatically!
```

**Note:** Some changes (like database connection) may require restart.

---

## 🎭 **Environment-Specific Setup**

### **Development Environment**

```bash
# Use localhost
export SETU_DB_HOST="localhost"
export SETU_DB_NAME="setu_gateway_dev"
export SETU_DB_USER="postgres"
export SETU_DB_PASSWORD="postgres"
export SETU_LOG_LEVEL="debug"
export SETU_LOG_FORMAT="console"

./setu-gateway
```

### **Staging Environment**

```bash
# Use staging server
export SETU_DB_HOST="staging-db.example.com"
export SETU_DB_NAME="setu_gateway_staging"
export SETU_DB_USER="setu_gateway"
export SETU_DB_PASSWORD="staging_password"
export SETU_LOG_LEVEL="info"
export SETU_LOG_FORMAT="json"

./setu-gateway
```

### **Production Environment**

```bash
# Use production server
export SETU_DB_HOST="prod-db.example.com"
export SETU_DB_NAME="setu_gateway_prod"
export SETU_DB_USER="setu_gateway"
export SETU_DB_PASSWORD="production_secret"
export SETU_DB_SSL_MODE="require"
export SETU_LOG_LEVEL="warn"
export SETU_LOG_FORMAT="json"

./setu-gateway
```

---

## 📦 **Using .env Files**

Create `.env` file for easy management:

### **.env.development**
```env
SETU_DB_HOST=localhost
SETU_DB_PORT=5432
SETU_DB_NAME=setu_gateway_dev
SETU_DB_USER=postgres
SETU_DB_PASSWORD=postgres
SETU_LOG_LEVEL=debug
SETU_LOG_FORMAT=console
```

### **.env.production**
```env
SETU_DB_HOST=172.16.100.250
SETU_DB_PORT=5432
SETU_DB_NAME=setu_gateway
SETU_DB_USER=setu_gateway
SETU_DB_PASSWORD=setu_gateway1324!!
SETU_DB_SSL_MODE=disable
SETU_LOG_LEVEL=info
SETU_LOG_FORMAT=json
```

### **Load .env file:**

**Using dotenv CLI:**
```bash
# Install
npm install -g dotenv-cli

# Run with .env file
dotenv -e .env.production -- ./setu-gateway
```

**Manual (PowerShell):**
```powershell
# Load .env file
Get-Content .env.production | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        [Environment]::SetEnvironmentVariable($matches[1], $matches[2])
    }
}

.\setu-gateway.exe
```

**Manual (Bash):**
```bash
# Load .env file
set -a
source .env.production
set +a

./setu-gateway
```

---

## 🔐 **Security Best Practices**

### **1. Never Commit Passwords**

Add to `.gitignore`:
```gitignore
# Environment files
.env
.env.*
!.env.example

# Production configs
configs/production.yaml
configs/*production*
```

### **2. Use .env.example**

Create `.env.example` (safe to commit):
```env
SETU_DB_HOST=localhost
SETU_DB_PORT=5432
SETU_DB_NAME=setu_gateway
SETU_DB_USER=your_username
SETU_DB_PASSWORD=your_password
SETU_DB_SSL_MODE=disable
```

### **3. Use Secrets Manager (Production)**

For production, use secrets manager:
- AWS Secrets Manager
- HashiCorp Vault
- Azure Key Vault
- Kubernetes Secrets

Example with Vault:
```bash
# Get secrets from Vault
export SETU_DB_PASSWORD=$(vault kv get -field=password secret/database)

./setu-gateway
```

---

## 🧪 **Testing Different Configs**

### **Test with different database:**

```bash
# Test with localhost
export SETU_DB_HOST="localhost"
./setu-gateway

# Test with remote server (Ctrl+C to stop first)
export SETU_DB_HOST="172.16.100.250"
./setu-gateway
```

### **Verify config is applied:**

Gateway logs show config on startup:
```
INFO Starting Setu API Gateway
INFO Database connection established
INFO Database: setu_gateway@172.16.100.250:5432
```

---

## 🚀 **Deployment Scripts**

### **deploy.sh (Linux/Mac)**
```bash
#!/bin/bash

# Load production config
export SETU_DB_HOST="172.16.100.250"
export SETU_DB_NAME="setu_gateway"
export SETU_DB_USER="setu_gateway"
export SETU_DB_PASSWORD="production_secret"
export SETU_LOG_LEVEL="warn"
export SETU_LOG_FORMAT="json"

# Start gateway
./setu-gateway
```

### **deploy.bat (Windows)**
```batch
@echo off

REM Load production config
set SETU_DB_HOST=172.16.100.250
set SETU_DB_NAME=setu_gateway
set SETU_DB_USER=setu_gateway
set SETU_DB_PASSWORD=production_secret
set SETU_LOG_LEVEL=warn
set SETU_LOG_FORMAT=json

REM Start gateway
setu-gateway.exe
```

---

## 📊 **Config Priority**

Config values are loaded in this order (highest priority last):

1. **Default values** (lowest)
2. **YAML file** (`configs/gateway.yaml`)
3. **Custom YAML file** (`SETU_CONFIG=xxx.yaml`)
4. **Environment variables** (highest)

**Example:**
```yaml
# configs/gateway.yaml
database:
  postgres:
    host: localhost    # Will be overridden by env var
```

```bash
export SETU_DB_HOST="172.16.100.250"  # This wins!

./setu-gateway  # Uses 172.16.100.250
```

---

## 🔧 **Common Scenarios**

### **Scenario 1: Different developers, different databases**

Each developer sets their own env vars:
```bash
# Developer A
export SETU_DB_HOST="localhost"
./setu-gateway

# Developer B
export SETU_DB_HOST="192.168.1.100"
./setu-gateway
```

**Same binary, different configs!** ✅

### **Scenario 2: CI/CD Pipeline**

```bash
# In CI/CD script
export SETU_DB_HOST="${CI_DB_HOST}"
export SETU_DB_NAME="${CI_DB_NAME}"
export SETU_DB_USER="${CI_DB_USER}"
export SETU_DB_PASSWORD="${CI_DB_PASSWORD}"

./setu-gateway
```

### **Scenario 3: Quick Config Change**

```bash
# Running with localhost
./setu-gateway

# Need to switch to production?
# Ctrl+C to stop

export SETU_DB_HOST="172.16.100.250"
./setu-gateway

# No rebuild needed! ✅
```

---

## 📝 **Summary**

### **Config is NOT compiled:**
- ✅ YAML file read at runtime
- ✅ Can change config without rebuild
- ✅ Environment variables override YAML
- ✅ Multiple config files supported
- ✅ Hot-reload supported

### **Best practices:**
1. Use environment variables for sensitive data
2. Never commit passwords to Git
3. Use different config files per environment
4. Test config changes before production
5. Use secrets manager for production

### **Quick commands:**

```bash
# Change database (no rebuild!)
export SETU_DB_HOST="new-host"
export SETU_DB_PASSWORD="new-password"
./setu-gateway

# Use different config file
SETU_CONFIG=configs/production.yaml ./setu-gateway

# Check what config is loaded
# Gateway logs show database connection info on startup
```

---

**You have full flexibility to change config without recompiling!** 🎉
