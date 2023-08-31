package litep2p

import (
	"encoding/hex"
	"fmt"
	"sync"
	"testing"

	"github.com/4ever9/go-litep2p/pb"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

var sk1Str = "08021220fedf41f90031a0566086da443d9232c70db4fdaa7fcaab0693a50fbd9bef03cc"
var sk2Str = "08021220788b4f13b27e67b0f14eee3d7e9a79a2cb74f8d1d4a0487fea1f1794e906b8b9"

func TestNode_Connect(t *testing.T) {
	n1, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30010"))
	require.Nil(t, err)
	n2, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30011"))
	require.Nil(t, err)

	err = n1.Connect(n2.AddrInfo())
	require.Nil(t, err)
}

func TestNode_AsyncSend(t *testing.T) {
	n1, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30010"))
	require.Nil(t, err)
	n2, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30011"))
	require.Nil(t, err)
	err = n2.Start()
	require.Nil(t, err)

	err = n1.Connect(n2.AddrInfo())
	require.Nil(t, err)

	err = n1.AsyncSend(n2.ID(), []byte("m1"))
	require.Nil(t, err)

	err = n1.AsyncSend(n2.ID(), []byte("m2"))
	require.Nil(t, err)

	err = n1.AsyncSend(n2.ID(), []byte("m3"))
	require.Nil(t, err)
}

func TestNode_Send(t *testing.T) {
	n1, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30010"))
	require.Nil(t, err)
	n2, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30011"), WithNewStreamHandler(func(s network.Stream) {
		writer := pbio.NewDelimitedWriter(s)
		err := writer.WriteMsg(&pb.Message{Data: []byte("litep2p")})
		if err != nil {
			slog.Error(err.Error())
		}
	}))
	require.Nil(t, err)
	err = n2.Start()
	require.Nil(t, err)

	err = n1.Connect(n2.AddrInfo())
	require.Nil(t, err)

	data, err := n1.Send(n2.ID(), []byte("hello"))
	require.Nil(t, err)

	fmt.Printf("result: %+v\n", string(data))
}

func TestNode_Broadcast(t *testing.T) {
	for i := 0; i < 10; i++ {
		var wg sync.WaitGroup
		wg.Add(2)
		n1, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30010"))
		require.Nil(t, err)
		n2, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30011"), WithNewStreamHandler(func(s network.Stream) {
			slog.Info("Node2 receive")
			wg.Done()
		}))
		require.Nil(t, err)

		n3, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30012"), WithNewStreamHandler(func(s network.Stream) {
			slog.Info("Node3 receive")
			wg.Done()
		}))

		err = n2.Start()
		require.Nil(t, err)
		err = n3.Start()
		require.Nil(t, err)

		err = n1.Connect(n2.AddrInfo())
		require.Nil(t, err)
		err = n1.Connect(n3.AddrInfo())
		require.Nil(t, err)

		err = n1.Broadcast([]peer.ID{n2.ID(), n3.ID()}, []byte("hello"))
		require.Nil(t, err)

		wg.Wait()
	}
}

func loadSK(skStr string) crypto.PrivKey {
	data, err := hex.DecodeString(skStr)
	if err != nil {
		panic(err)
	}

	sk, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		panic(err)
	}

	return sk
}

func TestNode_Broadcast2(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)
	n1, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30010"), WithNewStreamHandler(func(s network.Stream) {
		slog.Info("Node1 receive")
		wg.Done()
	}))
	require.Nil(t, err)
	n2, err := New(WithLocalAddr("/ip4/127.0.0.1/tcp/30011"), WithNewStreamHandler(func(s network.Stream) {
		slog.Info("Node2 receive")
		wg.Done()
	}))
	require.Nil(t, err)

	err = n1.Start()
	err = n2.Start()
	require.Nil(t, err)

	err = n1.Connect(n2.AddrInfo())
	require.Nil(t, err)
	err = n2.Connect(n1.AddrInfo())
	require.Nil(t, err)

	err = n1.Broadcast([]peer.ID{n2.ID()}, []byte("hello"))
	require.Nil(t, err)

	err = n2.Broadcast([]peer.ID{n1.ID()}, []byte("hi"))

	wg.Wait()
}
