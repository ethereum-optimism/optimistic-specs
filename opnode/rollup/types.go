package rollup

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Genesis struct {
	// The L1 block that the rollup starts *after* (no derived transactions)
	L1 eth.BlockID `json:"l1"`
	// The L2 block the rollup starts from (no transactions, pre-configured state)
	L2 eth.BlockID `json:"l2"`
	// Timestamp of L2 block
	L2Time uint64 `json:"l2_time"`
}

type Config struct {
	// Genesis anchor point of the rollup
	Genesis Genesis `json:"genesis"`
	// Seconds per L2 block
	BlockTime uint64 `json:"block_time"`
	// Sequencer batches may not be more than MaxSequencerDrift seconds after
	// the L1 timestamp of the sequencing window end.
	//
	// Note: When L1 has many 1 second consecutive blocks, and L2 grows at fixed 2 seconds,
	// the L2 time may still grow beyond this difference.
	MaxSequencerDrift uint64 `json:"max_sequencer_drift"`
	// Number of epochs (L1 blocks) per sequencing window
	SeqWindowSize uint64 `json:"seq_window_size"`
	// Required to verify L1 signatures
	L1ChainID *big.Int `json:"l1_chain_id"`
	// Required to identify the L2 network and create p2p signatures unique for this chain.
	L2ChainID *big.Int `json:"l2_chain_id"`

	// Address of the key the sequencer uses to sign blocks on the P2P layer
	P2PSequencerAddress common.Address `json:"p2p_sequencer_address"`

	// Note: below addresses are part of the block-derivation process,
	// and required to be the same network-wide to stay in consensus.

	// L2 address receiving all L2 transaction fees
	FeeRecipientAddress common.Address `json:"fee_recipient_address"`
	// L1 address that batches are sent to
	BatchInboxAddress common.Address `json:"batch_inbox_address"`
	// Acceptable batch-sender address
	BatchSenderAddress common.Address `json:"batch_sender_address"`
	// L1 Deposit Contract Address
	DepositContractAddress common.Address `json:"deposit_contract_address"`
}

// Check verifies that the given configuration makes sense
func (cfg *Config) Check() error {
	if cfg.BlockTime == 0 {
		return fmt.Errorf("block time cannot be 0, got %d", cfg.BlockTime)
	}
	if cfg.SeqWindowSize < 2 {
		return fmt.Errorf("sequencing window size must at least be 2, got %d", cfg.SeqWindowSize)
	}
	if cfg.Genesis.L1.Hash == (common.Hash{}) {
		return errors.New("genesis l1 hash cannot be empty")
	}
	if cfg.Genesis.L2.Hash == (common.Hash{}) {
		return errors.New("genesis l2 hash cannot be empty")
	}
	if cfg.Genesis.L2.Hash == cfg.Genesis.L1.Hash {
		return errors.New("achievement get! rollup inception: L1 and L2 genesis cannot be the same")
	}
	if cfg.DepositContractAddress == (common.Address{}) {
		return errors.New("did not provide deposit contract address ")
	}
	return nil
}

func (c *Config) L1Signer() types.Signer {
	return types.NewLondonSigner(c.L1ChainID)
}

type Epoch uint64
