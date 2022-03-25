package node

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ethereum-optimism/optimistic-specs/opnode/l2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
	"github.com/protolambda/zssz"
)

// TODO(inphi): add metrics

var (
	ErrParseErr = &RPCErr{
		Code:          -32700,
		Message:       "parse error",
		HTTPErrorCode: 400,
	}
	ErrInternal = &RPCErr{
		Code:          JSONRPCErrorInternal,
		Message:       "internal error",
		HTTPErrorCode: 500,
	}
)

func ErrInvalidRequest(msg string) *RPCErr {
	return &RPCErr{
		Code:          -32601,
		Message:       msg,
		HTTPErrorCode: 400,
	}
}

type RPCHandler func(context.Context, *RPCReq) *RPCRes

type rpcServer struct {
	l2RPCClient            *rpc.Client
	server                 *http.Server
	withdrawalContractAddr common.Address
	log                    log.Logger
	appVersion             string
}

func newRPCServer(ctx context.Context, addr string, port int, l2RPCClient *rpc.Client, withdrawalContractAddress common.Address, log log.Logger, appVersionn string) *rpcServer {
	r := &rpcServer{
		l2RPCClient:            l2RPCClient,
		withdrawalContractAddr: withdrawalContractAddress,
		log:                    log,
		appVersion:             appVersionn,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.rootHandler)
	mux.HandleFunc("/healthz", r.healthzHandler)
	addrPort := fmt.Sprintf("%s:%d", addr, port)
	r.server = &http.Server{
		Handler: mux,
		Addr:    addrPort,
	}

	return r
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

func (s *rpcServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("received RPC request", "user_agent", r.Header.Get("user-agent"))

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.log.Error("error reading request body", "err", err)
		writeRPCError(s.log, w, nil, ErrInternal)
		return
	}

	req, err := ParseRPCReq(body)
	if err != nil {
		s.log.Info("error parsing RPC call", "err", err)
		writeRPCError(s.log, w, nil, err)
		return
	}

	if err := ValidateRPCReq(req); err != nil {
		writeRPCRes(s.log, w, NewRPCErrorRes(nil, err))
		return
	}

	if req.Method != "optimism_outputAtBlock" {
		writeRPCRes(s.log, w, NewRPCErrorRes(nil, fmt.Errorf("unknown method")))
		return
	}

	res := s.outputAtBlock(r.Context(), req)
	writeRPCRes(s.log, w, res)
}

func (s *rpcServer) healthzHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(s.appVersion))
}

func writeRPCError(log log.Logger, w http.ResponseWriter, id json.RawMessage, err error) {
	var res *RPCRes
	if r, ok := err.(*RPCErr); ok {
		res = NewRPCErrorRes(id, r)
	} else {
		res = NewRPCErrorRes(id, ErrInternal)
	}
	writeRPCRes(log, w, res)
}

func writeRPCRes(log log.Logger, w http.ResponseWriter, res *RPCRes) {
	statusCode := 200
	if res.IsError() && res.Error.HTTPErrorCode != 0 {
		statusCode = res.Error.HTTPErrorCode
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(res); err != nil {
		log.Error("error writing rpc response", "err", err)
		return
	}
}

type rpcBlock struct {
	*types.Header
	Hash         common.Hash        `json:"hash"`
	Tranasctions types.Transactions `json:"transactions"`
}

type getProof struct {
	Address     common.Address `json:"address"`
	StorageHash common.Hash    `json:"storage_hash"`
}

func (s *rpcServer) outputAtBlock(ctx context.Context, req *RPCReq) *RPCRes {
	var params []string
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewRPCErrorRes(req.ID, err)
	}
	if len(params) != 1 {
		return NewRPCErrorRes(req.ID, fmt.Errorf("invalid number of parameters"))
	}

	blockTag := params[0]

	if blockTag == "safe" {
		// TODO: handle this via getBlockNumber() - FINALIZED_PERIOD
	}

	var block *rpcBlock
	err := s.l2RPCClient.CallContext(ctx, &block, "eth_getBlockByNumber", blockTag, false)
	if err == nil {
		return NewRPCErrorRes(req.ID, err)
	}
	if block == nil {
		return NewRPCErrorRes(req.ID, ethereum.NotFound)
	}

	var getProofResponse *getProof
	err = s.l2RPCClient.CallContext(ctx, &getProofResponse, "eth_getProof", s.withdrawalContractAddr, nil, blockTag)
	if err != nil {
		return NewRPCErrorRes(req.ID, err)
	} else if getProofResponse == nil {
		return NewRPCErrorRes(req.ID, ethereum.NotFound)
	}

	var random l2.Bytes32
	copy(random[:], block.Difficulty.Bytes())

	txs := make([]l2.Data, 0, len(block.Tranasctions))
	for _, tx := range block.Tranasctions {
		w := new(bytes.Buffer)
		if err := tx.EncodeRLP(w); err != nil {
			return NewRPCErrorRes(req.ID, fmt.Errorf("invalid transaction found in block"))
		}
		txs = append(txs, w.Bytes())
	}

	payload := &l2.ExecutionPayload{
		ParentHash:    block.ParentHash,
		FeeRecipient:  block.Coinbase,
		StateRoot:     l2.Bytes32(block.Root),
		ReceiptsRoot:  l2.Bytes32(block.ReceiptHash),
		LogsBloom:     l2.Bytes256(block.Bloom),
		Random:        random,
		BlockNumber:   hexutil.Uint64(block.Number.Uint64()),
		GasLimit:      hexutil.Uint64(block.GasLimit),
		GasUsed:       hexutil.Uint64(block.GasUsed),
		Timestamp:     hexutil.Uint64(block.Time),
		ExtraData:     block.Extra,
		BaseFeePerGas: *uint256.NewInt(block.BaseFee.Uint64()),
		BlockHash:     block.Hash,
		Transactions:  txs,
	}

	l2Output := &l2.L2Output{
		StateRoot:              payload.StateRoot,
		WithdrawalStorageRoot:  l2.Bytes32(getProofResponse.StorageHash),
		LatestBlock:            payload,
		HistoryAccumulatorRoot: l2.Bytes32{}, // unused. zeroed
		Extension:              l2.Bytes32{}, // unused. zeroed
	}

	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	l2OutputSSZ := zssz.GetSSZ((*l2.L2Output)(nil))
	if _, err := zssz.Encode(bufWriter, &l2Output, l2OutputSSZ); err != nil {
		return NewRPCErrorRes(req.ID, err)
	}
	if err := bufWriter.Flush(); err != nil {
		return NewRPCErrorRes(req.ID, err)
	}

	var l2OutputRootVersion l2.Bytes32 // it's zero for now

	hash := crypto.Keccak256(buf.Bytes())
	//crypto.Keccak256(l2OutputRootVersion[:], block.Root()[:], res.StorageHash.Bytes(), )

	var l2OutputRootHash l2.Bytes32
	copy(l2OutputRootHash[:], hash)

	res := new(RPCRes)
	res.ID = req.ID
	res.Result = []l2.Bytes32{l2OutputRootVersion, l2OutputRootHash}
	res.JSONRPC = JSONRPCVersion
	return res
}
