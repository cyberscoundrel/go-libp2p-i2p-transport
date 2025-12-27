# I2P Test Instance Management Scripts

These scripts help you manage I2P instances for testing, eliminating the need to wait 45+ minutes for I2P network integration every time you run tests.

---

## 🚀 **Quick Start**

### Windows
```powershell
# Start I2P instances and wait until ready
.\scripts\start_i2p_instances.ps1

# Run tests (in another terminal)
go test -v -tags=integration -run TestBidirectionalCommunication

# Stop instances when done
docker-compose down
```

### Linux/Mac
```bash
# Start I2P instances and wait until ready
./scripts/start_i2p_instances.sh

# Run tests (in another terminal)
go test -v -tags=integration -run TestBidirectionalCommunication

# Stop instances when done
docker-compose down
```

---

## 📋 **Scripts Overview**

### `start_i2p_instances` (PowerShell/Bash)
**Purpose**: Start I2P Docker containers and wait until they're fully integrated into the I2P network.

**What it does**:
1. Starts `i2p-node1` and `i2p-node2` Docker containers
2. Waits for SAM interfaces to be accessible
3. Waits for reseed to complete
4. Waits for netDb to populate (200+ router infos)
5. Confirms routers are ready for testing

**Duration**: 5-15 minutes (first time), 2-5 minutes (subsequent runs)

**Usage**:
```powershell
# Windows
.\scripts\start_i2p_instances.ps1

# Linux/Mac
./scripts/start_i2p_instances.sh
```

---

### `check_i2p_ready` (PowerShell/Bash)
**Purpose**: Check if a specific I2P router is ready for use.

**What it checks**:
1. ✓ Container is running
2. ✓ SAM interface is accessible
3. ✓ Reseed completed successfully
4. ✓ NetDb has sufficient router infos (default: 200+)
5. ✓ Router has been running long enough
6. ✓ No critical errors in logs

**Usage**:
```powershell
# Windows
.\scripts\check_i2p_ready.ps1 -ContainerName "i2p-node1" -MinNetDbSize 200 -TimeoutSeconds 3600

# Linux/Mac
./scripts/check_i2p_ready.sh i2p-node1 200 3600
```

**Parameters**:
- `ContainerName`: Name of the Docker container (required)
- `MinNetDbSize`: Minimum number of router infos (default: 200)
- `TimeoutSeconds`: Maximum wait time in seconds (default: 3600 = 1 hour)

**Exit Codes**:
- `0`: Router is ready
- `1`: Router is not ready or error occurred

---

## 🎯 **Workflow**

### First Time Setup
```
1. Start instances:
   .\scripts\start_i2p_instances.ps1
   
   This will:
   - Start Docker containers
   - Wait for reseed (2-5 min)
   - Wait for netDb population (5-10 min)
   - Confirm readiness
   
   Total time: 5-15 minutes

2. Run tests:
   go test -v -tags=integration -run TestBidirectionalCommunication
   
   Tests run quickly (5-10 min) since instances are pre-integrated

3. Keep instances running:
   Leave containers running for subsequent test runs
   
4. Stop when done:
   docker-compose down
```

### Subsequent Test Runs
```
1. Check if instances are still running:
   docker ps | grep i2p-node
   
2. If running, just run tests:
   go test -v -tags=integration -run TestBidirectionalCommunication
   
3. If not running, restart:
   .\scripts\start_i2p_instances.ps1
   (Will be faster: 2-5 minutes)
```

---

## 📊 **Readiness Criteria**

An I2P router is considered "ready" when:

| Check | Requirement | Why |
|-------|-------------|-----|
| Container Status | Running | Basic requirement |
| SAM Interface | Accessible | Tests need SAM to work |
| Reseed | Completed | Router needs peer info |
| NetDb Size | 200+ router infos | Sufficient peers for routing |
| Uptime | 5+ minutes | Tunnels need time to build |
| Errors | < 10 critical errors | Router is healthy |

---

## 🔍 **Troubleshooting**

### Problem: "SAM interface not responding"
**Solution**: Wait longer or restart containers
```powershell
docker-compose down
docker-compose up -d
.\scripts\start_i2p_instances.ps1
```

### Problem: "Reseed not completed"
**Solution**: Check internet connection and firewall
```powershell
# Check if container can reach internet
docker exec i2p-node1 ping -c 3 8.8.8.8

# Check reseed servers
docker exec i2p-node1 curl -I https://reseed.i2p-projekt.de/
```

