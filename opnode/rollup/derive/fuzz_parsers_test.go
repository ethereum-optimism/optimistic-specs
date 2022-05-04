package derive

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/deposit"
	"github.com/ethereum-optimism/optimistic-specs/opnode/contracts/l1block"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/go-cmp/cmp"
)

var (
	pk, _                  = crypto.GenerateKey()
	addr                   = common.Address{0x42, 0xff}
	opts, _                = bind.NewKeyedTransactorWithChainID(pk, common.Big1)
	from                   = crypto.PubkeyToAddress(pk.PublicKey)
	portalContract, _      = deposit.NewOptimismPortal(addr, nil)
	l1BlockInfoContract, _ = l1block.NewL1Block(addr, nil)
)

func cap_byte_slice(b []byte, c int) []byte {
	if len(b) <= c {
		return b
	} else {
		return b[:c]
	}
}

func BytesToBigInt(b []byte) *big.Int {
	return new(big.Int).SetBytes(cap_byte_slice(b, 32))
}

func BigEqual(a, b *big.Int) bool {
	if a == nil || b == nil {
		return a == b
	} else {
		return a.Cmp(b) == 0
	}
}

// FuzzL1InfoRoundTrip checks that our encoder round trips properly
func FuzzL1InfoRoundTrip(f *testing.F) {
	f.Fuzz(func(t *testing.T, number, time uint64, baseFee, hash []byte, seqNumber uint64) {
		in := L1BlockInfo{
			Number:         number,
			Time:           time,
			BaseFee:        BytesToBigInt(baseFee),
			BlockHash:      common.BytesToHash(hash),
			SequenceNumber: seqNumber,
		}
		enc, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("Failed to marshal binary: %v", err)
		}
		var out L1BlockInfo
		err = out.UnmarshalBinary(enc)
		if err != nil {
			t.Fatalf("Failed to unmarshal binary: %v", err)
		}
		if !cmp.Equal(in, out, cmp.Comparer(BigEqual)) {
			t.Fatalf("The data did not round trip correctly. in: %v. out: %v", in, out)
		}

	})
}

// FuzzL1InfoAgainstContract checks the custom marshalling functions against the contract
// bindings to ensure that our functions are up to date and match the bindings.
func FuzzL1InfoAgainstContract(f *testing.F) {
	f.Fuzz(func(t *testing.T, number, time uint64, baseFee, hash []byte, seqNumber uint64) {
		expected := L1BlockInfo{
			Number:         number,
			Time:           time,
			BaseFee:        BytesToBigInt(baseFee),
			BlockHash:      common.BytesToHash(hash),
			SequenceNumber: seqNumber,
		}

		// Setup opts
		opts.GasPrice = big.NewInt(100)
		opts.GasLimit = 100_000
		opts.NoSend = true
		opts.Nonce = common.Big0
		// Create the SetL1BlockValues transaction
		tx, err := l1BlockInfoContract.SetL1BlockValues(
			opts,
			number,
			time,
			BytesToBigInt(baseFee),
			common.BytesToHash(hash),
			seqNumber,
		)
		if err != nil {
			t.Fatalf("Failed to create the transaction: %v", err)
		}

		// Check that our encoder produces the same value and that we
		// can decode the contract values exactly
		enc, err := expected.MarshalBinary()
		if err != nil {
			t.Fatalf("Failed to marshal binary: %v", err)
		}
		if !bytes.Equal(enc, tx.Data()) {
			t.Fatalf("Custom marshal does not match contract bindings")
		}

		var actual L1BlockInfo
		err = actual.UnmarshalBinary(tx.Data())
		if err != nil {
			t.Fatalf("Failed to unmarshal binary: %v", err)
		}

		if !cmp.Equal(expected, actual, cmp.Comparer(BigEqual)) {
			t.Fatalf("The data did not round trip correctly. expected: %v. actual: %v", expected, actual)
		}

	})
}

