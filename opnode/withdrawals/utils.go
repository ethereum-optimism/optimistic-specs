package withdrawals

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/l2os/bindings/l2oo"
	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/deposit"
	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/withdrawer"
	"github.com/ethereum-optimism/optimistic-specs/opnode/predeploy"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)

// WaitForFinalizationPeriod waits until the timestamp has been submitted to the L2 Output Oracle on L1 and
// then waits for the finalization period to be up.
// This functions polls and can block for a very long time if used on mainnet.
// This returns the timestamp to use for the proof generation.
func WaitForFinalizationPeriod(ctx context.Context, client *ethclient.Client, portalAddr common.Address, timestamp uint64) (uint64, error) {
	opts := &bind.CallOpts{Context: ctx}
	timestampBig := new(big.Int).SetUint64(timestamp)

	portal, err := deposit.NewOptimismPortalCaller(portalAddr, client)
	if err != nil {
		return 0, err
	}
	l2OOAddress, err := portal.L2ORACLE(opts)
	if err != nil {
		return 0, err
	}
	l2OO, err := l2oo.NewL2OutputOracleCaller(l2OOAddress, client)
	if err != nil {
		return 0, err
	}

	finalizationPeriod, err := portal.FINALIZATIONPERIOD(opts)
	if err != nil {
		return 0, err
	}

	next, err := l2OO.LatestBlockTimestamp(opts)
	if err != nil {
		return 0, err
	}

	// Now poll
	var ticker *time.Ticker
	diff := new(big.Int).Sub(timestampBig, next)
	if diff.Cmp(big.NewInt(60)) > 0 {
		ticker = time.NewTicker(time.Minute)
	} else {
		ticker = time.NewTicker(time.Second)
	}

loop:
	for {
		select {
		case <-ticker.C:
			next, err = l2OO.LatestBlockTimestamp(opts)
			if err != nil {
				return 0, err
			}
			// Already passed next
			if next.Cmp(timestampBig) > 0 {
				break loop
			}
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	// Now wait for it to be finalized
	output, err := l2OO.GetL2Output(opts, next)
	if err != nil {
		return 0, err
	}
	targetTimestamp := new(big.Int).Add(output.Timestamp, finalizationPeriod)
	targetTime := time.Unix(targetTimestamp.Int64(), 0)
	// Assume clock is relatively correct
	time.Sleep(time.Until(targetTime))
	// Poll for L1 Block to have a time greater than the target time
	ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			header, err := client.HeaderByNumber(ctx, nil)
			if err != nil {
				return 0, err
			}
			if header.Time > targetTimestamp.Uint64() {
				return next.Uint64(), nil
			}
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

}

type ProofClient interface {
	TransactionReceipt(context.Context, common.Hash) (*types.Receipt, error)
	GetProof(context.Context, common.Address, []string, *big.Int) (*gethclient.AccountResult, error)
}

type ec = *ethclient.Client
type gc = *gethclient.Client

type Client struct {
	ec
	gc
}

// Ensure that ProofClient and Client interfaces are valid
var _ ProofClient = Client{}

// NewClient wraps a RPC client with both ethclient and gethclient methods.
// Implements ProofClient
func NewClient(client *rpc.Client) *Client {
	return &Client{
		ethclient.NewClient(client),
		gethclient.New(client),
	}

}

// FinalizedWithdrawalParameters is the set of paramets to pass to the FinalizedWithdrawal function
type FinalizedWithdrawalParameters struct {
	Nonce           *big.Int
	Sender          common.Address
	Target          common.Address
	Value           *big.Int
	GasLimit        *big.Int
	Timestamp       *big.Int
	Data            []byte
	OutputRootProof deposit.WithdrawalVerifierOutputRootProof
	WithdrawalProof []byte // RLP Encoded list of trie nodes to prove L2 storage
}

// FinalizeWithdrawalParameters queries L2 to generate all withdrawal parameters and proof necessary to finalize an withdrawal on L1.
// The header provided is very imporant. It should be a block (timestamp) for which there is a submitted output in the L2 Output Oracle
// contract. If not, the withdrawal will fail as it the storage proof cannot be verified if there is no submitted state root.
func FinalizeWithdrawalParameters(ctx context.Context, l2client ProofClient, txHash common.Hash, header *types.Header) (FinalizedWithdrawalParameters, error) {
	// Transaction receipt
	receipt, err := l2client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}
	// Parse the receipt
	ev, err := ParseWithdrawalInitiated(receipt)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}
	// Generate then verify the withdrawal proof
	withdrawalHash, err := WithdrawalHash(ev)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}
	slot := StorageSlotOfWithdrawalHash(withdrawalHash)
	p, err := l2client.GetProof(ctx, predeploy.WithdrawalContractAddress, []string{slot.String()}, header.Number)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}
	// TODO: Could skip this step, but it's nice to double check it
	err = VerifyProof(header.Root, p)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}
	if len(p.StorageProof) != 1 {
		return FinalizedWithdrawalParameters{}, errors.New("invalid amount of storage proofs")
	}

	// Encode it as expected by the contract
	trieNodes := make([][]byte, len(p.StorageProof[0].Proof))
	for i, s := range p.StorageProof[0].Proof {
		trieNodes[i] = common.FromHex(s)
	}

	withdrawalProof, err := rlp.EncodeToBytes(trieNodes)
	if err != nil {
		return FinalizedWithdrawalParameters{}, err
	}

	return FinalizedWithdrawalParameters{
		Nonce:     ev.Nonce,
		Sender:    ev.Sender,
		Target:    ev.Target,
		Value:     ev.Value,
		GasLimit:  ev.GasLimit,
		Timestamp: new(big.Int).SetUint64(header.Time),
		Data:      ev.Data,
		OutputRootProof: deposit.WithdrawalVerifierOutputRootProof{
			Version:               [32]byte{}, // Empty for version 1
			StateRoot:             header.Root,
			WithdrawerStorageRoot: p.StorageHash,
			LatestBlockhash:       header.Hash(),
		},
		WithdrawalProof: withdrawalProof,
	}, nil
}

