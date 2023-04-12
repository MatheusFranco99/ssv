package genesis

import (
	"bytes"
	"testing"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"

	"github.com/stretchr/testify/require"
)

func TestForkV1_Encoding(t *testing.T) {
	msg := &spectypes.SSVMessage{
		MsgType: spectypes.SSVConsensusMsgType,
		MsgID:   specalea.ControllerIdToMessageID([]byte("xxxxxxxxxxx_ATTESTER")),
		Data:    []byte("data"),
	}
	f := &ForkGenesis{}

	b, err := f.EncodeNetworkMsg(msg)
	require.NoError(t, err)
	require.Greater(t, len(b), 0)

	res, err := f.DecodeNetworkMsg(b)
	require.NoError(t, err)
	require.Equal(t, msg.MsgType, res.MsgType)
	require.True(t, bytes.Equal(msg.Data, res.Data))
}
