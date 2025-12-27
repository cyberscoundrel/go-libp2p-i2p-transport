FROM purplei2p/i2pd:latest

# Create i2pd configuration file with SAM enabled
RUN mkdir -p /home/i2pd/data && \
    echo '[sam]' > /home/i2pd/data/i2pd.conf && \
    echo 'enabled = true' >> /home/i2pd/data/i2pd.conf && \
    echo 'address = 0.0.0.0' >> /home/i2pd/data/i2pd.conf && \
    echo 'port = 7656' >> /home/i2pd/data/i2pd.conf && \
    echo '' >> /home/i2pd/data/i2pd.conf && \
    echo '[limits]' >> /home/i2pd/data/i2pd.conf && \
    echo 'transittunnels = 2500' >> /home/i2pd/data/i2pd.conf && \
    echo '' >> /home/i2pd/data/i2pd.conf && \
    echo '[log]' >> /home/i2pd/data/i2pd.conf && \
    echo 'loglevel = info' >> /home/i2pd/data/i2pd.conf

# Use the configuration file
CMD ["/usr/local/bin/i2pd", "--conf=/home/i2pd/data/i2pd.conf", "--datadir=/home/i2pd/data"]

