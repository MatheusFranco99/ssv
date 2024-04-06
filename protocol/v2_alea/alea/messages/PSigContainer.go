package messages

import (
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"
)

type PSigContainer struct {
	Signatures map[types.OperatorID][]byte
	// Quorum is the number of min signatures needed for quorum
	Quorum uint64
}

func NewPSigContainer(quorum uint64) *PSigContainer {
	return &PSigContainer{
		Quorum:     quorum,
		Signatures: make(map[types.OperatorID][]byte),
	}
}

func (ps *PSigContainer) AddSignature(signer types.OperatorID, signature types.Signature) {
	if ps.Signatures[signer] == nil {
		ps.Signatures[signer] = make([]byte, 96)
		copy(ps.Signatures[signer], signature)
	}
}

func (ps *PSigContainer) ReconstructSignature(root []byte, validatorPubKey []byte) (types.Signature, error) {
	// Reconstruct signatures
	signature, err := types.ReconstructSignatures(ps.Signatures)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reconstruct signatures")
	}
	return signature.Serialize(), nil
}

func (ps *PSigContainer) ReconstructSignatureAndVerify(root []byte, validatorPubKey []byte) (types.Signature, error) {
	// Reconstruct signatures
	signature, err := types.ReconstructSignatures(ps.Signatures)
	if err != nil {
		return nil, errors.Wrap(err, "failed to reconstruct signatures")
	}
	return signature.Serialize(), nil
}

func (ps *PSigContainer) AggregateSignatures(validatorPubKey []byte) (types.Signature, error) {
	// Reconstruct signatures

	var ans types.Signature
	for _, signature := range ps.Signatures {
		if ans == nil || len(ans) == 0 {
			ans = signature
		} else {
			new_sig, err := ans.Aggregate(signature)
			if err != nil {
				return []byte{}, nil
			}
			ans = new_sig
		}
	}
	return ans, nil
}

func (ps *PSigContainer) GetLen() int {
	return int(len(ps.Signatures))
}

func (ps *PSigContainer) HasQuorum() bool {
	return uint64(len(ps.Signatures)) >= ps.Quorum
}
