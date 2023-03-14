package forks

import (
	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
)

// Fork is the interface for fork
type Fork interface {
	EncodeSignedMsg(msg *specqbft.SignedMessage) ([]byte, error)
	DecodeSignedMsg(data []byte) (*specqbft.SignedMessage, error)
}
