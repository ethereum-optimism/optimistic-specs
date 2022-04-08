package test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/l2os"
	"github.com/ethereum-optimism/optimistic-specs/l2os/bindings/l2oo"
	"github.com/ethereum-optimism/optimistic-specs/l2os/rollupclient"
	"github.com/ethereum-optimism/optimistic-specs/l2os/txmgr"
	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/deposit"
	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/internal/testlog"
	rollupNode "github.com/ethereum-optimism/optimistic-specs/opnode/node"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"

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
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func waitForTransaction(hash common.Hash, client *ethclient.Client, timeout time.Duration) (*types.Receipt, error) {
	timeoutCh := time.After(timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if receipt != nil && err == nil {
			return receipt, nil
		} else if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}

		select {
		case <-timeoutCh:
			return nil, errors.New("timeout")
		case <-time.After(100 * time.Millisecond):
		}
	}
}

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
		},
		cliqueSigners:          []string{"m/44'/60'/0'/0/0"},
		depositContractAddress: derive.DepositContractAddr.Hex(),
		l1InfoPredeployAddress: derive.L1InfoPredeployAddr.Hex(),
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
			Path: "m/44'/60'/0'/0/0",
		},
	})
	require.Nil(t, err)

	// Verifier Rollup Node
	nodeCfg := &rollupNode.Config{
		L1NodeAddr:    endpoint(cfg.l1.nodeConfig),
		L2EngineAddrs: []string{endpoint(cfg.l2Verifier.nodeConfig)},
		L2NodeAddr:    endpoint(cfg.l2Verifier.nodeConfig),
		L1TrustRPC:    false, // would be faster to enable, but we want to catch if the RPC is buggy
		Rollup: rollup.Config{
			Genesis: rollup.Genesis{
				L1:     l1GenesisID,
				L2:     l2GenesisID,
				L2Time: l2GenesisTime,
			},
			BlockTime:         1,
			MaxSequencerDrift: 10,
			SeqWindowSize:     2,
			L1ChainID:         big.NewInt(900),
			// TODO pick defaults
			FeeRecipientAddress: common.Address{0xff, 0x01},
			BatchInboxAddress:   common.Address{0xff, 0x02},
			BatchSenderAddress:  submitterAddress,
		},
	}
	node, err := rollupNode.New(context.Background(), nodeCfg, testlog.Logger(t, log.LvlError), "")
	require.Nil(t, err)

	err = node.Start(context.Background())
	require.Nil(t, err)
	defer node.Stop()

	// Sequencer Rollup Node
	sequenceCfg := &rollupNode.Config{
		L1NodeAddr:    endpoint(cfg.l1.nodeConfig),
		L2EngineAddrs: []string{endpoint(cfg.l2Sequencer.nodeConfig)},
		L2NodeAddr:    endpoint(cfg.l2Verifier.nodeConfig),
		L1TrustRPC:    true, // test RPC cache usage
		Rollup: rollup.Config{
			Genesis: rollup.Genesis{
				L1:     l1GenesisID,
				L2:     l2GenesisID,
				L2Time: l2GenesisTime,
			},
			BlockTime:         1,
			MaxSequencerDrift: 10,
			SeqWindowSize:     2,
			L1ChainID:         big.NewInt(900),
			// TODO pick defaults
			FeeRecipientAddress: common.Address{0xff, 0x01},
			BatchInboxAddress:   common.Address{0xff, 0x02},
			BatchSenderAddress:  submitterAddress,
		},
		Sequencer:        true,
		SubmitterPrivKey: bssPrivKey,
		RPCListenAddr:    "127.0.0.1",
		RPCListenPort:    9093,
	}
	sequencer, err := rollupNode.New(context.Background(), sequenceCfg, testlog.Logger(t, log.LvlError), "")
	require.Nil(t, err)

	err = sequencer.Start(context.Background())
	require.Nil(t, err)
	defer sequencer.Stop()

	rollupRPCClient, err := rpc.DialContext(context.Background(), fmt.Sprintf("http://%s:%d", sequenceCfg.RPCListenAddr, sequenceCfg.RPCListenPort))
	require.Nil(t, err)
	rollupClient := rollupclient.NewRollupClient(rollupRPCClient)

	// Deploy StateRootOracle
	l2OutputPrivKey, err := cfg.wallet.PrivateKey(accounts.Account{
		URL: accounts.URL{
			Path: l2OutputHDPath,
		},
	})
	require.Nil(t, err)
	l2OutputAddr := crypto.PubkeyToAddress(l2OutputPrivKey.PublicKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	nonce, err := l1Client.NonceAt(ctx, l2OutputAddr, nil)
	require.Nil(t, err)

	opts, err := bind.NewKeyedTransactorWithChainID(
		l2OutputPrivKey, cfg.l1.ethConfig.Genesis.Config.ChainID,
	)
	require.Nil(t, err)
	opts.Nonce = big.NewInt(int64(nonce))

	submissionFrequency := big.NewInt(10) // 10 seconds
	l2BlockTime := big.NewInt(2)          // 2 seconds
	l2ooAddr, tx, l2OutputOracle, err := l2oo.DeployMockL2OutputOracle(
		opts, l1Client, submissionFrequency, l2BlockTime, [32]byte{}, big.NewInt(0),
	)
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = txmgr.WaitMined(ctx, l1Client, tx, time.Second, 1)
	require.Nil(t, err)

	initialSroTimestamp, err := l2OutputOracle.LatestBlockTimestamp(&bind.CallOpts{})
	require.Nil(t, err)

	// L2Output Submitter
	l2OutputSubmitter, err := l2os.NewL2OutputSubmitter(l2os.Config{
		L1EthRpc:                  endpoint(cfg.l1.nodeConfig),
		L2EthRpc:                  endpoint(cfg.l2Verifier.nodeConfig),
		RollupRpc:                 fmt.Sprintf("http://%s:%d", sequenceCfg.RPCListenAddr, sequenceCfg.RPCListenPort),
		L2OOAddress:               l2ooAddr.String(),
		PollInterval:              5 * time.Second,
		NumConfirmations:          1,
		ResubmissionTimeout:       5 * time.Second,
		SafeAbortNonceTooLowCount: 3,
		LogLevel:                  "error",
		Mnemonic:                  cfg.mnemonic,
		L2OutputHDPath:            l2OutputHDPath,
	}, "")
	require.Nil(t, err)

	err = l2OutputSubmitter.Start()
	require.Nil(t, err)
	defer l2OutputSubmitter.Stop()

	// Send Transaction & wait for success
	contractAddr := common.HexToAddress(cfg.depositContractAddress)
	fromAddr := common.HexToAddress("0x30ec912c5b1d14aa6d1cb9aa7a6682415c4f7eb0")

	// Contract
	depositContract, err := deposit.NewDeposit(contractAddr, l1Client)
	require.Nil(t, err)

	// Signer
	ks := l1Node.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	opts, err = bind.NewKeyStoreTransactorWithChainID(ks, ks.Accounts()[0], big.NewInt(int64(cfg.l1.ethConfig.NetworkId)))
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	startBalance, err := l2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err, "Could not get start balance")

	// Finally send TX
	mintAmount := big.NewInt(1_000_000_000_000)
	opts.Value = mintAmount
	l1DepTx, err := depositContract.DepositTransaction(opts, fromAddr, common.Big0, big.NewInt(1_000_000), false, nil)
	require.Nil(t, err, "with deposit tx")

	receipt, err := waitForTransaction(l1DepTx.Hash(), l1Client, 6*time.Second)
	require.Nil(t, err, "Waiting for tx")

	reconstructedDep, err := derive.UnmarshalLogEvent(receipt.Logs[0])
	require.NoError(t, err)
	l2DepTx := types.NewTx(reconstructedDep)
	receipt, err = waitForTransaction(l2DepTx.Hash(), l2Client, 6*time.Second)
	require.NoError(t, err)
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful)

	// Confirm balance
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	endBalance, err := l2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)

	diff := new(big.Int)
	diff = diff.Sub(endBalance, startBalance)
	require.Equal(t, diff, mintAmount, "Did not get expected balance change")

	// Wait for batch submitter to update L2 output oracle.
	timeoutCh := time.After(15 * time.Second)
	for {
		l2ooTimestamp, err := l2OutputOracle.LatestBlockTimestamp(&bind.CallOpts{})
		require.Nil(t, err)

		// Wait for the L2 output oracle to have been changed from the initial
		// timestamp set in the contract constructor.
		if l2ooTimestamp.Cmp(initialSroTimestamp) > 0 {
			// Retrieve the l2 output committed at this updated timestamp.
			committedL2Output, err := l2OutputOracle.L2Outputs(&bind.CallOpts{}, l2ooTimestamp)
			require.Nil(t, err)

			// Compute the committed L2 output's L2 block number.
			l2ooBlockNumber, err := l2OutputOracle.ComputeL2BlockNumber(
				&bind.CallOpts{}, l2ooTimestamp,
			)
			require.Nil(t, err)

			// Fetch the corresponding L2 block and assert the committed L2
			// output matches the block's state root.
			//
			// NOTE: This assertion will change once the L2 output format is
			// finalized.
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			l2Output, err := rollupClient.OutputAtBlock(ctx, l2ooBlockNumber)
			require.Nil(t, err)
			require.Len(t, l2Output, 2)

			require.Equal(t, l2Output[1][:], committedL2Output[:])
			break
		}

		select {
		case <-timeoutCh:
			t.Fatalf("State root oracle not updated")
		case <-time.After(time.Second):
		}
	}

	// Submit TX to L2 sequencer node
	toAddr := common.Address{0xff, 0xff}
	tx = types.MustSignNewTx(ethPrivKey, types.LatestSignerForChainID(new(big.Int).SetUint64(cfg.l2Verifier.ethConfig.NetworkId)), &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(cfg.l2Verifier.ethConfig.NetworkId)),
		Nonce:     1, // guess
		To:        &toAddr,
		Value:     big.NewInt(1_000_000_000),
		GasTipCap: big.NewInt(10),
		GasFeeCap: big.NewInt(200),
		Gas:       21000,
	})
	err = l2SequencerClient.SendTransaction(context.Background(), tx)
	require.Nil(t, err, "Sending L2 tx to sequencer")

	receipt, err = waitForTransaction(tx.Hash(), l2Client, 6*time.Second)
	require.Nil(t, err, "Waiting for L2 tx on verifier")

	verifBlock, err := l2Client.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	seqBlock, err := l2SequencerClient.BlockByNumber(context.Background(), receipt.BlockNumber)
	require.Nil(t, err)
	require.Equal(t, verifBlock.Hash(), seqBlock.Hash(), "Verifier and sequencer blocks not the same after including a batch tx")

}