### Problem: "NetDb size below minimum"
**Solution**: Wait longer (can take 10-15 minutes)
```powershell
# Check current netDb size
docker exec i2p-node1 sh -c "find /i2p/.i2p/netDb -name '*.dat' | wc -l"

# Wait and check again
Start-Sleep -Seconds 60
docker exec i2p-node1 sh -c "find /i2p/.i2p/netDb -name '*.dat' | wc -l"
```

### Problem: "Container not found"
**Solution**: Make sure you're in the project directory
```powershell
cd path\to\go-libp2p-i2p-transport
.\scripts\start_i2p_instances.ps1
```

---

## 💡 **Tips**

### Keep Instances Running
Leave I2P instances running between test runs to avoid waiting for integration:
```powershell
# Start once
.\scripts\start_i2p_instances.ps1

# Run tests multiple times
go test -v -tags=integration -run TestBidirectionalCommunication
go test -v -tags=integration -run TestBidirectionalCommunication
go test -v -tags=integration -run TestBidirectionalCommunication

# Stop when done for the day
docker-compose down
```

### Check Status Anytime
```powershell
# Quick check
.\scripts\check_i2p_ready.ps1 -ContainerName "i2p-node1"

# Detailed logs
docker logs i2p-node1 --tail=50

# NetDb size
docker exec i2p-node1 sh -c "find /i2p/.i2p/netDb -name '*.dat' | wc -l"
```

### Save Integrated State (Advanced)
```powershell
# After instances are fully integrated, save their state
docker commit i2p-node1 i2p-integrated:node1
docker commit i2p-node2 i2p-integrated:node2

# Update docker-compose.yml to use integrated images
# image: i2p-integrated:node1

# Future starts will be instant!
```

---

## 📈 **Performance Comparison**

### Without Pre-Running Instances
```
Test Run 1: 60 minutes (45 min wait + 15 min test)
Test Run 2: 60 minutes (45 min wait + 15 min test)
Test Run 3: 60 minutes (45 min wait + 15 min test)
Total: 180 minutes for 3 test runs
```

### With Pre-Running Instances
```
Initial Setup: 15 minutes (one time)
Test Run 1: 10 minutes
Test Run 2: 10 minutes
Test Run 3: 10 minutes
Total: 45 minutes for 3 test runs (4x faster!)
```

---

## 🎉 **Benefits**

1. **Faster Testing**: 5-10 minutes instead of 60 minutes per test
2. **Iterative Development**: Run tests multiple times quickly
3. **Reliable**: Pre-integrated instances are more stable
4. **Flexible**: Keep instances running as long as needed
5. **Transparent**: Clear readiness criteria and status checks

---

## 📝 **Example Session**

```powershell
PS> .\scripts\start_i2p_instances.ps1
=== Starting I2P Test Instances ===

Starting Docker containers...
[+] Running 2/2
 ✔ Container i2p-node1  Started
 ✔ Container i2p-node2  Started

Containers started. Waiting for I2P routers to be ready...

=== Checking Node 1 ===
Check 1: Container status...
✓ Container is running

Check 2: SAM interface...
✓ SAM interface responding on port 7656

Check 3: Reseed status...
✓ Reseed completed
  Reseed successful, fetched 154 router infos

Check 4: NetDb size (minimum: 200)...
✓ NetDb size: 248 router infos

Check 5: Router uptime...
✓ Router uptime: 08:32

Check 6: Error check...
✓ No critical errors (3 errors found)

=== I2P Router Ready! ===
✅ Router is ready for testing!

=== Checking Node 2 ===
[... similar output ...]

=== All I2P Instances Ready! ===

SAM Addresses:
  Node 1: 127.0.0.1:7656
  Node 2: 127.0.0.1:7756

You can now run tests with:
  go test -v -tags=integration -run TestBidirectionalCommunication

PS> go test -v -tags=integration -run TestBidirectionalCommunication
=== RUN   TestBidirectionalCommunication
    integration_test.go:270: ✓ SAM interface ready at 127.0.0.1:7656
    integration_test.go:280: ✓ SAM interface ready at 127.0.0.1:7756
    [... test runs quickly ...]
--- PASS: TestBidirectionalCommunication (8.45s)
PASS
```

---

**Status**: Production Ready  
**Maintenance**: None required  
**Support**: See main README.md for issues

🎉 **Happy Testing!**

