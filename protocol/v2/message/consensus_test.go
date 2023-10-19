package message

import (
	"sort"
	"testing"

	protocoltesting "github.com/MatheusFranco99/ssv/protocol/v2/testing"

	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/stretchr/testify/require"
)

func TestAggregateSorting(t *testing.T) {
	uids := []spectypes.OperatorID{spectypes.OperatorID(1), spectypes.OperatorID(2), spectypes.OperatorID(3), spectypes.OperatorID(4)}
	secretKeys, _ := protocoltesting.GenerateBLSKeys(uids...)

	identifier := []byte("pk")

	generateSignedMsg := func(operatorId spectypes.OperatorID) *specqbft.SignedMessage {
		return protocoltesting.SignMsg(t, secretKeys, []spectypes.OperatorID{operatorId}, &specqbft.Message{
			MsgType:    specqbft.CommitMsgType,
			Height:     0,
			Round:      1,
			Identifier: identifier,
		})
	}

	signedMessage := generateSignedMsg(1)
	for i := 2; i <= 4; i++ {
		sig := generateSignedMsg(spectypes.OperatorID(i))
		require.NoError(t, Aggregate(signedMessage, sig))
	}

	sorted := sort.SliceIsSorted(signedMessage.Signers, func(i, j int) bool {
		return signedMessage.Signers[i] < signedMessage.Signers[j]
	})
	require.True(t, sorted)
}
