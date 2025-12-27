# Start I2P instances and wait until they're ready for testing

$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectDir = Split-Path -Parent $scriptDir

Write-Host "=== Starting I2P Test Instances ===" -ForegroundColor Cyan
Write-Host ""

# Change to project directory
Set-Location $projectDir

# Start containers
Write-Host "Starting Docker containers..."
docker-compose up -d i2p-node1 i2p-node2

Write-Host ""
Write-Host "Containers started. Waiting for I2P routers to be ready..."
Write-Host "This can take 5-15 minutes for initial integration."
Write-Host ""

# Wait for both nodes to be ready
Write-Host "=== Checking Node 1 ===" -ForegroundColor Cyan
& "$scriptDir\check_i2p_ready.ps1" -ContainerName "i2p-node1" -MinNetDbSize 200 -TimeoutSeconds 3600

Write-Host ""
Write-Host "=== Checking Node 2 ===" -ForegroundColor Cyan
& "$scriptDir\check_i2p_ready.ps1" -ContainerName "i2p-node2" -MinNetDbSize 200 -TimeoutSeconds 3600

Write-Host ""
Write-Host "=== All I2P Instances Ready! ===" -ForegroundColor Green
Write-Host ""
Write-Host "SAM Addresses:"
$node1Port = (docker port i2p-node1 7656) -replace '.*:', ''
$node2Port = (docker port i2p-node2 7656) -replace '.*:', ''
Write-Host "  Node 1: 127.0.0.1:$node1Port"
Write-Host "  Node 2: 127.0.0.1:$node2Port"
Write-Host ""
Write-Host "You can now run tests with:"
Write-Host "  go test -v -tags=integration -run TestBidirectionalCommunication"
Write-Host ""
Write-Host "To stop instances:"
Write-Host "  docker-compose down"
Write-Host ""

