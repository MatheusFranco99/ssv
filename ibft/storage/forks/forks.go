package forks

import (
	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
)

// Fork is the interface for fork
type Fork interface {
	EncodeSignedMsg(msg *messages.SignedMessage) ([]byte, error)
	DecodeSignedMsg(data []byte) (*messages.SignedMessage, error)
}
