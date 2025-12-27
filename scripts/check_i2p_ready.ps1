# Check if I2P router is ready for use
# Returns 0 if ready, 1 if not ready

param(
    [Parameter(Mandatory=$true)]
    [string]$ContainerName,

    [int]$MinNetDbSize = 200,

    [int]$TimeoutSeconds = 3600
)

$ErrorActionPreference = "Stop"

Write-Host "=== Checking I2P Readiness for $ContainerName ===" -ForegroundColor Cyan
Write-Host ""

$startTime = Get-Date

function Check-Timeout {
    $elapsed = (Get-Date) - $startTime
    if ($elapsed.TotalSeconds -gt $TimeoutSeconds) {
        Write-Host "[X] Timeout after $([int]$elapsed.TotalSeconds)s (max: ${TimeoutSeconds}s)" -ForegroundColor Red
        return $false
    }
    return $true
}

# Check 1: Container is running
Write-Host "Check 1: Container status..."
try {
    $status = docker inspect --format='{{.State.Status}}' $ContainerName 2>$null
    if ($status -ne "running") {
        Write-Host "[X] Container $ContainerName is not running" -ForegroundColor Red
        exit 1
    }
    Write-Host "[OK] Container is running" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "[X] Container $ContainerName not found" -ForegroundColor Red
    exit 1
}

# Check 2: SAM interface is accessible
Write-Host "Check 2: SAM interface..."
$samPort = (docker port $ContainerName 7656 2>$null) -replace '.*:', ''
if (-not $samPort) {
    Write-Host "[X] SAM port not exposed" -ForegroundColor Red
    exit 1
}

# Try to connect to SAM
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $tcpClient.Connect("127.0.0.1", $samPort)
    $tcpClient.Close()
    Write-Host "[OK] SAM interface responding on port $samPort" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "[X] SAM interface not responding on port $samPort" -ForegroundColor Red
    exit 1
}

# Check 3: Reseed completed
Write-Host "Check 3: Reseed status..."
$reseedCheck = docker exec $ContainerName cat /i2p/.i2p/wrapper.log 2>$null | Select-String -Pattern "reseed successful" | Select-Object -Last 1

if (-not $reseedCheck) {
    Write-Host "[!] Reseed not completed yet" -ForegroundColor Yellow
    Write-Host "   Waiting for reseed to complete..."

    while (Check-Timeout) {
        Start-Sleep -Seconds 10
        $reseedCheck = docker exec $ContainerName cat /i2p/.i2p/wrapper.log 2>$null | Select-String -Pattern "reseed successful" | Select-Object -Last 1
        if ($reseedCheck) {
            break
        }
        $elapsed = (Get-Date) - $startTime
        $elapsedSeconds = [int]$elapsed.TotalSeconds
        Write-Host "   Still waiting... ($elapsedSeconds seconds elapsed)"
    }

    if (-not $reseedCheck) {
        Write-Host "[X] Reseed did not complete in time" -ForegroundColor Red
        exit 1
    }
}
Write-Host "[OK] Reseed completed" -ForegroundColor Green
Write-Host "  $reseedCheck"
Write-Host ""

# Check 4: NetDb size
Write-Host "Check 4: NetDb size (minimum: $MinNetDbSize)..."
$netDbSize = [int](docker exec $ContainerName sh -c "find /i2p/.i2p/netDb -name '*.dat' 2>/dev/null | wc -l").Trim()

if ($netDbSize -lt $MinNetDbSize) {
    Write-Host "[!] NetDb size: $netDbSize (below minimum: $MinNetDbSize)" -ForegroundColor Yellow
    Write-Host "   Waiting for netDb to populate..."

    while (Check-Timeout) {
        Start-Sleep -Seconds 30
        $netDbSize = [int](docker exec $ContainerName sh -c "find /i2p/.i2p/netDb -name '*.dat' 2>/dev/null | wc -l").Trim()
        $elapsed = (Get-Date) - $startTime
        $elapsedSeconds = [int]$elapsed.TotalSeconds
        Write-Host "   NetDb size: $netDbSize / $MinNetDbSize ($elapsedSeconds seconds elapsed)"

        if ($netDbSize -ge $MinNetDbSize) {
            break
        }
    }

    if ($netDbSize -lt $MinNetDbSize) {
        Write-Host "[X] NetDb did not reach minimum size in time" -ForegroundColor Red
        exit 1
    }
}
Write-Host "[OK] NetDb size: $netDbSize router infos" -ForegroundColor Green
Write-Host ""

# Check 5: Router uptime
Write-Host "Check 5: Router uptime..."
$uptime = (docker exec $ContainerName sh -c "ps -o etime= -p 1" 2>$null).Trim()
Write-Host "[OK] Router uptime: $uptime" -ForegroundColor Green
Write-Host ""

# Check 6: No critical errors in logs
Write-Host "Check 6: Error check..."
$errorCount = (docker exec $ContainerName cat /i2p/.i2p/wrapper.log 2>$null | Select-String -Pattern "ERROR" | Measure-Object).Count
if ($errorCount -gt 10) {
    Write-Host "[!] Warning: $errorCount errors found in logs (may be normal)" -ForegroundColor Yellow
} else {
    Write-Host "[OK] No critical errors ($errorCount errors found)" -ForegroundColor Green
}
Write-Host ""

# Final summary
$totalElapsed = (Get-Date) - $startTime
Write-Host "=== I2P Router Ready! ===" -ForegroundColor Green
Write-Host "Container: $ContainerName"
Write-Host "SAM Port: $samPort"
Write-Host "NetDb Size: $netDbSize router infos"
Write-Host "Uptime: $uptime"
Write-Host "Check Duration: $([int]$totalElapsed.TotalSeconds)s"
Write-Host ""
Write-Host "[OK] Router is ready for testing!" -ForegroundColor Green
Write-Host ""
Write-Host "SAM Address: 127.0.0.1:$samPort"

exit 0
