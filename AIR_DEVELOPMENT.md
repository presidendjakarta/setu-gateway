# Air Live Reload - Development Guide

## 🚀 Quick Start

### **Start Development Mode**

```bash
# Option 1: Using Make (Recommended)
make dev

# Option 2: Direct air command
air

# Option 3: Debug mode (verbose logging)
air -d
# or
make dev-debug
```

---

## 📋 **What Air Does**

Air automatically:
1. ✅ Watches for file changes (`.go`, `.yaml`, `.yml`, `.json`)
2. ✅ Rebuilds the binary when code changes
3. ✅ Restarts the gateway automatically
4. ✅ Shows colored logs for build/run status

**No need to manually rebuild!** Just save and air handles the rest.

---

## 🎯 **How It Works**

```
You edit code → Air detects change → Rebuild binary → Restart gateway → Repeat!
     ↓                ↓                    ↓                 ↓
  save file      watch files         go build         run new binary
```

---

## 📁 **What Air Watches**

### **Files:**
- ✅ `*.go` - Go source files
- ✅ `*.yaml`, `*.yml` - Config files
- ✅ `*.json` - JSON files

### **Directories:**
- ✅ `./` - Root directory
- ✅ `./internal/` - Internal packages
- ✅ `./pkg/` - Public packages
- ✅ `./cmd/` - Command files
- ✅ `./configs/` - Configuration files

### **Excluded:**
- ❌ `tmp/` - Build artifacts
- ❌ `vendor/` - Dependencies
- ❌ `.git/` - Git files
- ❌ `*_test.go` - Test files

---

## 🔧 **Configuration**

Air config is in `.air.toml`:

```toml
[build]
cmd = "go build -o ./tmp/setu-gateway.exe ./cmd/gateway"
bin = "./tmp/setu-gateway.exe"
include_ext = ["go", "yaml", "yml", "json"]
delay = 1000  # Wait 1s before rebuild
```

---

## 💡 **Usage Examples**

### **Example 1: Edit Route Logic**

```bash
# Start air
make dev

# Air output:
#Watching...
# Building...
# Running...

# Edit internal/router/router.go
# Save file

# Air automatically:
# 1. Detects change
# 2. Rebuilds binary
# 3. Restarts gateway
# Done! New code is running
```

### **Example 2: Change Config**

```bash
# Air is running
make dev

# Edit configs/gateway.yaml
# Change log level to debug

logging:
  level: debug

# Save file

# Air detects YAML change and restarts gateway
# New config applied!
```

### **Example 3: Debug Mode**

```bash
# Start with verbose logging
air -d

# Shows detailed logs:
# - File change events
# - Build commands
# - Process signals
# - Error details
```

---

## 🎨 **Output Colors**

Air uses colors for different events:

- 🟣 **Magenta** - Main air messages
- 🔵 **Cyan** - File watcher events
- 🟡 **Yellow** - Build events
- 🟢 **Green** - Runner events

---

## ⚙️ **Advanced Usage**

### **Custom Config File**

```bash
# Use custom air config
air -c .air-custom.toml
```

### **Build Only (No Run)**

```bash
# Build without running
air --build-only
```

### **Generate Config**

```bash
# Generate default config
air init
```

---

## 🐛 **Troubleshooting**

### **Issue: Air not detecting changes**

**Solution 1:** Enable polling
```toml
# .air.toml
[build]
poll = true
poll_interval = 500
```

**Solution 2:** Check file exclusions
```bash
# Make sure your files aren't excluded
cat .air.toml | grep exclude
```

### **Issue: Build errors**

**Check logs:**
```bash
# Air shows build errors in yellow
# Fix the error and save again
# Air will rebuild automatically
```

### **Issue: Port already in use**

**Solution:**
```bash
# Air should stop old process automatically
# If not, manually kill:

# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8080 | xargs kill -9
```

### **Issue: Slow rebuilds**

**Optimize:**
```toml
# .air.toml
[build]
delay = 500  # Reduce delay
exclude_unchanged = true
```

---

## 📊 **Air vs Manual Development**

