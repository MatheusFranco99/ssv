package message

import (
	"sort"

	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
)

// Aggregate is a utility that helps to ensure sorted signers
func Aggregate(signedMsg *messages.SignedMessage, s spectypes.MessageSignature) error {
	if err := signedMsg.Aggregate(s); err != nil {
		return err
	}
	sort.Slice(signedMsg.Signers, func(i, j int) bool {
		return signedMsg.Signers[i] < signedMsg.Signers[j]
	})
	return nil
}
