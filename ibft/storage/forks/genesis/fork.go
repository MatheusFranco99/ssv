package genesis

import (
	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
)

// ForkGenesis is the genesis fork for controller
type ForkGenesis struct {
}

// EncodeSignedMsg encodes signed message
func (f ForkGenesis) EncodeSignedMsg(msg *messages.SignedMessage) ([]byte, error) {
	return msg.Encode()
}

// DecodeSignedMsg decodes signed message
func (f ForkGenesis) DecodeSignedMsg(data []byte) (*messages.SignedMessage, error) {
	msgV1 := &messages.SignedMessage{}
	err := msgV1.Decode(data)
	if err != nil {
		return nil, err
	}
	return msgV1, nil
}
