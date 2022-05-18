package p2p

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/l2"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/rpc"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/sync"
	"github.com/libp2p/go-libp2p-core/connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	tswarm "github.com/libp2p/go-libp2p-swarm/testing"
	yamux "github.com/libp2p/go-libp2p-yamux"
	lconf "github.com/libp2p/go-libp2p/config"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	"github.com/stretchr/testify/require"
)

func TestingConfig(t *testing.T) *Config {
	p, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")
	mtpt, err := lconf.MuxerConstructor(yamux.DefaultTransport)
	require.NoError(t, err)
	mux := lconf.MsMuxC{MuxC: mtpt, ID: "/yamux/1.0.0"}

	return &Config{
		Priv:                (*ecdsa.PrivateKey)((p).(*crypto.Secp256k1PrivateKey)),
		DisableP2P:          false,
		NoDiscovery:         true, // we statically peer during most tests.
		ListenIP:            net.IP{127, 0, 0, 1},
		ListenTCPPort:       0, // bind to any available port
		StaticPeers:         nil,
		HostMux:             []lconf.MsMuxC{mux},
		NoTransportSecurity: true,
		PeersLo:             1,
		PeersHi:             10,
		PeersGrace:          time.Second * 10,
		NAT:                 false,
		UserAgent:           "optimism-testing",
		TimeoutNegotiation:  time.Second * 2,
		TimeoutAccept:       time.Second * 2,
		TimeoutDial:         time.Second * 2,
		Store:               sync.MutexWrap(ds.NewMapDatastore()),
		ConnGater: func(conf *Config) (connmgr.ConnectionGater, error) {
			return tswarm.DefaultMockConnectionGater(), nil
		},
		ConnMngr: DefaultConnManager,
	}
}

// Simplified p2p test, to debug/test basic libp2p things with
func TestP2PSimple(t *testing.T) {
	confA := TestingConfig(t)
	confB := TestingConfig(t)
	hostA, err := confA.Host(testlog.Logger(t, log.LvlError).New("host", "A"))
	require.NoError(t, err, "failed to launch host A")
	defer hostA.Close()
	hostB, err := confB.Host(testlog.Logger(t, log.LvlError).New("host", "B"))
	require.NoError(t, err, "failed to launch host B")
	defer hostB.Close()
	err = hostA.Connect(context.Background(), peer.AddrInfo{ID: hostB.ID(), Addrs: hostB.Addrs()})
	require.NoError(t, err, "failed to connect to peer B from peer A")
	require.Equal(t, hostB.Network().Connectedness(hostA.ID()), network.Connected)
}

type mockGossipIn struct {
	OnUnsafeL2PayloadFn func(ctx context.Context, from peer.ID, msg *l2.ExecutionPayload) error
}

func (m *mockGossipIn) OnUnsafeL2Payload(ctx context.Context, from peer.ID, msg *l2.ExecutionPayload) error {
	if m.OnUnsafeL2PayloadFn != nil {
		return m.OnUnsafeL2PayloadFn(ctx, from, msg)
	}
	return nil
}