### **Without Air (Manual):**
```bash
# Edit code
# Save file
go build -o setu-gateway.exe ./cmd/gateway  # Manual build
./setu-gateway.exe                           # Manual run
# Test changes
# Ctrl+C to stop
# Repeat...
```

### **With Air (Automatic):**
```bash
# Start air once
make dev

# Edit code
# Save file
# Air automatically builds and restarts!
# Test changes
# Repeat... (no manual steps!)
```

**Saves 5-10 seconds per iteration!** ⚡

---

## 🎯 **Best Practices**

### **1. Keep Air Running**

```bash
# Start air in one terminal
make dev

# Use another terminal for:
# - Testing (curl)
# - Database queries
# - Git operations
```

### **2. Use Debug Logging**

```bash
# Start with debug logs
air -d

# Or set log level in config
logging:
  level: debug
```

### **3. Watch Config Files**

Air already watches `configs/` directory, so YAML changes trigger restart.

### **4. Exclude Large Directories**

```toml
# .air.toml
exclude_dir = ["tmp", "vendor", "web", "node_modules"]
```

### **5. Use Make Commands**

```bash
make dev          # Normal mode
make dev-debug    # Debug mode
make install-air  # Install air
```

---

## 📝 **Workflow Example**

### **Complete Development Session:**

```bash
# 1. Start PostgreSQL (if not running)
docker-compose up -d postgres
# or
# Start PostgreSQL service

# 2. Start air development mode
make dev

# Output:
#   __    _   ___  
#  / /\  | | | |_) 
# /_/--\ |_| |_| \_
#
# watching .
# !exclude tmp
# !exclude vendor
# building...
# running...

# 3. Edit code in your IDE
# internal/gateway/gateway.go

# 4. Save file
# Air automatically:
# !receive event WRITE
# building...
# running...

# 5. Test changes
curl http://localhost:8080/health

# 6. Edit config
# configs/gateway.yaml

# 7. Save config
# Air automatically restarts

# 8. Continue developing...

# 9. Stop air
# Ctrl+C
```

---

## 🔥 **Pro Tips**

### **Tip 1: Split Terminal**

Use split terminals:
```
Terminal 1: Air running (make dev)
Terminal 2: Testing (curl commands)
Terminal 3: Database (psql)
```

### **Tip 2: Use IDE Integration**

Most IDEs support air:
- **VS Code**: Just save files
- **GoLand**: Auto-save triggers air
- **Vim/Neovim**: `:w` triggers air

### **Tip 3: Monitor Build Time**

Air shows build time:
```
building...
running...
[debug] build took 1.234s
```

### **Tip 4: Hot-Reload Config**

Gateway has built-in config hot-reload, so YAML changes apply even faster!

---

## 📚 **Commands Reference**

### **Make Commands:**

| Command | Description |
|---------|-------------|
| `make dev` | Start air (normal mode) |
| `make dev-debug` | Start air (debug mode) |
| `make install-air` | Install air binary |
| `make build` | Build without air |
| `make run` | Build and run without air |
| `make clean` | Clean tmp files |

### **Air Commands:**

| Command | Description |
|---------|-------------|
| `air` | Start live reload |
| `air -d` | Start with debug logging |
| `air -c file.toml` | Use custom config |
| `air init` | Generate config |
| `air --build-only` | Build without run |

---

## ✅ **Checklist**

Before starting air:

- [ ] PostgreSQL is running
- [ ] Database migrations completed
- [ ] Go dependencies installed (`go mod tidy`)
- [ ] No syntax errors in code
- [ ] Port 8080 is available

---

## 🎉 **Summary**

**Air makes development faster and easier:**

✅ **Automatic rebuilds** - No manual `go build`
✅ **Automatic restarts** - No manual start/stop
✅ **Config watching** - YAML changes trigger restart
✅ **Fast iteration** - Save and test immediately
✅ **Debug mode** - Verbose logging with `-d`

**Start developing:**
```bash
make dev        # Normal mode
# or
air -d          # Debug mode
```

**Happy coding!** 🚀
