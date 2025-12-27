package i2p

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	"github.com/joomcode/errorx"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/transport"
	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
)

type I2PTransport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader transport.Upgrader

	// Resource manager for connection scope management
	ResourceManager network.ResourceManager

	sam             *sam3.SAM
	i2PKeys         i2pkeys.I2PKeys
	primarySession  *sam3.PrimarySession
	outboundSession *sam3.StreamSession
	inboundSession  *sam3.StreamSession
	//sync.RWMutex
}

var _ transport.Transport = &I2PTransport{}

type Option func(*I2PTransport) error

type TransportBuilderFunc = func(transport.Upgrader, network.ResourceManager) (*I2PTransport, error)

// returns a function that when called by go-libp2p, creates an I2PTransport
// Initializes SAM sessions/tunnel which can take about 4-25 seconds depending
// on i2p network conditions
func I2PTransportBuilder(sam *sam3.SAM,
	i2pKeys i2pkeys.I2PKeys, outboundPort string, rngSeed int) (TransportBuilderFunc, ma.Multiaddr, error) {
	rand.Seed(int64(rngSeed))

	randSessionSuffix := strconv.Itoa(rand.Int())

	samPrimarySession, err := sam.NewPrimarySession("primarySession-"+randSessionSuffix, i2pKeys, sam3.Options_Default)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create Primary session with I2P SAM")
	}

	// Create inbound session listening on port 0 (default/any port)
	// This will accept incoming connections on the default streaming port
	inboundSession, err := samPrimarySession.NewStreamSubSession("inboundSession-" + randSessionSuffix)
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create inboundSession subsession with I2P SAM")
	}

	// Create outbound session with FROM_PORT=1 to avoid duplicate protocol/port
	// Java I2P requires unique protocol+port combinations per primary session
	// Using port 1 for outbound to differentiate from inbound's port 0
	outboundSession, err := samPrimarySession.NewStreamSubSessionWithPorts("outboundSession-"+randSessionSuffix, "1", "0")
	if err != nil {
		return nil, nil, errorx.Decorate(err, "Failed to create outbound subsession with I2P SAM")
	}

	i2pDestination, err := I2PAddrToMultiAddr(samPrimarySession.Addr().String())
	if err != nil {
		return nil, nil, err
	}

	return func(upgrader transport.Upgrader, rcmgr network.ResourceManager) (*I2PTransport, error) {
		i2p := &I2PTransport{
			Upgrader:        upgrader,
			ResourceManager: rcmgr,
			sam:             sam,
			i2PKeys:         i2pKeys,
			primarySession:  samPrimarySession,
			outboundSession: outboundSession,
			inboundSession:  inboundSession,
		}
		return i2p, nil

	}, i2pDestination, nil
}

var dialMatcher = mafmt.Or(
	mafmt.Base(ma.P_GARLIC64),
	mafmt.Base(ma.P_GARLIC32),
)

// CanDial returns true if this transport believes it can dial the given multiaddr.
func (i2p *I2PTransport) CanDial(addr ma.Multiaddr) bool {
	return dialMatcher.Matches(addr)
}

func (i2p *I2PTransport) Dial(ctx context.Context, remoteAddress ma.Multiaddr, peerID peer.ID) (transport.CapableConn, error) {
	//In case libp2p tries to dial a non-garlic address, we should error early
	if !i2p.CanDial(remoteAddress) {
		return nil, fmt.Errorf("can't dial %q: not a valid I2P address", remoteAddress)
	}

	remoteNetAddr, err := MultiAddrToI2PAddr(remoteAddress)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to convert multiaddr to I2P address")
	}

	// Check if context is already cancelled before dialing
	if ctx.Err() != nil {
		return nil, errorx.Decorate(ctx.Err(), "context cancelled before dial attempt")
	}

	// Dial with context monitoring
	conn, err := i2p.outboundSession.DialI2P(i2pkeys.I2PAddr(remoteNetAddr))
	if err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return nil, errorx.Decorate(ctx.Err(), "dial cancelled or timed out")
		}
		return nil, errorx.Decorate(err, "failed to dial I2P address %s (this may indicate I2P tunnels are not established)", remoteNetAddr)
	}

	// Check context again after dial
	if ctx.Err() != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, errorx.Decorate(ctx.Err(), "context cancelled after dial")
	}

	// Verify connection is not nil
	if conn == nil {
		return nil, fmt.Errorf("DialI2P returned nil connection without error")
	}

	localAddress, err := I2PAddrToMultiAddr(i2p.outboundSession.LocalAddr().String())
	if err != nil {
		conn.Close() // Clean up the connection
		return nil, errorx.Decorate(err, "unable to construct multi-addr from local address")
	}

	outboundConnection, err := NewConnection(conn, localAddress, remoteAddress)
	if err != nil {
		conn.Close() // Clean up the connection
		return nil, errorx.Decorate(err, "failed to construct Connection wrapper")
	}

	// Verify upgrader is not nil
	if i2p.Upgrader == nil {
		outboundConnection.Close()
		return nil, fmt.Errorf("transport upgrader is nil")
	}

	// Check context one more time before upgrade
	if ctx.Err() != nil {
		outboundConnection.Close()
		return nil, errorx.Decorate(ctx.Err(), "context cancelled before upgrade")
	}

	// Create connection scope from resource manager
	var connScope network.ConnManagementScope
	if i2p.ResourceManager != nil {
		connScope, err = i2p.ResourceManager.OpenConnection(network.DirOutbound, false, remoteAddress)
		if err != nil {
			outboundConnection.Close()
			return nil, errorx.Decorate(err, "failed to open connection scope")
		}
		defer func() {
			if err != nil {
				connScope.Done()
			}
		}()
	}

	// Upgrade the connection with proper connection scope
	upgradedConn, err := i2p.Upgrader.Upgrade(ctx, i2p, outboundConnection, network.DirOutbound, peerID, connScope)
	if err != nil {
		outboundConnection.Close()
		// Check if context was cancelled during upgrade
		if ctx.Err() != nil {
			return nil, errorx.Decorate(ctx.Err(), "connection upgrade cancelled or timed out")
		}
		return nil, errorx.Decorate(err, "failed to upgrade connection")
	}

	// Verify upgraded connection is not nil
	if upgradedConn == nil {
		if connScope != nil {
			connScope.Done()
		}
		return nil, fmt.Errorf("upgrader returned nil connection without error")
	}

	return upgradedConn, nil
}

// input argument isn't used because we'll be listening on whichever destination is provided
// by i2p
func (i2p *I2PTransport) Listen(_ ma.Multiaddr) (transport.Listener, error) {
	streamListener, err := i2p.inboundSession.Listen()
	if err != nil {
		return nil, errorx.Decorate(err, "Unable to call listen on SAM session")
	}

	listener, err := NewTransportListener(streamListener)
	if err != nil {
		return nil, errorx.Decorate(err, "Failed to initialize transport listener")
	}

	return i2p.Upgrader.UpgradeListener(i2p, listener), nil
}

// Closes all SAM sessions by closing the PRIMARY session
func (i2p *I2PTransport) Close() {
	i2p.primarySession.Close()
}

// Protocols returns the list of protocols this transport can dial.
func (i2p *I2PTransport) Protocols() []int {
	return []int{ma.P_GARLIC64, ma.P_GARLIC32, ma.P_TCP}
}

// Proxy always returns false for the I2P transport.
func (i2p *I2PTransport) Proxy() bool {
	return false
}

func (i2p *I2PTransport) String() string {
	return "I2P"
}
