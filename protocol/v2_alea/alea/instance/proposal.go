package instance

import (
	"crypto/sha256"
	"encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// uponProposal process proposal message
// Assumes proposal message is valid!
func (i *Instance) uponProposal(signedProposal *messages.SignedMessage, proposeMsgContainer *specalea.MsgContainer) error {

	// get Data
	// proposalDataReceived, err := signedProposal.Message.GetProposalData()
	// if err != nil {
	// 	return errors.Wrap(err, "uponProposal: could not get proposal data from signedProposal")
	// }

	// sender
	senderID := signedProposal.GetSigners()[0]

	i.logger.Debug("$$$$$$ UponProposal start", zap.Int("sender", int(senderID)))

	// // check if message has been already delivered
	// if i.State.Delivered.HasProposal(proposalDataReceived) {

	// 	return errors.New("proposal already delivered")
	// }

	// // Add message to container
	// proposeMsgContainer.AddMsg(signedProposal)

	// // add to vcbc state
	// i.State.VCBCState.AppendToM(i.State.Share.OperatorID, i.State.VCBCState.Priority, proposalDataReceived)

	// // Check if container has less maximum size. If so, returns
	// if len(i.State.VCBCState.GetM(i.State.Share.OperatorID, i.State.VCBCState.Priority)) < i.State.BatchSize {

	// 	return nil
	// }

	// i.logger.Debug("$$$$$$ UponProposal startvcbc", zap.Int("sender", int(senderID)))

	// // else, starts VCBC
	// i.StartVCBC(i.State.VCBCState.Priority)

	// // Increment priority
	// i.State.VCBCState.Priority += 1

	// i.logger.Debug("$$$$$$ UponProposal finish", zap.Int("sender", int(senderID)))

	return nil
}

func (i *Instance) onStartValue(signedProposal *messages.SignedMessage) error {

	// // get Data
	// proposalDataReceived, err := signedProposal.Message.GetProposalData()
	// if err != nil {
	// 	return errors.Wrap(err, "OnStartValue: could not get proposal data from signedProposal")
	// }

	// value := proposalDataReceived.Data

	// // sender
	// senderID := signedProposal.GetSigners()[0]

	// i.logger.Debug("$$$$$$ OnStartValue start", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)), zap.ByteString("value", value))

	// // check if message has been already delivered
	// if i.State.Delivered.HasProposal(proposalDataReceived) {
	// 	i.logger.Debug("$$$$$$ OnStartValue finish proposal already delivered", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)), zap.ByteString("value", value))
	// 	return errors.New("proposal already delivered")
	// }

	// add to vcbc state
	// i.State.VCBCState.AppendToM(i.State.Share.OperatorID, i.State.VCBCState.Priority, proposalDataReceived)

	// // Check if container has less maximum size. If so, returns
	// if len(i.State.VCBCState.GetM(i.State.Share.OperatorID, i.State.VCBCState.Priority)) < i.State.BatchSize {
	// 	i.logger.Debug("$$$$$$ OnStartValue finish didnt reach batch value", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)), zap.Int("current batch size", int(len(i.State.VCBCState.GetM(i.State.Share.OperatorID, i.State.VCBCState.Priority)))))
	// 	return nil
	// }

	// // i.logger.Debug("$$$$$$ OnStartValue startvcbc")

	// // else, starts VCBC
	// i.StartVCBC(i.State.VCBCState.Priority)

	// // Increment priority
	// i.State.VCBCState.Priority += 1

	// i.logger.Debug("$$$$$$ OnStartValue finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)), zap.ByteString("value", value))

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

func GetDataHash(data []byte) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "GetDataHash: could not encode data")
	}
	ret := sha256.Sum256(encoded)
	return ret[:], nil
}

func isValidProposal(
	state *messages.State,
	config alea.IConfig,
	signedProposal *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedProposal.Message.MsgType != messages.ProposalMsgType {
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
func CreateProposal(state *messages.State, config alea.IConfig, value []byte) (*messages.SignedMessage, error) {
	proposalData := &specalea.ProposalData{
		Data: value,
		// RoundChangeJustification: roundChanges,
		// PrepareJustification:     prepares,
	}
	dataByts, err := proposalData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode proposal data")
	}
	msg := &messages.Message{
		MsgType:    messages.ProposalMsgType,
		Height:     state.Height,
		Round:      state.AleaDefaultRound,
		Identifier: state.ID,
		Data:       dataByts,
	}

	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing prepare msg")
	}

	signedMsg := &messages.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