// Full setup, using negotiated transport security and muxes
func TestP2PFull(t *testing.T) {
	pA, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")
	pB, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")

	mplexC, err := mplexC()
	require.NoError(t, err)
	yamuxC, err := yamuxC()
	require.NoError(t, err)
	noiseC, err := noiseC()
	require.NoError(t, err)
	tlsC, err := tlsC()
	require.NoError(t, err)

	confA := Config{
		Priv:                (*ecdsa.PrivateKey)((pA).(*crypto.Secp256k1PrivateKey)),
		DisableP2P:          false,
		NoDiscovery:         true,
		ListenIP:            net.IP{127, 0, 0, 1},
		ListenTCPPort:       0, // bind to any available port
		StaticPeers:         nil,
		HostMux:             []lconf.MsMuxC{yamuxC, mplexC},
		HostSecurity:        []lconf.MsSecC{noiseC, tlsC},
		NoTransportSecurity: false,
		PeersLo:             1,
		PeersHi:             10,
		PeersGrace:          time.Second * 10,
		NAT:                 false,
		UserAgent:           "optimism-testing",
		TimeoutNegotiation:  time.Second * 2,
		TimeoutAccept:       time.Second * 2,
		TimeoutDial:         time.Second * 2,
		Store:               sync.MutexWrap(ds.NewMapDatastore()),
		ConnGater:           DefaultConnGater,
		ConnMngr:            DefaultConnManager,
	}
	// copy config A, and change the settings for B
	confB := confA
	confB.Priv = (*ecdsa.PrivateKey)((pB).(*crypto.Secp256k1PrivateKey))
	confB.Store = sync.MutexWrap(ds.NewMapDatastore())
	// TODO: maybe swap the order of sec/mux preferences, to test that negotiation works

	logA := testlog.Logger(t, log.LvlError).New("host", "A")
	nodeA, err := NewNodeP2P(context.Background(), &rollup.Config{}, logA, &confA, &mockGossipIn{})
	require.NoError(t, err)
	defer nodeA.Close()

	conns := make(chan network.Conn, 1)
	hostA := nodeA.Host()
	hostA.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			conns <- conn
		}})

	backend := NewP2PAPIBackend(nodeA, logA)
	srv := rpc.NewServer()
	require.NoError(t, srv.RegisterName("opp2p", backend))
	client := rpc.DialInProc(srv)
	p2pClientA := NewClient(client)

	// Set up B to connect statically
	confB.StaticPeers, err = peer.AddrInfoToP2pAddrs(&peer.AddrInfo{ID: hostA.ID(), Addrs: hostA.Addrs()})
	require.NoError(t, err)

	logB := testlog.Logger(t, log.LvlError).New("host", "B")

	nodeB, err := NewNodeP2P(context.Background(), &rollup.Config{}, logB, &confB, &mockGossipIn{})
	require.NoError(t, err)
	defer nodeB.Close()
	hostB := nodeB.Host()

	select {
	case <-time.After(time.Second):
		t.Fatal("failed to connect new host")
	case c := <-conns:
		require.Equal(t, hostB.ID(), c.RemotePeer())
	}

	ctx := context.Background()

	selfInfoA, err := p2pClientA.Self(ctx)
	require.NoError(t, err)
	require.Equal(t, selfInfoA.PeerID, hostA.ID())

	_, err = p2pClientA.DiscoveryTable(ctx)
	// rpc does not preserve error type
	require.Equal(t, err.Error(), DisabledDiscovery.Error(), "expecting discv5 to be disabled")

	require.NoError(t, p2pClientA.BlockPeer(ctx, hostB.ID()))
	blockedPeers, err := p2pClientA.ListBlockedPeers(ctx)
	require.NoError(t, err)
	require.Equal(t, []peer.ID{hostB.ID()}, blockedPeers)
	require.NoError(t, p2pClientA.UnblockPeer(ctx, hostB.ID()))

	require.NoError(t, p2pClientA.BlockAddr(ctx, net.IP{123, 123, 123, 123}))
	blockedIPs, err := p2pClientA.ListBlockedAddrs(ctx)
	require.NoError(t, err)
	require.Len(t, blockedIPs, 1)
	require.Equal(t, net.IP{123, 123, 123, 123}, blockedIPs[0].To4())
	require.NoError(t, p2pClientA.UnblockAddr(ctx, net.IP{123, 123, 123, 123}))

	subnet := &net.IPNet{IP: net.IP{123, 0, 0, 0}.To16(), Mask: net.IPMask{0xff, 0, 0, 0}}
	require.NoError(t, p2pClientA.BlockSubnet(ctx, subnet))
	blockedSubnets, err := p2pClientA.ListBlockedSubnets(ctx)
	require.NoError(t, err)
	require.Len(t, blockedSubnets, 1)
	require.Equal(t, subnet, blockedSubnets[0])
	require.NoError(t, p2pClientA.UnblockSubnet(ctx, subnet))

	// Ask host A for all peer information they have
	peerDump, err := p2pClientA.Peers(ctx, false)
	require.Nil(t, err)
	require.Contains(t, peerDump.Peers, hostB.ID().String())
	data := peerDump.Peers[hostB.ID().String()]
	require.Equal(t, data.Direction, network.DirInbound)

	stats, err := p2pClientA.PeerStats(ctx)
	require.Nil(t, err)
	require.Equal(t, uint(1), stats.Connected)

	// disconnect
	require.NoError(t, p2pClientA.DisconnectPeer(ctx, hostB.ID()))
	peerDump, err = p2pClientA.Peers(ctx, false)
	require.Nil(t, err)
	data = peerDump.Peers[hostB.ID().String()]
	require.Equal(t, data.Connectedness, network.NotConnected)

	// reconnect
	addrsB, err := peer.AddrInfoToP2pAddrs(&peer.AddrInfo{ID: hostB.ID(), Addrs: hostB.Addrs()})
	require.NoError(t, err)
	require.NoError(t, p2pClientA.ConnectPeer(ctx, addrsB[0].String()))

	require.NoError(t, p2pClientA.ProtectPeer(ctx, hostB.ID()))
	require.NoError(t, p2pClientA.UnprotectPeer(ctx, hostB.ID()))
}

