package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

// TODO(inphi): add metrics

type rpcServer struct {
	server *http.Server
	log    log.Logger
}

func newRPCServer(ctx context.Context, addr string, port int, l2RPCClient *rpc.Client, withdrawalContractAddress common.Address, log log.Logger, appVersion string) (*rpcServer, error) {
	api := &nodeAPI{
		l2RPCClient:            l2RPCClient,
		l2EthClient:            ethclient.NewClient(l2RPCClient),
		withdrawalContractAddr: withdrawalContractAddress,
		log:                    log,
	}
	apis := []rpc.API{{
		Namespace:     "optimism",
		Service:       api,
		Public:        true,
		Authenticated: false,
	}}

	srv := rpc.NewServer()
	if err := node.RegisterApis(apis, nil, srv, true); err != nil {
		return nil, err
	}
	nodeHandler := node.NewHTTPHandlerStack(srv, nil, []string{addr}, nil)

	mux := http.NewServeMux()
	mux.Handle("/", nodeHandler)
	mux.HandleFunc("/healthz", healthzHandler(appVersion))
	addrPort := fmt.Sprintf("%s:%d", addr, port)

	r := &rpcServer{
		log: log,
		server: &http.Server{
			Handler: mux,
			Addr:    addrPort,
		},
	}
	return r, nil
}

func (s *rpcServer) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				s.log.Info("RPC server shutdown")
				return
			}
			s.log.Error("error starting RPC server", "err", err)
		}
	}()
}

func (r *rpcServer) Stop() {
	_ = r.server.Shutdown(context.Background())
}

func healthzHandler(appVersion string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(appVersion))
	}
}

type nodeAPI struct {
	l2RPCClient            *rpc.Client
	l2EthClient            *ethclient.Client
	withdrawalContractAddr common.Address
	log                    log.Logger
}

type getProof struct {
	Address     common.Address `json:"address"`
	StorageHash common.Hash    `json:"storage_hash"`
}

func (n *nodeAPI) OutputAtBlock(ctx context.Context, number rpc.BlockNumber) ([]l2.Bytes32, error) {
	// TODO: rpc.BlockNumber doesn't support the "safe" tag. Need a new type

	numberB := big.NewInt(number.Int64())
	block, err := n.l2EthClient.BlockByNumber(ctx, numberB)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, ethereum.NotFound
	}

	var getProofResponse *getProof
	err = n.l2RPCClient.CallContext(ctx, &getProofResponse, "eth_getProof", n.withdrawalContractAddr, nil, toBlockNumArg(numberB))
	if err != nil {
		return nil, err
	} else if getProofResponse == nil {
		return nil, ethereum.NotFound
	}

	var l2OutputRootVersion l2.Bytes32 // it's zero for now

	var buf bytes.Buffer
	buf.Write(l2OutputRootVersion[:])
	buf.Write(block.Root().Bytes())
	buf.Write(getProofResponse.StorageHash[:])
	buf.Write(block.Hash().Bytes())
	hash := crypto.Keccak256(buf.Bytes())

	var l2OutputRootHash l2.Bytes32
	copy(l2OutputRootHash[:], hash)

	return []l2.Bytes32{l2OutputRootVersion, l2OutputRootHash}, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}
