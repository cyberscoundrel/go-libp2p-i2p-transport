package i2p

import (
	"fmt"
	"net"

	"github.com/joomcode/errorx"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/eyedeekay/sam3"
)

// this struct only exists to satisfy the interface requirements for libp2p connection
// upgrader
type TransportListener struct {
	streamListener *sam3.StreamListener
	multiAddr      ma.Multiaddr
}

func NewTransportListener(streamListener *sam3.StreamListener) (*TransportListener, error) {
	multiAddr, err := I2PAddrToMultiAddr(streamListener.Addr().String())
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to create MultiAddr from i2p")
	}

	return &TransportListener{
		streamListener: streamListener,
		multiAddr:      multiAddr,
	}, nil
}

func (t *TransportListener) Accept() (manet.Conn, error) {
	conn, err := t.streamListener.Accept()
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to accept connection")
	}

	// Verify connection is not nil
	if conn == nil {
		return nil, fmt.Errorf("accepted connection is nil")
	}

	localAddress, err := I2PAddrToMultiAddr(t.streamListener.Addr().String())
	if err != nil {
		conn.Close()
		return nil, errorx.Decorate(err, "Unable to construct multi-addr from local address")
	}

	remoteAddress, err := I2PAddrToMultiAddr(conn.RemoteAddr().String())
	if err != nil {
		conn.Close()
		return nil, errorx.Decorate(err, "Unable to construct multi-addr from remote address")
	}

	inboundConnection, err := NewConnection(conn, localAddress, remoteAddress)
	if err != nil {
		conn.Close()
		return nil, errorx.Decorate(err, "Failed to construct Connection type")
	}

	return inboundConnection, nil
}

func (t *TransportListener) Close() error {
	return t.streamListener.Close()
}

func (t *TransportListener) Addr() net.Addr {
	return t.streamListener.Addr()
}

func (t *TransportListener) Multiaddr() ma.Multiaddr {
	return t.multiAddr
}