func TestDiscovery(t *testing.T) {
	pA, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")
	pB, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")
	pC, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	require.NoError(t, err, "failed to generate new p2p priv key")

	logA := testlog.Logger(t, log.LvlError).New("host", "A")
	logB := testlog.Logger(t, log.LvlError).New("host", "B")
	logC := testlog.Logger(t, log.LvlError).New("host", "C")

	mplexC, err := mplexC()
	require.NoError(t, err)
	yamuxC, err := yamuxC()
	require.NoError(t, err)
	noiseC, err := noiseC()
	require.NoError(t, err)
	tlsC, err := tlsC()
	require.NoError(t, err)

	discDBA, err := enode.OpenDB("") // "" = memory db
	require.NoError(t, err)
	discDBB, err := enode.OpenDB("")
	require.NoError(t, err)
	discDBC, err := enode.OpenDB("")
	require.NoError(t, err)

	rollupCfg := &rollup.Config{L2ChainID: big.NewInt(901)}

	confA := Config{
		Priv:                (*ecdsa.PrivateKey)((pA).(*crypto.Secp256k1PrivateKey)),
		DisableP2P:          false,
		NoDiscovery:         false,
		AdvertiseIP:         net.IP{127, 0, 0, 1},
		ListenUDPPort:       0, // bind to any available port
		ListenIP:            net.IP{127, 0, 0, 1},
		ListenTCPPort:       0, // bind to any available port
		StaticPeers:         nil,
		HostMux:             []lconf.MsMuxC{yamuxC, mplexC},
		HostSecurity:        []lconf.MsSecC{noiseC, tlsC},
		NoTransportSecurity: false,
		PeersLo:             1,
		PeersHi:             10,
		PeersGrace:          time.Second * 10,
		NAT:                 false,
		UserAgent:           "optimism-testing",
		TimeoutNegotiation:  time.Second * 2,
		TimeoutAccept:       time.Second * 2,
		TimeoutDial:         time.Second * 2,
		Store:               sync.MutexWrap(ds.NewMapDatastore()),
		DiscoveryDB:         discDBA,
		ConnGater:           DefaultConnGater,
		ConnMngr:            DefaultConnManager,
	}
	// copy config A, and change the settings for B
	confB := confA
	confB.Priv = (*ecdsa.PrivateKey)((pB).(*crypto.Secp256k1PrivateKey))
	confB.Store = sync.MutexWrap(ds.NewMapDatastore())
	confB.DiscoveryDB = discDBB

	resourcesCtx, resourcesCancel := context.WithCancel(context.Background())
	defer resourcesCancel()

	nodeA, err := NewNodeP2P(context.Background(), rollupCfg, logA, &confA, &mockGossipIn{})
	require.NoError(t, err)
	defer nodeA.Close()
	hostA := nodeA.Host()
	go nodeA.DiscoveryProcess(resourcesCtx, logA, rollupCfg, 10)

	// Add A as bootnode to B
	confB.Bootnodes = []*enode.Node{nodeA.Dv5Udp().Self()}
	// Copy B config to C, and ensure they have a different priv / peerstore
	confC := confB
	confC.Priv = (*ecdsa.PrivateKey)((pC).(*crypto.Secp256k1PrivateKey))
	confC.Store = sync.MutexWrap(ds.NewMapDatastore())
	confB.DiscoveryDB = discDBC

	// Start B
	nodeB, err := NewNodeP2P(context.Background(), rollupCfg, logB, &confB, &mockGossipIn{})
	require.NoError(t, err)
	defer nodeB.Close()
	hostB := nodeB.Host()
	go nodeB.DiscoveryProcess(resourcesCtx, logB, rollupCfg, 10)

	// Track connections to B
	connsB := make(chan network.Conn, 2)
	hostB.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, conn network.Conn) {
			log.Info("connection to B", "peer", conn.RemotePeer())
			connsB <- conn
		}})

	// Start C
	nodeC, err := NewNodeP2P(context.Background(), rollupCfg, logC, &confC, &mockGossipIn{})
	require.NoError(t, err)
	defer nodeC.Close()
	hostC := nodeC.Host()
	go nodeC.DiscoveryProcess(resourcesCtx, logC, rollupCfg, 10)

	// B and C don't know each other yet, but both have A as a bootnode.
	// It should only be a matter of time for them to connect, if they discover each other via A.
	var firstPeersOfB []peer.ID
	for i := 0; i < 2; i++ {
		select {
		case <-time.After(time.Second * 30):
			t.Fatal("failed to get connection to B in time")
		case c := <-connsB:
			firstPeersOfB = append(firstPeersOfB, c.RemotePeer())
		}
	}
	// B should be connected to the bootnode it used (it's a valid optimism node to connect to here)
	require.Contains(t, firstPeersOfB, hostA.ID())
	// C should be connected, although this one might take more time to discover
	require.Contains(t, firstPeersOfB, hostC.ID())
}

// Most tests should use mocknets instead of using the actual local host network
func TestP2PMocknet(t *testing.T) {
	mnet, err := mocknet.FullMeshConnected(3)
	require.NoError(t, err, "failed to setup mocknet")
	defer mnet.Close()
	hosts := mnet.Hosts()
	hostA, hostB, hostC := hosts[0], hosts[1], hosts[2]
	require.Equal(t, hostA.Network().Connectedness(hostB.ID()), network.Connected)
	require.Equal(t, hostA.Network().Connectedness(hostC.ID()), network.Connected)
	require.Equal(t, hostB.Network().Connectedness(hostC.ID()), network.Connected)
}
