package l2

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

func (s *Source) getPayload(ctx context.Context, payloadId PayloadID) (*ExecutionPayload, error) {
	e := s.log.New("payload_id", payloadId)
	e.Debug("getting payload")
	var result ExecutionPayload
	err := s.rpc.CallContext(ctx, &result, "engine_getPayloadV1", payloadId)
	if err != nil {
		e = log.New("payload_id", "err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := ErrorCode(rpcErr.ErrorCode())
			if code != UnavailablePayload {
				e.Warn("unexpected error code in get-payload response", "code", code)
			} else {
				e.Warn("unavailable payload in get-payload request")
			}
		} else {
			e.Error("failed to get payload")
		}
		return nil, err
	}
	e.Debug("Received payload")
	return &result, nil
}

func (s *Source) executePayload(ctx context.Context, payload *ExecutionPayload) (*ExecutePayloadResult, error) {
	e := s.log.New("block_hash", payload.BlockHash)
	e.Debug("sending payload for execution")
	var result ExecutePayloadResult
	err := s.rpc.CallContext(ctx, &result, "engine_executePayloadV1", payload)
	if err != nil {
		e.Error("Payload execution failed", "err", err)
		return nil, err
	}
	e.Debug("Received payload execution result", "status", result.Status, "latestValidHash", result.LatestValidHash, "message", result.ValidationError)
	return &result, nil
}

func (s *Source) forkchoiceUpdated(ctx context.Context, state *ForkchoiceState, attr *PayloadAttributes) (ForkchoiceUpdatedResult, error) {
	e := s.log.New("state", state, "attr", attr)
	e.Debug("Sharing forkchoice-updated signal")

	var result ForkchoiceUpdatedResult
	err := s.rpc.CallContext(ctx, &result, "engine_forkchoiceUpdatedV1", state, attr)
	if err == nil {
		e.Debug("Shared forkchoice-updated signal")
		if attr != nil {
			e.Debug("Received payload id", "payloadId", result.PayloadID)
		}
		return result, nil
	} else {
		e = e.New("err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := ErrorCode(rpcErr.ErrorCode())
			e.Warn("Unexpected error code in forkchoice-updated response", "code", code)
		} else {
			e.Error("Failed to share forkchoice-updated signal")
		}
		return result, err
	}
}

// ExecutePayload executes the payload and parses the return status into a useful error code
func ExecutePayloadStatic(ctx context.Context, rpc internalRPC, payload *ExecutionPayload) error {
	execCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	execRes, err := rpc.executePayload(execCtx, payload)
	if err != nil {
		return fmt.Errorf("failed to execute payload: %v", err)
	}
	switch execRes.Status {
	case ExecutionValid:
		return nil
	case ExecutionSyncing:
		return fmt.Errorf("failed to execute payload %s, node is syncing, latest valid hash is %s", payload.ID(), execRes.LatestValidHash)
	case ExecutionInvalid:
		return fmt.Errorf("execution payload %s was INVALID! Latest valid hash is %s, ignoring bad block: %q", payload.ID(), execRes.LatestValidHash, execRes.ValidationError)
	default:
		return fmt.Errorf("unknown execution status on %s: %q, ", payload.ID(), string(execRes.Status))
	}
}

// ForkchoiceUpdate updates the forkchoive for L2 and parses the return status into a useful error code
func ForkchoiceUpdate(ctx context.Context, rpc internalRPC, fc *ForkchoiceState, attr *PayloadAttributes) (*ForkchoiceUpdatedResult, error) {
	fcCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	fcRes, err := rpc.forkchoiceUpdated(fcCtx, fc, attr)
	if err != nil {
		return nil, fmt.Errorf("failed to update forkchoice: %v", err)
	}
	switch fcRes.Status {
	case UpdateSyncing:
		return nil, fmt.Errorf("updated forkchoice, but node is syncing: %v", err)
	case UpdateSuccess:
		return &fcRes, nil
	default:
		return nil, fmt.Errorf("unknown forkchoice status on %s: %q, ", fc.SafeBlockHash, string(fcRes.Status))
	}
}
