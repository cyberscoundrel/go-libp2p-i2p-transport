#!/bin/bash
# Check if I2P router is ready for use
# Returns 0 if ready, 1 if not ready

set -e

CONTAINER_NAME=$1
MIN_NETDB_SIZE=${2:-200}  # Minimum number of router infos
TIMEOUT=${3:-3600}        # Maximum wait time in seconds (default 1 hour)

if [ -z "$CONTAINER_NAME" ]; then
    echo "Usage: $0 <container-name> [min-netdb-size] [timeout-seconds]"
    echo "Example: $0 i2p-node1 200 3600"
    exit 1
fi

echo "=== Checking I2P Readiness for $CONTAINER_NAME ==="
echo ""

START_TIME=$(date +%s)

# Function to check elapsed time
check_timeout() {
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
    if [ $ELAPSED -gt $TIMEOUT ]; then
        echo "❌ Timeout after ${ELAPSED}s (max: ${TIMEOUT}s)"
        return 1
    fi
    return 0
}

# Check 1: Container is running
echo "Check 1: Container status..."
if ! docker inspect --format='{{.State.Status}}' "$CONTAINER_NAME" 2>/dev/null | grep -q "running"; then
    echo "❌ Container $CONTAINER_NAME is not running"
    exit 1
fi
echo "✓ Container is running"
echo ""

# Check 2: SAM interface is accessible
echo "Check 2: SAM interface..."
SAM_PORT=$(docker port "$CONTAINER_NAME" 7656 2>/dev/null | cut -d: -f2)
if [ -z "$SAM_PORT" ]; then
    echo "❌ SAM port not exposed"
    exit 1
fi

# Try to connect to SAM
if ! timeout 5 bash -c "echo 'HELLO VERSION' | nc localhost $SAM_PORT" >/dev/null 2>&1; then
    echo "❌ SAM interface not responding on port $SAM_PORT"
    exit 1
fi
echo "✓ SAM interface responding on port $SAM_PORT"
echo ""

# Check 3: Reseed completed
echo "Check 3: Reseed status..."
RESEED_CHECK=$(docker exec "$CONTAINER_NAME" cat /i2p/.i2p/wrapper.log 2>/dev/null | grep -i "reseed successful" | tail -1)
if [ -z "$RESEED_CHECK" ]; then
    echo "❌ Reseed not completed yet"
    echo "   Waiting for reseed to complete..."
    
    # Wait for reseed with timeout
    while check_timeout; do
        sleep 10
        RESEED_CHECK=$(docker exec "$CONTAINER_NAME" cat /i2p/.i2p/wrapper.log 2>/dev/null | grep -i "reseed successful" | tail -1)
        if [ -n "$RESEED_CHECK" ]; then
            break
        fi
        ELAPSED=$(($(date +%s) - START_TIME))
        echo "   Still waiting... (${ELAPSED}s elapsed)"
    done
    
    if [ -z "$RESEED_CHECK" ]; then
        echo "❌ Reseed did not complete in time"
        exit 1
    fi
fi
echo "✓ Reseed completed"
echo "  $RESEED_CHECK"
echo ""

# Check 4: NetDb size
echo "Check 4: NetDb size (minimum: $MIN_NETDB_SIZE)..."
NETDB_SIZE=$(docker exec "$CONTAINER_NAME" sh -c "find /i2p/.i2p/netDb -name '*.dat' 2>/dev/null | wc -l" | tr -d ' ')

if [ "$NETDB_SIZE" -lt "$MIN_NETDB_SIZE" ]; then
    echo "⚠ NetDb size: $NETDB_SIZE (below minimum: $MIN_NETDB_SIZE)"
    echo "   Waiting for netDb to populate..."
    
    # Wait for netDb to reach minimum size
    while check_timeout; do
        sleep 30
        NETDB_SIZE=$(docker exec "$CONTAINER_NAME" sh -c "find /i2p/.i2p/netDb -name '*.dat' 2>/dev/null | wc -l" | tr -d ' ')
        ELAPSED=$(($(date +%s) - START_TIME))
        echo "   NetDb size: $NETDB_SIZE / $MIN_NETDB_SIZE (${ELAPSED}s elapsed)"
        
        if [ "$NETDB_SIZE" -ge "$MIN_NETDB_SIZE" ]; then
            break
        fi
    done
    
    if [ "$NETDB_SIZE" -lt "$MIN_NETDB_SIZE" ]; then
        echo "❌ NetDb did not reach minimum size in time"
        exit 1
    fi
fi
echo "✓ NetDb size: $NETDB_SIZE router infos"
echo ""

# Check 5: Router uptime (should be at least 5 minutes for tunnels to start building)
echo "Check 5: Router uptime..."
UPTIME_CHECK=$(docker exec "$CONTAINER_NAME" sh -c "ps -o etime= -p 1" 2>/dev/null | tr -d ' ')
echo "✓ Router uptime: $UPTIME_CHECK"
echo ""

# Check 6: No critical errors in logs
echo "Check 6: Error check..."
ERROR_COUNT=$(docker exec "$CONTAINER_NAME" cat /i2p/.i2p/wrapper.log 2>/dev/null | grep -i "ERROR" | wc -l | tr -d ' ')
if [ "$ERROR_COUNT" -gt 10 ]; then
    echo "⚠ Warning: $ERROR_COUNT errors found in logs (may be normal)"
else
    echo "✓ No critical errors ($ERROR_COUNT errors found)"
fi
echo ""

# Final summary
TOTAL_ELAPSED=$(($(date +%s) - START_TIME))
echo "=== I2P Router Ready! ==="
echo "Container: $CONTAINER_NAME"
echo "SAM Port: $SAM_PORT"
echo "NetDb Size: $NETDB_SIZE router infos"
echo "Uptime: $UPTIME_CHECK"
echo "Check Duration: ${TOTAL_ELAPSED}s"
echo ""
echo "✅ Router is ready for testing!"
echo ""
echo "SAM Address: 127.0.0.1:$SAM_PORT"

exit 0

