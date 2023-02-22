package instance

import (
	"crypto/sha256"
	"encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/pkg/errors"
)

// uponProposal process proposal message
// Assumes proposal message is valid!
func (i *Instance) uponProposal(signedProposal *specalea.SignedMessage, proposeMsgContainer *specalea.MsgContainer) error {

	// get Data
	proposalDataReceived, err := signedProposal.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "uponProposal: could not get proposal data from signedProposal")
	}

	// check if message has been already delivered
	if i.State.Delivered.HasProposal(proposalDataReceived) {

		return errors.New("proposal already delivered")
	}

	// Add message to container
	proposeMsgContainer.AddMsg(signedProposal)

	// add to vcbc state
	i.State.VCBCState.AppendToM(i.State.Share.OperatorID, i.State.VCBCState.Priority, proposalDataReceived)

	// Check if container has less maximum size. If so, returns
	if len(i.State.VCBCState.GetM(i.State.Share.OperatorID, i.State.VCBCState.Priority)) < i.State.BatchSize {

		return nil
	}

	// else, starts VCBC
	i.StartVCBC(i.State.VCBCState.Priority)

	// Increment priority
	i.State.VCBCState.Priority += 1
	return nil
}

// Encode returns the list encoded bytes or error
func EncodeProposals(proposals []*specalea.ProposalData) ([]byte, error) {
	return json.Marshal(proposals)
}

// Decode returns error if decoding failed
func DecodeProposals(data []byte) ([]*specalea.ProposalData, error) {
	proposals := make([]*specalea.ProposalData, 0)
	err := json.Unmarshal(data, &proposals)
	if err != nil {
		return nil, errors.Wrap(err, "DecodeProposals: could not unmarshal proposals")
	}
	return proposals, nil
}

// GetHash returns the SHA-256 hash
func GetProposalsHash(proposals []*specalea.ProposalData) ([]byte, error) {
	encoded, err := EncodeProposals(proposals)
	if err != nil {
		return nil, errors.Wrap(err, "GetProposalsHash: could not encode proposals")
	}
	ret := sha256.Sum256(encoded)
	return ret[:], nil
}

func isValidProposal(
	state *specalea.State,
	config alea.IConfig,
	signedProposal *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedProposal.Message.MsgType != specalea.ProposalMsgType {
		return errors.New("msg type is not proposal")
	}
	if signedProposal.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedProposal.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedProposal.Signature.VerifyByOperators(signedProposal, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	proposalData, err := signedProposal.Message.GetProposalData()
	if err != nil {
		return errors.Wrap(err, "could not get proposal data")
	}
	if err := proposalData.Validate(); err != nil {
		return errors.Wrap(err, "proposalData invalid")
	}

	return nil
}

// CreateProposal
func CreateProposal(state *specalea.State, config alea.IConfig, value []byte) (*specalea.SignedMessage, error) {
	proposalData := &specalea.ProposalData{
		Data: value,
		// RoundChangeJustification: roundChanges,
		// PrepareJustification:     prepares,
	}
	dataByts, err := proposalData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode proposal data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.ProposalMsgType,
		Height:     state.Height,
		Round:      state.AleaDefaultRound,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing prepare msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