// Standard ABI types copied from golang ABI tests
var (
	Uint256Type, _ = abi.NewType("uint256", "", nil)
	BytesType, _   = abi.NewType("bytes", "", nil)
	AddressType, _ = abi.NewType("address", "", nil)
)

// WithdrawalHash computes the hash of the withdrawal that was stored in the L2 withdrawal contract state.
// TODO:
// 	- I don't like having to use the ABI Generated struct
// 	- There should be a better way to run the ABI encoding
//	- These needs to be fuzzed against the solidity
func WithdrawalHash(ev *withdrawer.WithdrawerWithdrawalInitiated) (common.Hash, error) {
	//  abi.encode(nonce, msg.sender, _target, msg.value, _gasLimit, _data)
	args := abi.Arguments{
		{Name: "nonce", Type: Uint256Type},
		{Name: "sender", Type: AddressType},
		{Name: "target", Type: AddressType},
		{Name: "value", Type: Uint256Type},
		{Name: "gasLimit", Type: Uint256Type},
		{Name: "data", Type: BytesType},
	}
	enc, err := args.Pack(ev.Nonce, ev.Sender, ev.Target, ev.Value, ev.GasLimit, ev.Data)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack for withdrawal hash: %w", err)
	}
	return crypto.Keccak256Hash(enc), nil
}

// ParseWithdrawalInitiated parses
func ParseWithdrawalInitiated(receipt *types.Receipt) (*withdrawer.WithdrawerWithdrawalInitiated, error) {
	contract, err := withdrawer.NewWithdrawer(common.Address{}, nil)
	if err != nil {
		return nil, err
	}
	if len(receipt.Logs) != 1 {
		return nil, errors.New("invalid length of logs")
	}
	ev, err := contract.ParseWithdrawalInitiated(*receipt.Logs[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse log: %w", err)
	}
	return ev, nil
}

// StorageSlotOfWithdrawalHash determines the storage slot of the Withdrawer contract to look at
// given a WithdrawalHash
func StorageSlotOfWithdrawalHash(hash common.Hash) common.Hash {
	// The withdrawals mapping is the second (0 indexed) storage element in the Withdrawer contract.
	// To determine the storage slot, use keccak256(withdrawalHash ++ p)
	// Where p is the 32 byte value of the storage slot and ++ is concatenation
	buf := make([]byte, 64)
	copy(buf, hash[:])
	buf[63] = 1
	return crypto.Keccak256Hash(buf)
}
