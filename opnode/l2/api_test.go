package l2

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestExecutionPayloadUnmarshal(t *testing.T) {
	f, err := os.Open("testdata/executionPayload_good.json")
	require.NoError(t, err)

	var payload ExecutionPayload
	enc := json.NewDecoder(f)
	require.NoError(t, enc.Decode(&payload))

	txs := payload.Transactions()
	require.Equal(t, "0xe14d325bb8c1479761afc17f6ceb5f96b3350adeeea8a46d95b8b12538c70a35", payload.ParentHashField.String())
	require.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", payload.FeeRecipient.String())
	require.Equal(t, "0x797f6cc71520124ff5c828c83f5f03b521d3eb6ce7a2d68053b0a7301d821bd9", payload.StateRoot.String())
	require.Equal(t, "0x3fd774ff8fe515813d7a8f9d3748e58e857bc002823a458a93be90a3bc2e0894", payload.ReceiptsRoot.String())
	require.Equal(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", payload.LogsBloom.String())
	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000000", payload.PrevRandao.String())
	require.Equal(t, "0x1", payload.BlockNumber.String())
	require.Equal(t, "0x4c5e51", payload.GasLimit.String())
	require.Equal(t, "0x0", payload.GasUsed.String())
	require.Equal(t, "0x624f6bb7", payload.Timestamp.String())
	require.Equal(t, "0x", payload.ExtraData.String())
	require.Equal(t, "0x7", payload.BaseFeePerGas.String())
	require.Equal(t, "0x7a0bb12e6a2c78a6996252fd691529f285f721300abe5da650f977dfcbbabf75", payload.BlockHash.String())
	require.Equal(t, 1, len(txs))
	require.EqualValues(t, 1, txs[0].BlockHeight())
	require.Equal(t, &derive.L1InfoPredeployAddr, txs[0].To())
	require.Equal(t, new(big.Int), txs[0].Mint())
	require.EqualValues(t, 99999999, txs[0].Gas())
	require.Equal(t, common.FromHex("0x8c89c386000000000000000100000000624f6bca00000000000000000000000000000000000000000000000000000000342770c019e19e8b3faa3b597eeb94c2f88b0d484b77c75595d80a8239447b0c2a3fc593"), txs[0].Data())

	f.Close()

	f, err = os.Open("testdata/executionPayload_bad.json")
	require.NoError(t, err)
	defer f.Close()

	var badPayload ExecutionPayload
	enc = json.NewDecoder(f)
	require.Errorf(t, enc.Decode(&badPayload), "error unmarshaling execution payload transaction 1")
}
