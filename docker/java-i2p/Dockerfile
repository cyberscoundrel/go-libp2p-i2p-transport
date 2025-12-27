FROM geti2p/i2p:latest

# Copy SAM configuration file
COPY 00-sam.config /tmp/00-sam.config

# Create entrypoint script that fixes I2P's SAM configuration
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo 'echo "=== I2P SAM Configuration Script ==="' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Create I2P config directory' >> /entrypoint.sh && \
    echo 'mkdir -p /i2p/.i2p/clients.config.d' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Install our SAM configuration' >> /entrypoint.sh && \
    echo 'cp -f /tmp/00-sam.config /i2p/.i2p/clients.config.d/00-sam.config' >> /entrypoint.sh && \
    echo 'echo "Initial SAM configuration installed"' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Create router.config with network integration settings' >> /entrypoint.sh && \
    echo 'cat > /i2p/.i2p/router.config << "EOF"' >> /entrypoint.sh && \
    echo 'i2np.udp.port=9111' >> /entrypoint.sh && \
    echo 'i2np.tcp.port=9112' >> /entrypoint.sh && \
    echo 'i2np.udp.internalPort=9111' >> /entrypoint.sh && \
    echo 'i2np.tcp.internalPort=9112' >> /entrypoint.sh && \
    echo 'i2np.ntcp.autoip=true' >> /entrypoint.sh && \
    echo 'i2np.upnp.enable=false' >> /entrypoint.sh && \
    echo 'router.forceUnreachable=false' >> /entrypoint.sh && \
    echo 'router.dynamicKeys=false' >> /entrypoint.sh && \
    echo 'router.reseedDisable=false' >> /entrypoint.sh && \
    echo 'router.reseedSSLDisable=false' >> /entrypoint.sh && \
    echo 'router.reseedSSLRequired=true' >> /entrypoint.sh && \
    echo 'router.sharePercentage=80' >> /entrypoint.sh && \
    echo 'router.participatingTunnels=true' >> /entrypoint.sh && \
    echo 'router.buildHandlerThreads=4' >> /entrypoint.sh && \
    echo 'router.maxParticipatingTunnels=100' >> /entrypoint.sh && \
    echo 'EOF' >> /entrypoint.sh && \
    echo 'echo "Router configuration created with network integration enabled"' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Start I2P in background' >> /entrypoint.sh && \
    echo 'echo "Starting I2P router with network integration..."' >> /entrypoint.sh && \
    echo 'echo "Router will connect to I2P network and build tunnels"' >> /entrypoint.sh && \
    echo 'echo "This may take 5-15 minutes for first connection"' >> /entrypoint.sh && \
    echo '/startapp.sh &' >> /entrypoint.sh && \
    echo 'I2P_PID=$!' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Wait for I2P to create its SAM config file' >> /entrypoint.sh && \
    echo 'echo "Waiting for I2P to generate SAM config..."' >> /entrypoint.sh && \
    echo 'SAM_CONFIG="/i2p/.i2p/clients.config.d/01-net.i2p.sam.SAMBridge-clients.config"' >> /entrypoint.sh && \
    echo 'for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do' >> /entrypoint.sh && \
    echo '  if [ -f "$SAM_CONFIG" ]; then' >> /entrypoint.sh && \
    echo '    echo "SAM config file found! Fixing startOnLoad setting..."' >> /entrypoint.sh && \
    echo '    # Replace startOnLoad=false with startOnLoad=true' >> /entrypoint.sh && \
    echo '    sed -i "s/startOnLoad=false/startOnLoad=true/" "$SAM_CONFIG"' >> /entrypoint.sh && \
    echo '    # Replace delay=120 with delay=1' >> /entrypoint.sh && \
    echo '    sed -i "s/delay=120/delay=1/" "$SAM_CONFIG"' >> /entrypoint.sh && \
    echo '    # Configure SAM port if specified' >> /entrypoint.sh && \
    echo '    if [ -n "${SAM_PORT}" ]; then' >> /entrypoint.sh && \
    echo '      sed -i "s/listenPort=7656/listenPort=${SAM_PORT}/" "$SAM_CONFIG"' >> /entrypoint.sh && \
    echo '      echo "SAM port configured to ${SAM_PORT}"' >> /entrypoint.sh && \
    echo '    fi' >> /entrypoint.sh && \
    echo '    echo "SAM configuration fixed! SAM will now start."' >> /entrypoint.sh && \
    echo '    cat "$SAM_CONFIG"' >> /entrypoint.sh && \
    echo '    break' >> /entrypoint.sh && \
    echo '  fi' >> /entrypoint.sh && \
    echo '  echo "Waiting for SAM config... ($i/20)"' >> /entrypoint.sh && \
    echo '  sleep 3' >> /entrypoint.sh && \
    echo 'done' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Wait for I2P process' >> /entrypoint.sh && \
    echo 'echo "I2P is running. SAM should start shortly..."' >> /entrypoint.sh && \
    echo 'wait $I2P_PID' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

