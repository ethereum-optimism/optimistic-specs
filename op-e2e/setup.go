package op_e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	bss "github.com/ethereum-optimism/optimism/op-batcher"
	"github.com/ethereum-optimism/optimism/op-bindings/deposit"
	"github.com/ethereum-optimism/optimism/op-bindings/l1block"
	"github.com/ethereum-optimism/optimism/op-bindings/l2oo"
	"github.com/ethereum-optimism/optimism/op-bindings/withdrawer"
	rollupNode "github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	"github.com/ethereum-optimism/optimism/op-node/predeploy"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	l2os "github.com/ethereum-optimism/optimism/op-proposer"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	mocknet "github.com/libp2p/go-libp2p/p2p/net/mock"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// deriveAddress returns the address associated derivation path for the wallet.
// It will panic if the derivation path is not correctly formatted.
func deriveAddress(w accounts.Wallet, path string) common.Address {
	return deriveAccount(w, path).Address
}

// deriveAccount returns the account associated derivation path for the wallet.
// It will panic if the derivation path is not correctly formatted.
func deriveAccount(w accounts.Wallet, path string) accounts.Account {
	derivPath := hdwallet.MustParseDerivationPath(path)
	account, err := w.Derive(derivPath, true)
	if err != nil {
		panic(err)
	}
	return account
}

type L2OOContractConfig struct {
	SubmissionFrequency   *big.Int
	L2StartTime           *big.Int
	L2BlockTime           *big.Int
	GenesisL2Output       [32]byte
	HistoricalTotalBlocks *big.Int
}

type DepositContractConfig struct {
	L2Oracle           common.Address
	FinalizationPeriod *big.Int
}

type SystemConfig struct {
	Mnemonic                   string
	Premine                    map[string]int // Derivation path -> amount in ETH (not wei)
	CliqueSignerDerivationPath string
	L2OutputHDPath             string
	BatchSubmitterHDPath       string
	P2PSignerHDPath            string
	DeployerHDPath             string
	L1InfoPredeployAddress     common.Address

	L2OOCfg    L2OOContractConfig
	DepositCFG DepositContractConfig

	L1WsAddr     string
	L1WsPort     int
	L1ChainID    *big.Int
	L2ChainID    *big.Int
	Nodes        map[string]*rollupNode.Config // Per node config. Don't use populate rollup.Config
	Loggers      map[string]log.Logger
	RollupConfig rollup.Config // Shared rollup configs

	L1BlockTime uint64

	// map of outbound connections to other nodes. Node names prefixed with "~" are unconnected but linked.
	// A nil map disables P2P completely.
	// Any node name not in the topology will not have p2p enabled.
	P2PTopology map[string][]string
}

type System struct {
	cfg SystemConfig

	// Retain wallet
	wallet *hdwallet.Wallet

	// Connections to running nodes
	nodes                      map[string]*node.Node
	backends                   map[string]*eth.Ethereum
	Clients                    map[string]*ethclient.Client
	RolupGenesis               rollup.Genesis
	rollupNodes                map[string]*rollupNode.OpNode
	l2OutputSubmitter          *l2os.L2OutputSubmitter
	sequencerHistoryDBFileName string
	batchSubmitter             *bss.BatchSubmitter
	L2OOContractAddr           common.Address
	DepositContractAddr        common.Address
	Mocknet                    mocknet.Mocknet
}

func precompileAlloc() core.GenesisAlloc {
	alloc := make(map[common.Address]core.GenesisAccount)
	var addr [common.AddressLength]byte
	for i := 0; i < 256; i++ {
		addr[common.AddressLength-1] = byte(i)
		alloc[addr] = core.GenesisAccount{Balance: common.Big1}
	}
	return alloc
}

func cliqueExtraData(w accounts.Wallet, signers []string) []byte {
	// 32 Empty bytes
	ret := make([]byte, 32)
	// Signer addresses
	for _, signer := range signers {
		address := deriveAddress(w, signer)
		// Was not able to automatically do this
		for i := 0; i < len(address); i++ {
			ret = append(ret, address[i])
		}
	}
	// 65 Empty bytes
	t := make([]byte, 65)
	return append(ret, t...)
}

