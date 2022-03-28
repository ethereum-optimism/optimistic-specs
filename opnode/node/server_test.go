package node

import (
	"context"
	"testing"

	"github.com/ethereum-optimism/optimistic-specs/opnode/internal/testlog"
	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

func TestOutputAtBlock(t *testing.T) {
	log := testlog.Logger(t, log.LvlError)
	l2Client := &mockL2Client{
		head: &types.Header{},
		root: new(common.Hash),
	}
	var addr common.Address
	server, err := newRPCServer(context.Background(), "localhost", 0, l2Client, addr, log, "0.0")
	assert.NoError(t, err)
	assert.NoError(t, server.Start())
	defer server.Stop()

	client, err := dialRPCClientWithBackoff(context.Background(), log, "http://"+server.Addr().String())
	assert.NoError(t, err)

	println(server.httpServer.Addr)

	var result []l2.Bytes32
	err = client.CallContext(context.Background(), &result, "optimism_outputAtBlock", "latest")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

type mockL2Client struct {
	head  *types.Header
	root  *common.Hash
	proof []string
}

func (c *mockL2Client) GetBlockHeader(ctx context.Context, blockTag string) (*types.Header, error) {
	return c.head, nil
}

func (c *mockL2Client) GetProof(ctx context.Context, address common.Address, blockTag string) (*common.Hash, []string, error) {
	return c.root, c.proof, nil
}
