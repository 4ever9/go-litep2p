package litep2p

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/4ever9/go-litep2p/pb"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/multiformats/go-multiaddr"
)

type Node struct {
	config        Config
	host          host.Host
	peerAddrInfos []peer.AddrInfo
}

func (n *Node) ID() peer.ID {
	return n.host.ID()
}

func (n *Node) AddrInfo() peer.AddrInfo {
	return peer.AddrInfo{
		ID:    n.ID(),
		Addrs: n.host.Addrs(),
	}
}

func (n *Node) PrivKey() crypto.PrivKey {
	return n.config.privKey
}

func (n *Node) PeerInfo(id peer.ID) (peer.AddrInfo, error) {
	for i := 0; i < len(n.peerAddrInfos); i++ {
		if n.peerAddrInfos[i].ID == id {
			return n.peerAddrInfos[i], nil
		}
	}

	return peer.AddrInfo{}, fmt.Errorf("can not find peer info by id: %s", id)
}

func (n *Node) GetPeers() []peer.AddrInfo {
	return n.peerAddrInfos
}

func (n *Node) ConnectedPeerIds() []peer.ID {
	return n.host.Network().Peers()
}

func (n *Node) LocalAddr() string {
	return n.config.localAddr
}

func (n *Node) PeersNum() int {
	return len(n.host.Network().Peers())
}

func (n *Node) IsConnected(id peer.ID) bool {
	return n.host.Network().Connectedness(id) == network.Connected
}

func (n *Node) GetRemotePubKey(id peer.ID) (crypto.PubKey, error) {
	conns := n.host.Network().ConnsToPeer(id)

	for _, conn := range conns {
		return conn.RemotePublicKey(), nil
	}

	return nil, fmt.Errorf("not found remote pubkey by %s", id)
}

func (n *Node) Stop() error {
	return n.host.Close()
}

func (n *Node) Connect(addr peer.AddrInfo) error {
	if err := n.host.Connect(context.Background(), addr); err != nil {
		return err
	}

	n.host.Peerstore().AddAddrs(addr.ID, addr.Addrs, peerstore.PermanentAddrTTL)

	return nil
}

func (n *Node) Disconnect(id peer.ID) error {
	return n.host.Network().ClosePeer(id)
}

func (n *Node) AsyncSend(id peer.ID, data []byte) error {
	s, err := n.host.NewStream(context.Background(), id, n.config.protocolID)
	if err != nil {
		return fmt.Errorf("new stream: %w", err)
	}

	writer := pbio.NewDelimitedWriter(s)
	if err := writer.WriteMsg(&pb.Message{Data: data}); err != nil {
		return fmt.Errorf("write msg: %w", err)
	}

	return nil
}

func (n *Node) Send(id peer.ID, data []byte) ([]byte, error) {
	s, err := n.host.NewStream(context.Background(), id, n.config.protocolID)
	if err != nil {
		return nil, fmt.Errorf("new stream: %w", err)
	}

	writer := pbio.NewDelimitedWriter(s)
	if err := writer.WriteMsg(&pb.Message{Data: data}); err != nil {
		return nil, fmt.Errorf("write msg: %w", err)
	}

	msg, err := waitMsg(s, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return msg.Data, err
}

func (n *Node) Broadcast(ids []peer.ID, data []byte) error {
	errCh := make(chan error)
	for _, id := range ids {
		go func(id peer.ID) {
			if err := n.AsyncSend(id, data); err != nil {
				errCh <- fmt.Errorf("send to %s: %w", id, err)
				return
			}
			errCh <- nil
		}(id)
	}

	var err error
	for i := 0; i < len(ids); i++ {
		e := <-errCh
		if e != nil {
			err = fmt.Errorf("%w, %w", e, err)
		}
	}

	return err
}

func New(opts ...Option) (*Node, error) {
	var cfg Config
	if err := cfg.Apply(opts...); err != nil {
		return nil, fmt.Errorf("wrong config: %w", err)
	}

	host, err := libp2p.New([]libp2p.Option{
		libp2p.Identity(cfg.privKey),
		libp2p.ListenAddrStrings(cfg.localAddr),
	}...)
	if err != nil {
		return nil, fmt.Errorf("create libp2p: %w", err)
	}

	addrInfos := make([]peer.AddrInfo, 0, len(cfg.bootstrap))
	for i, addr := range cfg.bootstrap {
		maddr, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("wrong bootstrap multiaddr(%d, %s): %w", i, addr, err)
		}

		addrInfo, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			return nil, fmt.Errorf("addr info from multiaddr(%d, %v): %w", i, addr, err)
		}

		addrInfos = append(addrInfos, *addrInfo)
	}

	return &Node{
		config:        cfg,
		host:          host,
		peerAddrInfos: addrInfos,
	}, nil
}

func (n *Node) Start() error {
	if n.config.newStreamHandler != nil {
		n.host.SetStreamHandler(n.config.protocolID, n.config.newStreamHandler)
	} else {
		n.host.SetStreamHandler(n.config.protocolID, n.handleNewStream)
	}

	return nil
}

func (n *Node) handleNewStream(s network.Stream) {

}

// waitMsg wait the incoming messages within time duration.
func waitMsg(s io.Reader, timeout time.Duration) (*pb.Message, error) {
	reader := pbio.NewDelimitedReader(s, network.MessageSizeMax)
	errCh := make(chan error)
	msg := &pb.Message{}

	go func() {
		if err := reader.ReadMsg(msg); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	select {
	case r := <-errCh:
		cancel()
		if r == nil {
			return msg, nil
		}
		return nil, r
	case <-ctx.Done():
		cancel()
		return nil, fmt.Errorf("wait msg timeout")
	}
}
