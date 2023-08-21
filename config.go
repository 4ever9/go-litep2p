package litep2p

import (
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	defaultConnectTimeout = 10 * time.Second
	defaultSendTimeout    = 2 * time.Second
	defaultWaitTimeout    = 5 * time.Second
)

type timeout struct {
	connectTimeout time.Duration
	sendTimeout    time.Duration
	waitTimeout    time.Duration
}

func defaultTimeout() *timeout {
	return &timeout{
		connectTimeout: defaultConnectTimeout,
		sendTimeout:    defaultSendTimeout,
		waitTimeout:    defaultWaitTimeout,
	}
}

type Config struct {
	localAddr        string
	privKey          crypto.PrivKey
	protocolID       protocol.ID
	bootstrap        []string
	newStreamHandler func(network.Stream)
}

type Option func(*Config)

func WithPrivateKey(privKey crypto.PrivKey) Option {
	return func(config *Config) {
		config.privKey = privKey
	}
}

func WithLocalAddr(addr string) Option {
	return func(config *Config) {
		config.localAddr = addr
	}
}

func WithProtocolID(id protocol.ID) Option {
	return func(config *Config) {
		config.protocolID = id
	}
}

func WithBootstrap(peers []string) Option {
	return func(config *Config) {
		config.bootstrap = peers
	}
}

func WithNewStreamHandler(handler func(network.Stream)) Option {
	return func(config *Config) {
		config.newStreamHandler = handler
	}
}

func checkConfig(config *Config) error {
	if config.localAddr == "" {
		return fmt.Errorf("empty local address")
	}

	return nil
}

func (cfg *Config) Apply(opts ...Option) error {
	cfg.protocolID = "/litep2p/1.0"
	for _, opt := range opts {
		opt(cfg)
	}

	if err := checkConfig(cfg); err != nil {
		return fmt.Errorf("check config: %w", err)
	}

	return nil
}
