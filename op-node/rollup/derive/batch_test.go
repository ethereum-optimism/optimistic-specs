package derive

import (
	"bytes"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/rollup"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestBatchRoundTrip(t *testing.T) {
	batches := []*BatchData{
		{
			BatchV1: BatchV1{
				Epoch:        0,
				Timestamp:    0,
				Transactions: []hexutil.Bytes{},
			},
		},
		{
			BatchV1: BatchV1{
				Epoch:        1,
				Timestamp:    1647026951,
				Transactions: []hexutil.Bytes{[]byte{0, 0, 0}, []byte{0x76, 0xfd, 0x7c}},
			},
		},
	}

	for i, batch := range batches {
		enc, err := batch.MarshalBinary()
		assert.NoError(t, err)
		var dec BatchData
		err = dec.UnmarshalBinary(enc)
		assert.NoError(t, err)
		assert.Equal(t, batch, &dec, "Batch not equal test case %v", i)
	}
	var buf bytes.Buffer
	err := EncodeBatches(&rollup.Config{}, batches, &buf)
	assert.NoError(t, err)
	out, err := DecodeBatches(&rollup.Config{}, &buf)
	assert.NoError(t, err)
	assert.Equal(t, batches, out)
}