// FuzzUnmarshallLogEvent runs a deposit event through the EVM and checks that output of the abigen parsing matches
// what was inputted and what we parsed during the UnmarshalLogEvent function (which turns it into a deposit tx)
// The purpose is to check that we can never create a transaction that emits a log that we cannot parse as well
// as ensuring that our custom marshalling matches abigen.
func FuzzUnmarshallLogEvent(f *testing.F) {
	b := func(i int64) []byte {
		return big.NewInt(i).Bytes()
	}
	type setup struct {
		to         common.Address
		mint       int64
		value      int64
		gasLimit   uint64
		data       string
		isCreation bool
	}
	cases := []setup{
		{
			mint:     100,
			value:    50,
			gasLimit: 100000,
		},
	}
	for _, c := range cases {
		f.Add(c.to.Bytes(), b(c.mint), b(c.value), []byte(c.data), c.gasLimit, c.isCreation)
	}

	f.Fuzz(func(t *testing.T, _to, _mint, _value, data []byte, l2GasLimit uint64, isCreation bool) {
		to := common.BytesToAddress(_to)
		mint := BytesToBigInt(_mint)
		value := BytesToBigInt(_value)

		// Setup opts
		opts.Value = mint
		opts.GasPrice = common.Big2
		opts.GasLimit = 500_000
		opts.NoSend = true
		opts.Nonce = common.Big0
		// Create the deposit transaction
		tx, err := portalContract.DepositTransaction(opts, to, value, l2GasLimit, isCreation, data)
		if err != nil {
			t.Fatal(err)
		}

		state, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		if err != nil {
			t.Fatal(err)
		}
		state.SetBalance(from, BytesToBigInt([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}))
		state.SetCode(addr, common.FromHex(deposit.OptimismPortalDeployedBin))
		_, err = state.Commit(false)
		if err != nil {
			t.Fatal(err)
		}

		cfg := runtime.Config{
			Origin:   from,
			Value:    tx.Value(),
			State:    state,
			GasLimit: opts.GasLimit,
		}

		_, _, err = runtime.Call(addr, tx.Data(), &cfg)
		logs := state.Logs()
		if err == nil && len(logs) != 1 {
			t.Fatal("No logs or error after execution")
		} else if err != nil {
			return
		}

		// Test that our custom parsing matches the ABI parsing
		depositEvent, err := portalContract.ParseTransactionDeposited(*(logs[0]))
		if err != nil {
			t.Fatalf("Could not parse log that was emitted by the deposit contract: %v", err)
		}
		depositEvent.Raw = types.Log{} // Clear out the log

		// Verify that is passes our custom unmarshalling logic
		dep, err := UnmarshalLogEvent(logs[0])
		if err != nil {
			t.Fatalf("Could not unmarshal log that was emitted by the deposit contract: %v", err)
		}

		reconstructed := &deposit.OptimismPortalTransactionDeposited{
			From:       dep.From,
			Value:      dep.Value,
			GasLimit:   dep.Gas,
			IsCreation: dep.To == nil,
			Data:       dep.Data,
			Raw:        types.Log{},
		}
		if dep.To != nil {
			reconstructed.To = *dep.To
		}
		if dep.Mint != nil {
			reconstructed.Mint = dep.Mint
		} else {
			reconstructed.Mint = common.Big0
		}

		if !cmp.Equal(depositEvent, reconstructed, cmp.Comparer(BigEqual)) {
			t.Fatalf("The deposit tx did not match. tx: %v. actual: %v", reconstructed, depositEvent)
		}

		inputArgs := &deposit.OptimismPortalTransactionDeposited{
			From:       from,
			To:         to,
			Mint:       mint,
			Value:      value,
			GasLimit:   l2GasLimit,
			IsCreation: isCreation,
			Data:       data,
			Raw:        types.Log{},
		}
		if !cmp.Equal(depositEvent, inputArgs, cmp.Comparer(BigEqual)) {
			t.Fatalf("The input args did not match. input: %v. actual: %v", inputArgs, depositEvent)
		}
	})
}
