package genesis

import (
	"fmt"
	"testing"

	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"

	"github.com/stretchr/testify/require"
)

func TestSSVMsgID(t *testing.T) {
	t.Run("consensus msg", func(t *testing.T) {
		f := ForkGenesis{}
		msgData := `{"message":{"type":3,"round":1,"identifier":"OTFiZGZjOWQxYzU4NzZkYTEwY...","height":28276,"value":"mB0aAAAAAAA4AAAAAAAAADpTC1djq..."},"signature":"jrB0+Z9zyzzVaUpDMTlCt6Om9mj...","signer_ids":[2,3,4]}`
		msg := spectypes.SSVMessage{
			MsgType: spectypes.SSVConsensusMsgType,
			MsgID:   specqbft.ControllerIdToMessageID([]byte("OTFiZGZjOWQxYzU4NzZkYTEwY")),
			Data:    []byte(msgData),
		}
		raw, err := msg.Encode()
		require.NoError(t, err)
		mid := f.MsgID()(raw)
		require.Greater(t, len(mid), 0)
		require.Equal(t, "0c42d6c5dc88b9a15d8c3b9b", fmt.Sprintf("%x", mid))
	})

	t.Run("empty msg", func(t *testing.T) {
		f := ForkGenesis{}
		require.Len(t, f.MsgID()([]byte{}), 0)
		require.Len(t, f.MsgID()(nil), 0)
	})
}
