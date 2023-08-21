package litep2p

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Network interface {
	PeerHandler

	// Host
	Host() host.Host

	// Start it start the network service.
	Start() error

	// Stop it stop the network service.
	Stop() error

	// Connect connects peer by addr.
	Connect(peer.AddrInfo) error

	// Disconnect peer with id
	Disconnect(peer.ID) error

	// AsyncSend sends message to peer with peer id.
	AsyncSend(peer.ID, []byte) error

	// Send sends message to peer with peer id waiting response
	Send(peer.ID, []byte) ([]byte, error)

	// Broadcast message to all node
	Broadcast([]peer.ID, []byte) error
}

type PeerHandler interface {
	// ID get local peer id
	ID() peer.ID

	AddrInfo() peer.AddrInfo

	// PrivKey get peer private key
	PrivKey() crypto.PrivKey

	// PeerInfo get peer addr info by peer id
	PeerInfo(peer.ID) (peer.AddrInfo, error)

	// GetPeers get all network peers
	GetPeers() []peer.AddrInfo

	// LocalAddr get local peer addr
	LocalAddr() string

	// PeersNum get peers num connected
	PeersNum() int

	ConnectedPeerIds() []peer.ID

	// IsConnected check if it has an open connection to peer
	IsConnected(peer.ID) bool

	// GetRemotePubKey gets remote public key
	GetRemotePubKey(peer.ID) (crypto.PubKey, error)
}
