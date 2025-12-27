package i2p

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/eyedeekay/sam3"
	"github.com/eyedeekay/sam3/i2pkeys"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/sec"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/upgrader"
	"github.com/libp2p/go-libp2p/p2p/security/insecure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const SAMHost = "127.0.0.1:7656"

func makeInsecureMuxer(t *testing.T) (peer.ID, sec.SecureTransport) {
	t.Helper()
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	require.NoError(t, err)

	id, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	return id, insecure.NewWithIdentity(protocol.ID(insecure.ID), id, priv)
}

type ServerInfo struct {
	Addr   i2pkeys.I2PAddr
	PeerID peer.ID
}

func TestBuildI2PTransport(t *testing.T) {
	ch := make(chan *ServerInfo, 1)
	go setupServer(t, ch)

	serverAddrAndPeer := <-ch
	setupClient(t, serverAddrAndPeer.Addr, serverAddrAndPeer.PeerID, 2345)

}

func setupClient(t *testing.T, serverAddr i2pkeys.I2PAddr, serverPeerID peer.ID, randNum int) {
	log.Println("Starting client setup")
	sam, err := sam3.NewSAM(SAMHost)
	if err != nil {
		assert.Fail(t, "Failed to connect to SAM")
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		assert.Fail(t, "Failed to generate keys")
		return
	}

	builder, _, err := I2PTransportBuilder(sam, keys, "23459", int(time.Now().UnixNano()))
	assert.NoError(t, err)

	peerID, sm := makeInsecureMuxer(t)
	log.Println("Client Peer ID is: " + peerID.String())

	// Create resource manager with infinite limits for testing
	rcmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
	require.NoError(t, err)

	// Create upgrader using new API
	upg, err := upgrader.New(
		[]sec.SecureTransport{sm},
		[]upgrader.StreamMuxer{{
			ID:    yamux.ID,
			Muxer: yamux.DefaultTransport,
		}},
		nil,   // PSK
		rcmgr, // Resource Manager
		nil,   // Connection Gater
	)
	require.NoError(t, err)

	secureTransport, err := builder(upg, rcmgr)
	require.NoError(t, err)

	serverMultiAddr, err := I2PAddrToMultiAddr(string(serverAddr))
	log.Println("Dialing host on this destination: " + serverMultiAddr.String())

	for i := 0; i < 5; i++ {
		log.Println("Starting dial")
		conn, err := secureTransport.Dial(context.TODO(), serverMultiAddr, serverPeerID)
		if err != nil {
			assert.Fail(t, "Failed to dial", err)
			return
		}
		log.Println("Opening Stream")
		stream, err := conn.OpenStream(context.TODO())
		if err != nil {
			assert.Fail(t, "Failed to open outbound stream", err)
			return
		}

		stream.Write([]byte("Hello!"))
		stream.Close()
	}
}

func setupServer(t *testing.T, addrChan chan *ServerInfo) {
	sam, err := sam3.NewSAM(SAMHost)
	if err != nil {
		assert.Fail(t, "Failed to connect to SAM", err)
		addrChan <- nil
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		assert.Fail(t, "Failed to generate keys", err)
		addrChan <- nil
		return
	}

	port := "45793"
	builder, listenAddr, err := I2PTransportBuilder(sam, keys, port, int(time.Now().UnixNano()))
	assert.NoError(t, err)

	peerID, sm := makeInsecureMuxer(t)
	log.Println("Server Peer ID is: " + peerID.String())

	// Create resource manager with infinite limits for testing
	rcmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
	require.NoError(t, err)

	// Create upgrader using new API
	upg, err := upgrader.New(
		[]sec.SecureTransport{sm},
		[]upgrader.StreamMuxer{{
			ID:    yamux.ID,
			Muxer: yamux.DefaultTransport,
		}},
		nil,   // PSK
		rcmgr, // Resource Manager
		nil,   // Connection Gater
	)
	require.NoError(t, err)

	secureTransport, err := builder(upg, rcmgr)
	require.NoError(t, err)

	listener, err := secureTransport.Listen(listenAddr)
	if err != nil {
		assert.Fail(t, "Failed to create listener", err)
		addrChan <- nil
		return
	}

	serverInfo := &ServerInfo{
		i2pkeys.I2PAddr(listener.Addr().String()),
		peerID,
	}
	addrChan <- serverInfo
	log.Println("Listener Addr: " + listener.Addr().String())

	for i := 0; i < 5; i++ {
		capableConnection, err := listener.Accept()
		if err != nil {
			assert.Fail(t, "Failed to accept connection: "+err.Error())
		}

		stream, err := capableConnection.AcceptStream()

		buf := make([]byte, 1024)
		_, err = stream.Read(buf)
		stream.Write([]byte(capableConnection.LocalMultiaddr().String()))
		log.Println(capableConnection.RemoteMultiaddr())

		stream.Close()
	}

}