func (sys *System) Close() {
	if sys.l2OutputSubmitter != nil {
		sys.l2OutputSubmitter.Stop()
	}
	if sys.batchSubmitter != nil {
		sys.batchSubmitter.Stop()
	}
	if sys.sequencerHistoryDBFileName != "" {
		_ = os.Remove(sys.sequencerHistoryDBFileName)
	}

	for _, node := range sys.rollupNodes {
		node.Close()
	}
	for _, node := range sys.nodes {
		node.Close()
	}
	sys.Mocknet.Close()
}

func (cfg SystemConfig) start() (*System, error) {
	sys := &System{
		cfg:         cfg,
		nodes:       make(map[string]*node.Node),
		backends:    make(map[string]*eth.Ethereum),
		Clients:     make(map[string]*ethclient.Client),
		rollupNodes: make(map[string]*rollupNode.OpNode),
	}
	didErrAfterStart := false
	defer func() {
		if didErrAfterStart {
			for _, node := range sys.rollupNodes {
				node.Close()
			}
			for _, node := range sys.nodes {
				node.Close()
			}
		}
	}()

	// Wallet
	wallet, err := hdwallet.NewFromMnemonic(cfg.Mnemonic)
	if err != nil {
		return nil, fmt.Errorf("Failed to create wallet: %w", err)
	}
	sys.wallet = wallet

	// Create the BSS and set it's config here because it needs to be derived from the accounts
	bssPrivKey, err := wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: cfg.BatchSubmitterHDPath,
		},
	})
	if err != nil {
		return nil, err
	}
	batchSubmitterAddr := crypto.PubkeyToAddress(bssPrivKey.PublicKey)

	p2pSignerPrivKey, err := wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: cfg.P2PSignerHDPath,
		},
	})
	if err != nil {
		return nil, err
	}
	p2pSignerAddr := crypto.PubkeyToAddress(p2pSignerPrivKey.PublicKey)

	// Create the L2 Outputsubmitter Address and set it here because it needs to be derived from the accounts
	l2OOSubmitter, err := wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: cfg.L2OutputHDPath,
		},
	})
	if err != nil {
		return nil, err
	}
	l2OutputSubmitterAddr := crypto.PubkeyToAddress(l2OOSubmitter.PublicKey)

	// Genesis
	eth := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	l1Alloc := precompileAlloc()
	l2Alloc := precompileAlloc()

	for path, amt := range cfg.Premine {
		balance := big.NewInt(int64(amt))
		balance.Mul(balance, eth)
		addr := deriveAddress(wallet, path)
		l1Alloc[addr] = core.GenesisAccount{Balance: balance}
		l2Alloc[addr] = core.GenesisAccount{Balance: balance}
	}

	l2Alloc[cfg.L1InfoPredeployAddress] = core.GenesisAccount{Code: common.FromHex(l1block.L1blockDeployedBin), Balance: common.Big0}
	l2Alloc[predeploy.WithdrawalContractAddress] = core.GenesisAccount{Code: common.FromHex(withdrawer.WithdrawerDeployedBin), Balance: common.Big0}

	genesisTimestamp := uint64(time.Now().Unix())

	l1Genesis := &core.Genesis{
		Config: &params.ChainConfig{
			ChainID:             cfg.L1ChainID,
			HomesteadBlock:      common.Big0,
			EIP150Block:         common.Big0,
			EIP155Block:         common.Big0,
			EIP158Block:         common.Big0,
			ByzantiumBlock:      common.Big0,
			ConstantinopleBlock: common.Big0,
			PetersburgBlock:     common.Big0,
			IstanbulBlock:       common.Big0,
			BerlinBlock:         common.Big0,
			LondonBlock:         common.Big0,
			Clique: &params.CliqueConfig{
				Period: cfg.L1BlockTime,
				Epoch:  30000,
			},
		},
		Alloc:      l1Alloc,
		Difficulty: common.Big1,
		ExtraData:  cliqueExtraData(wallet, []string{cfg.CliqueSignerDerivationPath}),
		GasLimit:   5000000,
		Nonce:      4660,
		Timestamp:  genesisTimestamp,
		BaseFee:    big.NewInt(7),
	}
	l2Genesis := &core.Genesis{
		Config: &params.ChainConfig{
			ChainID:                 cfg.L2ChainID,
			HomesteadBlock:          common.Big0,
			EIP150Block:             common.Big0,
			EIP155Block:             common.Big0,
			EIP158Block:             common.Big0,
			ByzantiumBlock:          common.Big0,
			ConstantinopleBlock:     common.Big0,
			PetersburgBlock:         common.Big0,
			IstanbulBlock:           common.Big0,
			BerlinBlock:             common.Big0,
			LondonBlock:             common.Big0,
			MergeForkBlock:          common.Big0,
			TerminalTotalDifficulty: common.Big0,
		},
		Alloc:      l2Alloc,
		Difficulty: common.Big1,
		GasLimit:   5000000,
		Nonce:      4660,
		// must be equal (or higher, while within bounds) as the L1 anchor point of the rollup
		Timestamp: genesisTimestamp,
		BaseFee:   big.NewInt(7),
	}

	// Initialize nodes
	l1Node, l1Backend, err := initL1Geth(&cfg, wallet, l1Genesis)
	if err != nil {
		return nil, err
	}
	sys.nodes["l1"] = l1Node
	sys.backends["l1"] = l1Backend
	for name, l2Cfg := range cfg.Nodes {
		node, backend, err := initL2Geth(name, l2Cfg.L2EngineAddrs[0], cfg.L2ChainID, l2Genesis)
		if err != nil {
			return nil, err
		}
		sys.nodes[name] = node
		sys.backends[name] = backend
	}

	// Start
	err = l1Node.Start()
	if err != nil {
		didErrAfterStart = true
		return nil, err
	}
	err = l1Backend.StartMining(1)
	if err != nil {
		didErrAfterStart = true
		return nil, err
	}
	for name, node := range sys.nodes {
		if name == "l1" {
			continue
		}
		err = node.Start()
		if err != nil {
			didErrAfterStart = true
			return nil, err
		}
	}

	// Geth Clients
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	l1Client, err := ethclient.DialContext(ctx, fmt.Sprintf("ws://%s:%d", cfg.L1WsAddr, cfg.L1WsPort))
	if err != nil {
		didErrAfterStart = true
		return nil, err
	}
	sys.Clients["l1"] = l1Client
	for name, node := range sys.nodes {
		client, err := ethclient.DialContext(ctx, node.WSEndpoint())
		if err != nil {
			didErrAfterStart = true
			return nil, err
		}
		sys.Clients[name] = client
	}

	// Rollup Genesis
	l1GenesisID, _ := getGenesisInfo(l1Client)
	var l2Client *ethclient.Client
	for name, client := range sys.Clients {
		if name != "l1" {
			l2Client = client
			break
		}
	}
	l2GenesisID, l2GenesisTime := getGenesisInfo(l2Client)

	sys.RolupGenesis = rollup.Genesis{
		L1:     l1GenesisID,
		L2:     l2GenesisID,
		L2Time: l2GenesisTime,
	}

	sys.cfg.RollupConfig.Genesis = sys.RolupGenesis
	sys.cfg.RollupConfig.BatchSenderAddress = batchSubmitterAddr
	sys.cfg.RollupConfig.P2PSequencerAddress = p2pSignerAddr
	sys.cfg.L2OOCfg.L2StartTime = new(big.Int).SetUint64(l2GenesisTime)

	// Deploy Deposit Contract
	deployerPrivKey, err := sys.wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: sys.cfg.DeployerHDPath,
		},
	})
	if err != nil {
		return nil, err
	}

	opts, err := bind.NewKeyedTransactorWithChainID(deployerPrivKey, cfg.L1ChainID)
	if err != nil {
		return nil, err
	}

	// Deploy contracts
	sys.L2OOContractAddr, _, _, err = l2oo.DeployL2OutputOracle(
		opts,
		l1Client,
		sys.cfg.L2OOCfg.SubmissionFrequency,
		sys.cfg.L2OOCfg.L2BlockTime,
		sys.cfg.L2OOCfg.GenesisL2Output,
		sys.cfg.L2OOCfg.HistoricalTotalBlocks,
		sys.cfg.L2OOCfg.L2StartTime,
		l2OutputSubmitterAddr,
	)
	sys.cfg.DepositCFG.L2Oracle = sys.L2OOContractAddr
	if err != nil {
		return nil, err
	}
	var tx *types.Transaction
	sys.DepositContractAddr, tx, _, err = deposit.DeployOptimismPortal(
		opts,
		l1Client,
		sys.cfg.DepositCFG.L2Oracle,
		sys.cfg.DepositCFG.FinalizationPeriod,
	)
	if err != nil {
		return nil, err
	}

	_, err = waitForTransaction(tx.Hash(), l1Client, time.Duration(cfg.L1BlockTime)*time.Second*2)
	if err != nil {
		return nil, fmt.Errorf("waiting for OptimismPortal: %w", err)
	}

	sys.Mocknet = mocknet.New()

	p2pNodes := make(map[string]*p2p.Prepared)
	if cfg.P2PTopology != nil {
		// create the peer if it doesn't exist yet.
		initHostMaybe := func(name string) (*p2p.Prepared, error) {
			if p, ok := p2pNodes[name]; ok {
				return p, nil
			}
			h, err := sys.Mocknet.GenPeer()
			if err != nil {
				return nil, fmt.Errorf("failed to init p2p host for node %s", name)
			}
			h.Network()
			_, ok := cfg.Nodes[name]
			if !ok {
				return nil, fmt.Errorf("node %s from p2p topology not found in actual nodes map", name)
			}
			// TODO we can enable discv5 in the testnodes to test discovery of new peers.
			// Would need to mock though, and the discv5 implementation does not provide nice mocks here.
			p := &p2p.Prepared{
				HostP2P:   h,
				LocalNode: nil,
				UDPv5:     nil,
			}
			p2pNodes[name] = p
			return p, nil
		}
		for k, vs := range cfg.P2PTopology {
			peerA, err := initHostMaybe(k)
			if err != nil {
				return nil, fmt.Errorf("failed to setup mocknet peer %s", k)
			}
			for _, v := range vs {
				v = strings.TrimPrefix(v, "~")
				peerB, err := initHostMaybe(v)
				if err != nil {
					return nil, fmt.Errorf("failed to setup mocknet peer %s (peer of %s)", v, k)
				}
				if _, err := sys.Mocknet.LinkPeers(peerA.HostP2P.ID(), peerB.HostP2P.ID()); err != nil {
					return nil, fmt.Errorf("failed to setup mocknet link between %s and %s", k, v)
				}
				// connect the peers after starting the full rollup node
			}
		}
	}
	// Rollup nodes
	for name, nodeConfig := range cfg.Nodes {
		c := *nodeConfig // copy
		c.Rollup = sys.cfg.RollupConfig
		c.Rollup.DepositContractAddress = sys.DepositContractAddr

		if p, ok := p2pNodes[name]; ok {
			c.P2P = p

			if c.Sequencer {
				c.P2PSigner = &p2p.PreparedSigner{Signer: p2p.NewLocalSigner(p2pSignerPrivKey)}
			}
		}

		node, err := rollupNode.New(context.Background(), &c, cfg.Loggers[name], cfg.Loggers[name], "")
		if err != nil {
			didErrAfterStart = true
			return nil, err
		}
		err = node.Start(context.Background())
		if err != nil {
			didErrAfterStart = true
			return nil, err
		}
		sys.rollupNodes[name] = node
	}

	if cfg.P2PTopology != nil {
		// We only set up the connections after starting the actual nodes,
		// so GossipSub and other p2p protocols can be started before the connections go live.
		// This way protocol negotiation happens correctly.
		for k, vs := range cfg.P2PTopology {
			peerA := p2pNodes[k]
			for _, v := range vs {
				unconnected := strings.HasPrefix(v, "~")
				if unconnected {
					v = v[1:]
				}
				if !unconnected {
					peerB := p2pNodes[v]
					if _, err := sys.Mocknet.ConnectPeers(peerA.HostP2P.ID(), peerB.HostP2P.ID()); err != nil {
						return nil, fmt.Errorf("failed to setup mocknet connection between %s and %s", k, v)
					}
				}
			}
		}
	}

	rollupEndpoint := fmt.Sprintf(
		"http://%s:%d",
		sys.cfg.Nodes["sequencer"].RPC.ListenAddr,
		sys.cfg.Nodes["sequencer"].RPC.ListenPort,
	)

	// L2Output Submitter
	sys.l2OutputSubmitter, err = l2os.NewL2OutputSubmitter(l2os.Config{
		L1EthRpc:                  "ws://127.0.0.1:9090",
		L2EthRpc:                  sys.cfg.Nodes["sequencer"].L2NodeAddr,
		RollupRpc:                 rollupEndpoint,
		L2OOAddress:               sys.L2OOContractAddr.String(),
		PollInterval:              50 * time.Millisecond,
		NumConfirmations:          1,
		ResubmissionTimeout:       3 * time.Second,
		SafeAbortNonceTooLowCount: 3,
		LogLevel:                  "info",
		LogTerminal:               true,
		Mnemonic:                  sys.cfg.Mnemonic,
		L2OutputHDPath:            sys.cfg.L2OutputHDPath,
	}, "", log.New())
	if err != nil {
		return nil, fmt.Errorf("unable to setup l2 output submitter: %w", err)
	}

	if err := sys.l2OutputSubmitter.Start(); err != nil {
		return nil, fmt.Errorf("unable to start l2 output submitter: %w", err)
	}

	sequencerHistoryDBFile, err := ioutil.TempFile("", "bss.*.json")
	if err != nil {
		return nil, fmt.Errorf("unable to create sequencer history db file: %w", err)
	}
	sys.sequencerHistoryDBFileName = sequencerHistoryDBFile.Name()
	if err = sequencerHistoryDBFile.Close(); err != nil {
		return nil, fmt.Errorf("unable to close sequencer history db file: %w", err)
	}

	// Batch Submitter
	sys.batchSubmitter, err = bss.NewBatchSubmitter(bss.Config{
		L1EthRpc:                   "ws://127.0.0.1:9090",
		L2EthRpc:                   sys.cfg.Nodes["sequencer"].L2NodeAddr,
		RollupRpc:                  rollupEndpoint,
		MinL1TxSize:                1,
		MaxL1TxSize:                120000,
		PollInterval:               50 * time.Millisecond,
		NumConfirmations:           1,
		ResubmissionTimeout:        5 * time.Second,
		SafeAbortNonceTooLowCount:  3,
		LogLevel:                   "info",
		LogTerminal:                true,
		Mnemonic:                   sys.cfg.Mnemonic,
		SequencerHDPath:            sys.cfg.BatchSubmitterHDPath,
		SequencerHistoryDBFilename: sys.sequencerHistoryDBFileName,
		SequencerGenesisHash:       sys.RolupGenesis.L2.Hash.String(),
		SequencerBatchInboxAddress: sys.cfg.RollupConfig.BatchInboxAddress.String(),
	}, "", log.New())
	if err != nil {
		return nil, fmt.Errorf("failed to setup batch submitter: %w", err)
	}

	if err := sys.batchSubmitter.Start(); err != nil {
		return nil, fmt.Errorf("unable to start batch submitter: %w", err)
	}

	return sys, nil
}
