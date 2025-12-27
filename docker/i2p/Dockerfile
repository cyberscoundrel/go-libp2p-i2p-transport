# Dockerfile for I2P Router with SAM enabled
# Based on official I2P Java implementation

FROM eclipse-temurin:17-jre-jammy

# Install I2P
RUN apt-get update && \
    apt-get install -y wget gnupg2 lsb-release && \
    wget -q -O - https://geti2p.net/_static/i2p-debian-repo.key.asc | gpg --dearmor -o /usr/share/keyrings/i2p-archive-keyring.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/i2p-archive-keyring.gpg] https://deb.i2p.net/ jammy main" > /etc/apt/sources.list.d/i2p.list && \
    apt-get update && \
    apt-get install -y i2p && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create I2P user and directories
RUN useradd -m -s /bin/bash i2p && \
    mkdir -p /var/lib/i2p/i2p-config && \
    chown -R i2p:i2p /var/lib/i2p

# Copy I2P configuration files
COPY --chown=i2p:i2p i2ptunnel.config /var/lib/i2p/i2p-config/
COPY --chown=i2p:i2p clients.config /var/lib/i2p/i2p-config/

# Expose SAM port
EXPOSE 7656

# Expose I2P router ports
EXPOSE 7654 7655 7657 7658 4444 4445

USER i2p
WORKDIR /var/lib/i2p/i2p-config

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=120s --retries=3 \
    CMD netstat -an | grep 7656 > /dev/null || exit 1

# Start I2P router
CMD ["/usr/bin/i2prouter", "console"]

