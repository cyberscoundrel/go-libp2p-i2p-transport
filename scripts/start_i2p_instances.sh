#!/bin/bash
# Start I2P instances and wait until they're ready for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Starting I2P Test Instances ==="
echo ""

# Change to project directory
cd "$PROJECT_DIR"

# Start containers
echo "Starting Docker containers..."
docker-compose up -d i2p-node1 i2p-node2

echo ""
echo "Containers started. Waiting for I2P routers to be ready..."
echo "This can take 5-15 minutes for initial integration."
echo ""

# Wait for both nodes to be ready
echo "=== Checking Node 1 ==="
bash "$SCRIPT_DIR/check_i2p_ready.sh" i2p-node1 200 3600

echo ""
echo "=== Checking Node 2 ==="
bash "$SCRIPT_DIR/check_i2p_ready.sh" i2p-node2 200 3600

echo ""
echo "=== All I2P Instances Ready! ==="
echo ""
echo "SAM Addresses:"
NODE1_PORT=$(docker port i2p-node1 7656 | cut -d: -f2)
NODE2_PORT=$(docker port i2p-node2 7656 | cut -d: -f2)
echo "  Node 1: 127.0.0.1:$NODE1_PORT"
echo "  Node 2: 127.0.0.1:$NODE2_PORT"
echo ""
echo "You can now run tests with:"
echo "  go test -v -tags=integration -run TestBidirectionalCommunication"
echo ""
echo "To stop instances:"
echo "  docker-compose down"
echo ""

