package bss

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BatchSubmitter struct {
	Client    *ethclient.Client
	ToAddress common.Address
	ChainID   *big.Int
	PrivKey   *ecdsa.PrivateKey
}

// func NewSubmitter(client ethclient.Client, addr common.Address) *BatchSubmitter {
// 	return &BatchSubmitter{client: client, addr: addr}
// }

// Submit creates & submits a batch to L1. Blocks until the transaction is included.
// Return the tx hash as well as a possible error.
func (b *BatchSubmitter) Submit(batch derive.BatchV1) (common.Hash, error) {
	enc, err := batch.MarshalBinary()
	if err != nil {
		return common.Hash{}, err
	}
	contract := bind.NewBoundContract(b.ToAddress, abi.ABI{}, b.Client, b.Client, b.Client)
	opts, err := bind.NewKeyedTransactorWithChainID(
		b.PrivKey, b.ChainID,
	)
	if err != nil {
		return common.Hash{}, err
	}
	tx, err := contract.RawTransact(opts, enc)
	if err != nil {
		return common.Hash{}, err
	}

	timeout := time.After(20 * time.Second)

	for {
		receipt, err := b.Client.TransactionReceipt(context.Background(), tx.Hash())
		if receipt != nil {
			return tx.Hash(), nil
		} else if err != nil {
			return common.Hash{}, err
		}
		<-time.After(150 * time.Millisecond)

		select {
		case <-timeout:
			return common.Hash{}, errors.New("timeout")
		default:
		}
	}

}
