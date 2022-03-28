package node

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type l2EthClient interface {
	GetBlockHeader(ctx context.Context, blockTag string) (*types.Header, error)
	GetProof(ctx context.Context, address common.Address, blockTag string) (*common.Hash, []string, error)
}

type nodeAPI struct {
	client                 l2EthClient
	withdrawalContractAddr common.Address
	log                    log.Logger
}

func newNodeAPI(l2Client l2EthClient, withdrawalContractAddr common.Address, log log.Logger) *nodeAPI {
	return &nodeAPI{
		client:                 l2Client,
		withdrawalContractAddr: withdrawalContractAddr,
		log:                    log,
	}
}

func (n *nodeAPI) OutputAtBlock(ctx context.Context, number rpc.BlockNumber) ([]l2.Bytes32, error) {
	// TODO: rpc.BlockNumber doesn't support the "safe" tag. Need a new type

	head, err := n.client.GetBlockHeader(ctx, toBlockNumArg(number))
	if err != nil {
		n.log.Error("failed to get block", "err", err)
		return nil, err
	}
	if head == nil {
		return nil, ethereum.NotFound
	}

	accountRoot, proof, err := n.client.GetProof(ctx, n.withdrawalContractAddr, toBlockNumArg(number))
	if err != nil {
		n.log.Error("failed to get contract proof", "err", err)
		return nil, err
	}
	if accountRoot == nil {
		return nil, ethereum.NotFound
	}

	var l2OutputRootVersion l2.Bytes32 // it's zero for now

	// TODO: Figure out why this doesn't work work
	if err := VerifyAccountProof(head.Root, *accountRoot, proof); err != nil {
		n.log.Error("invalid withdrawal root detected in block", "stateRoot", head.Root, "blocknum", number, "msg", err)
		return nil, fmt.Errorf("invalid withdrawal root hash")
	}

	hash := ComputeL2OutputRoot(l2OutputRootVersion, head.Hash(), head.Root, *accountRoot)
	var l2OutputRootHash l2.Bytes32
	copy(l2OutputRootHash[:], hash)

	return []l2.Bytes32{l2OutputRootVersion, l2OutputRootHash}, nil
}

func toBlockNumArg(number rpc.BlockNumber) string {
	if number == rpc.LatestBlockNumber {
		return "latest"
	}
	if number == rpc.PendingBlockNumber {
		return "pending"
	}
	return hexutil.EncodeUint64(uint64(number.Int64()))
}
