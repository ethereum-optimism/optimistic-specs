package test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"

	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/deposit"
	"github.com/ethereum-optimism/optimistic-specs/opnode/internal/testlog"
	rollupNode "github.com/ethereum-optimism/optimistic-specs/opnode/node"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/stretchr/testify/require"
)

func getGenesisInfo(client *ethclient.Client) (id eth.BlockID, timestamp uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	block, err := client.BlockByNumber(ctx, common.Big0)
	if err != nil {
		panic(err)
	}
	return eth.BlockID{Hash: block.Hash(), Number: block.NumberU64()}, block.Time()
}

func endpoint(cfg *node.Config) string {
	return fmt.Sprintf("ws://%v", cfg.WSEndpoint())
}

// TestSystemE2E sets up a L1 Geth node, a rollup node, and a L2 geth node and then confirms that L1 deposits are reflected on L2.
// All nodes are run in process (but are the full nodes, not mocked or stubbed).
func TestSystemE2E(t *testing.T) {
	log.Root().SetHandler(log.DiscardHandler()) // Comment this out to see geth l1/l2 logs

	const l2OutputHDPath = "m/44'/60'/0'/0/3"
	const bssHDPath = "m/44'/60'/0'/0/4"

	// System Config
	cfg := &systemConfig{
		mnemonic: "squirrel green gallery layer logic title habit chase clog actress language enrich body plate fun pledge gap abuse mansion define either blast alien witness",
		l1: gethConfig{
			nodeConfig: &node.Config{
				Name:   "l1geth",
				WSHost: "127.0.0.1",
				WSPort: 9090,
			},
			ethConfig: &ethconfig.Config{
				NetworkId: 900,
			},
		},
		l2Verifier: gethConfig{
			nodeConfig: &node.Config{
				Name:   "l2gethVerify",
				WSHost: "127.0.0.1",
				WSPort: 9091,
			},
			ethConfig: &ethconfig.Config{
				NetworkId: 901,
			},
		},
		l2Sequencer: gethConfig{
			nodeConfig: &node.Config{
				Name:   "l2gethSeq",
				WSHost: "127.0.0.1",
				WSPort: 9092,
			},
			ethConfig: &ethconfig.Config{
				NetworkId: 901,
			},
		},
		premine: map[string]int{
			"m/44'/60'/0'/0/0": 10000000,
			"m/44'/60'/0'/0/1": 10000000,
			"m/44'/60'/0'/0/2": 10000000,
			l2OutputHDPath:     10000000,
			bssHDPath:          10000000,
			"m/44'/60'/0'/0/9": 10000000,
		},
		cliqueSigners:           []string{"m/44'/60'/0'/0/0"},
		depositContractAddress:  "0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001",
		l1InforPredeployAddress: "0x4242424242424242424242424242424242424242",
	}
	// Create genesis & assign it to ethconfigs
	initializeGenesis(cfg)

	// Start L1
	l1Node, l1Backend, err := l1Geth(cfg)
	require.Nil(t, err)
	defer l1Node.Close()

	err = l1Node.Start()
	require.Nil(t, err)

	err = l1Backend.StartMining(1)
	require.Nil(t, err)

	l1Client, err := ethclient.Dial(endpoint(cfg.l1.nodeConfig))
	require.Nil(t, err)
	l1GenesisID, _ := getGenesisInfo(l1Client)

	// Start L2
	l2Node, _, err := l2Geth(cfg)
	require.Nil(t, err)
	defer l2Node.Close()

	err = l2Node.Start()
	require.Nil(t, err)

	l2Client, err := ethclient.Dial(endpoint(cfg.l2Verifier.nodeConfig))
	require.Nil(t, err)
	l2GenesisID, l2GenesisTime := getGenesisInfo(l2Client)

	// Start L2
	l2SequencerNode, _, err := l2SequencerGeth(cfg)
	require.Nil(t, err)
	defer l2SequencerNode.Close()

	err = l2SequencerNode.Start()
	require.Nil(t, err)

	l2SequencerClient, err := ethclient.Dial(endpoint(cfg.l2Sequencer.nodeConfig))
	require.Nil(t, err)

	// BSS
	bssPrivKey, err := cfg.wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: bssHDPath,
		},
	})
	require.Nil(t, err)
	submitterAddress := crypto.PubkeyToAddress(bssPrivKey.PublicKey)

	// Account
	ethPrivKey, err := cfg.wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: "m/44'/60'/0'/0/9",
		},
	})
	require.Nil(t, err)

	// Verifier Rollup Node
	nodeCfg := &rollupNode.Config{
		L1NodeAddr:    endpoint(cfg.l1.nodeConfig),
		L2EngineAddrs: []string{endpoint(cfg.l2Verifier.nodeConfig)},
		Rollup: rollup.Config{
			Genesis: rollup.Genesis{
				L1:     l1GenesisID,
				L2:     l2GenesisID,
				L2Time: l2GenesisTime,
			},
			BlockTime:            1,
			MaxSequencerTimeDiff: 10,
			SeqWindowSize:        2,
			L1ChainID:            big.NewInt(900),
			// TODO pick defaults
			FeeRecipientAddress: common.Address{0xff, 0x01},
			BatchInboxAddress:   common.Address{0xff, 0x02},
			BatchSenderAddress:  submitterAddress,
		},
	}
	node, err := rollupNode.New(context.Background(), nodeCfg, testlog.Logger(t, log.LvlError))
	require.Nil(t, err)

	err = node.Start(context.Background())
	require.Nil(t, err)
	defer node.Stop()

	// Sequencer Rollup Node
	sequenceCfg := &rollupNode.Config{
		L1NodeAddr:    endpoint(cfg.l1.nodeConfig),
		L2EngineAddrs: []string{endpoint(cfg.l2Sequencer.nodeConfig)},
		Rollup: rollup.Config{
			Genesis: rollup.Genesis{
				L1:     l1GenesisID,
				L2:     l2GenesisID,
				L2Time: l2GenesisTime,
			},
			BlockTime:            1,
			MaxSequencerTimeDiff: 10,
			SeqWindowSize:        2,
			L1ChainID:            big.NewInt(900),
			// TODO pick defaults
			FeeRecipientAddress: common.Address{0xff, 0x01},
			BatchInboxAddress:   common.Address{0xff, 0x02},
			BatchSenderAddress:  submitterAddress,
		},
		Sequencer:        true,
		SubmitterPrivKey: bssPrivKey,
	}
	sequencer, err := rollupNode.New(context.Background(), sequenceCfg, testlog.Logger(t, log.LvlError))
	require.Nil(t, err)

	err = sequencer.Start(context.Background())
	require.Nil(t, err)
	defer sequencer.Stop()

	// Send Transaction & wait for success
	contractAddr := common.HexToAddress(cfg.depositContractAddress)
	fromAddr := common.HexToAddress("0x30ec912c5b1d14aa6d1cb9aa7a6682415c4f7eb0")

	// start balance
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	// Contract
	depositContract, err := deposit.NewDeposit(contractAddr, l1Client)
	require.Nil(t, err)

	// Signer
	ks := l1Node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	opts, err := bind.NewKeyStoreTransactorWithChainID(ks, ks.Accounts()[0], big.NewInt(int64(cfg.l1.ethConfig.NetworkId)))
	require.Nil(t, err)

	// Setup for L1 Confirmation
	watchChan := make(chan *deposit.DepositTransactionDeposited)
	watcher, err := depositContract.WatchTransactionDeposited(&bind.WatchOpts{}, watchChan, []common.Address{fromAddr}, []common.Address{fromAddr})
	require.Nil(t, err, "with watcher")
	defer watcher.Unsubscribe()

	// Finally send TX
	mintAmount := big.NewInt(1_000_000_000_000)
	_, err = depositContract.DepositTransaction(opts, fromAddr, mintAmount, big.NewInt(1_000_000), false, nil)
	require.Nil(t, err, "with deposit tx")

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx := types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(new(big.Int).SetUint64(cfg.l2Verifier.ethConfig.NetworkId)), &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(cfg.l2Verifier.ethConfig.NetworkId)),
		Nonce:     0,
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2SequencerClient.SendTransaction(context.Background(), tx)
	require.Nil(t, err)

	// Wait for tx to be mined on L1 (or timeout)
	select {
	case <-watchChan:
		// continue
	case err := <-watcher.Err():
		t.Fatalf("Failed on watcher channel: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for L1 tx to succeed")

	}

	// Wait (or timeout) for that block to show up on L2
	timeoutCh := time.After(6 * time.Second)
	for {
		// Confirm balance
		ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
		endBalance, err := l2Client.BalanceAt(ctx, fromAddr, nil)
		cancel()
		require.Nil(t, err)

		diff := new(big.Int)
		diff = diff.Sub(endBalance, startBalance)
		if diff.Cmp(mintAmount) == 0 {
			break loop1
		}
		select {
		case <-timeoutCh:
			t.Fatalf("Did not see balance change")
		case <-time.After(200 * time.Millisecond):
		}
	}
	// Confirm balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, diff, mintAmount, "Did not get expected balance change")

	var l2IncludedBlock *big.Int

	// Wait for tx to show up in chain (on sequencer)
	timeoutCh = time.After(6 * time.Second)
loop2:
	for {
		select {
		case <-timeoutCh:
			t.Fatal("Timeout waiting for l2 transaction")
		case <-time.After(200 * time.Millisecond):
		}

		receipt, err := l2Client.TransactionReceipt(context.Background(), tx.Hash())
		if receipt != nil && err == nil {
			l2IncludedBlock = receipt.BlockNumber
			break loop2
		} else if err != nil && !errors.Is(err, ethereum.NotFound) {
			require.Nil(t, err)
		}
	}

	verifBlock, err := l2Client.BlockByNumber(context.Background(), l2IncludedBlock)
	require.Nil(t, err)
	seqBlock, err := l2SequencerClient.BlockByNumber(context.Background(), l2IncludedBlock)
	require.Nil(t, err)
	require.Equal(t, verifBlock.Hash(), seqBlock.Hash(), "Verifier and sequencer blocks not the same after including a batch tx")

}
