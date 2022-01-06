package l2

import (
	"context"
	"fmt"
)

type BlockPreparer interface {
	GetPayload(ctx context.Context, payloadId PayloadID) (*ExecutionPayload, error)
	ForkchoiceUpdated(ctx context.Context, state *ForkchoiceState, attr *PayloadAttributes) (ForkchoiceUpdatedResult, error)
}

// DeriveBlock uses the engine API to derive a full L2 block from the block inputs.
// The fcState does not affect the block production, but may inform the engine of finality and head changes to sync towards before block computation.
func DeriveBlock(ctx context.Context, engine BlockPreparer, fcState *ForkchoiceState, attributes *PayloadAttributes) (*ExecutionPayload, error) {
	fcResult, err := engine.ForkchoiceUpdated(ctx, fcState, attributes)
	if err != nil {
		return nil, fmt.Errorf("engine failed to process forkchoice update for block derivation: %v", err)
	} else if fcResult.Status != UpdateSuccess {
		return nil, fmt.Errorf("engine not in sync, failed to derive block, status: %s", fcResult.Status)
	}

	payload, err := engine.GetPayload(ctx, fcResult.PayloadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payload: %v", err)
	}
	return payload, nil
}
