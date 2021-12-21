package l2

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
)

var (
	DepositEventABI     = "TransactionDeposited(address,address,uint256,uint256,bool,bytes)"
	DepositEventABIHash = crypto.Keccak256Hash([]byte(DepositEventABI))
	DepositContractAddr = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	SequencerAddr       = common.HexToAddress("0x4242424242424242424242424242424242424242") // TODO pick address
	BatchSubmissionAddr = common.HexToAddress("0x0000000000000000000000006f7074696d69736d") // TODO pick address
)

// A transaction submitted to the deposit contract on L1.
type DepositTransaction struct {
	From       common.Address
	To         common.Address // 0 with IsCreation == true nil for the empty/nil address (∅)
	Value      uint256.Int
	GasLimit   uint256.Int
	IsCreation bool
	Data       []byte
}

// Parse an EVM log entry emitted by the deposit contract in order to retrieve the deposit
// transaction.
func DecodeDepositLog(log *types.Log) *DepositTransaction {
	// parse log data for:
	//     event TransactionDeposited(
	//    	 address indexed from,
	//    	 address indexed to,
	//    	 uint256 value,
	//    	 uint256 gasLimit,
	//    	 bool isCreation,
	//    	 data data
	//     );

	if common.BytesToHash(log.Topics[0][:]) != DepositEventABIHash {
		panic("Invalid deposit event selector.")
	}

	from := common.BytesToAddress(log.Topics[1][12:])
	to := common.BytesToAddress(log.Topics[2][12:])

	var value, gasLimit, dataOffset, dataLen uint256.Int
	value.SetBytes(log.Data[0:32])
	gasLimit.SetBytes(log.Data[32:64])
	isCreation := log.Data[95] != 0
	dataOffset.SetBytes(log.Data[96:128])
	dataLen.SetBytes(log.Data[128:160])
	data := log.Data[160:]

	if dataOffset[0] != 128 {
		panic("Incorrect data offset.")
	}
	if dataLen[0] != uint64(len(log.Data)-160) {
		panic("Inconsistent data length.")
	}

	return &DepositTransaction{
		From:       from,
		To:         to,
		Value:      value,
		GasLimit:   gasLimit,
		IsCreation: isCreation,
		Data:       data,
	}
}

// TODO better name & description (+ sync with spec)
type InputItem struct {
	Block    *types.Block
	Receipts []*types.Receipt
}

// Sanity check that the receipts are consistent with the block data.
func (item *InputItem) CheckReceipts() bool {
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(item.Receipts), hasher)
	return item.Block.ReceiptHash() == computed
}

// TODO better name & description (+ sync with spec)
type OutputItem struct {
	// TODO: convert to L1 attributes deposit
	L1BlockHash    common.Hash
	L1BlockNumber  uint64
	L1BlockTime    uint64
	L1Random       Bytes32
	UserDepositTxs []*DepositTransaction
}

// Transform an input item (block + receipts) into an output item (block attributes + list of
// deposited transactions).
func Derive(input InputItem) OutputItem {

	if !input.CheckReceipts() {
		panic("Receipts are not consistent with the block's receipts root.")
	}

	userDepositTxs := []*DepositTransaction{}
	for _, rec := range input.Receipts {
		if rec.Status != types.ReceiptStatusSuccessful {
			continue
		}
		for _, log := range rec.Logs {
			if log.Address == DepositContractAddr {
				userDepositTxs = append(userDepositTxs, DecodeDepositLog(log))
			}
		}
	}

	return OutputItem{
		L1BlockHash:    input.Block.Hash(),
		L1BlockNumber:  input.Block.NumberU64(),
		L1BlockTime:    input.Block.Time(),
		L1Random:       Bytes32(input.Block.MixDigest()), // TODO change to Random (needs post-merge geth)
		UserDepositTxs: userDepositTxs,
	}
}

// Interact with the execution engine in order to obtain a L2 block (represented by its hash)
// from an output item (block attributes + list of deposited transactions).
func CreateL2Block(item OutputItem) (common.Hash, error) {

	return common.Hash{}, nil // TODO
}

/*
func CreateBlock(item OutputItem) {
	// TODO pick appropriate ctx, client
	ctx := context.Background()
	var client *rpc.Client
	var logger log.Logger
	var state *ForkchoiceState

	attributes := &PayloadAttributes{
		Timestamp:             Uint64Quantity(item.L1BlockTime),
		Random:                item.L1Random,
		SuggestedFeeRecipient: SequencerAddr,
		// TODO add transactions (system (L1 attributes) + user)
	}

	fcResult, err := ForkchoiceUpdated(ctx, client, logger, state, attributes)
	if err != nil {
		return common.Hash{}, fmt.Errorf("couldn't update forkchoice: %v", err)
	} else if fcResult.Status == UpdateSyncing {
		return common.Hash{}, errors.New("ForkchoiceUpdated returned SYNCING")
	}

	payload, err := GetPayload(ctx, client, logger, fcResult.PayloadID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("couldn't get payload: %v", err)
	}

	eResult, err := ExecutePayload(ctx, client, logger, payload)
	if err != nil {
		return common.Hash{}, fmt.Errorf("couldn't execute payload: %v", err)
	}

	switch eResult.Status {
	case "SYNCING":
		return common.Hash{}, errors.New("ExecutePayload returned SYNCING")
	case "VALID":
		return common.Hash(eResult.LatestValidHash), nil
	default: // INVALID
		return common.Hash{}, errors.New("ExecutePayload returned INVALID")
	}
}
*/
