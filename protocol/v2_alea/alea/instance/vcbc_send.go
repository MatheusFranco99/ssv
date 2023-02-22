package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/pkg/errors"
)

func (i *Instance) uponVCBCSend(signedMessage *specalea.SignedMessage) error {

	// get Data
	vcbcSendData, err := signedMessage.Message.GetVCBCSendData()
	if err != nil {
		errors.New("uponVCBCSend: could not get vcbcSendData data from signedMessage")
	}

	// check if it was already received. If not -> store
	if !i.State.VCBCState.HasM(vcbcSendData.Author, vcbcSendData.Priority) {
		i.State.VCBCState.SetM(vcbcSendData.Author, vcbcSendData.Priority, vcbcSendData.Proposals)
	}

	// get sender of the message
	senderID := signedMessage.GetSigners()[0]

	// if it's the sender and the author, just add the ready signature
	if senderID == i.State.Share.OperatorID && senderID == vcbcSendData.Author {
		i.AddOwnVCBCReady(vcbcSendData.Proposals, vcbcSendData.Priority)
		return nil
	}

	// if the Author of the VCBC is the same as the sender of the message -> sign and answer with READY
	if senderID == vcbcSendData.Author {

		hash, err := GetProposalsHash(vcbcSendData.Proposals)
		if err != nil {
			return errors.New("uponVCBCSend: could not get hash of proposals")
		}

		// create VCBCReady message with proof
		vcbcReadyMsg, err := CreateVCBCReady(i.State, i.config, hash, vcbcSendData.Priority, vcbcSendData.Author)
		if err != nil {
			return errors.New("uponVCBCSend: failed to create VCBCReady message with proof")
		}

		// FIX ME : send specifically to author
		i.Broadcast(vcbcReadyMsg)
	}

	return nil
}

func isValidVCBCSend(
	state *specalea.State,
	config alea.IConfig,
	signedMsg *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != specalea.VCBCSendMsgType {
		return errors.New("msg type is not VCBCSend")
	}
	if signedMsg.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedMsg.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	VCBCSendData, err := signedMsg.Message.GetVCBCSendData()
	if err != nil {
		return errors.Wrap(err, "could not get vcbcsend data")
	}
	if err := VCBCSendData.Validate(); err != nil {
		return errors.Wrap(err, "VCBCSendData invalid")
	}

	// author
	author := VCBCSendData.Author
	authorInCommittee := false
	for _, opID := range operators {
		if opID.OperatorID == author {
			authorInCommittee = true
		}
	}
	if !authorInCommittee {
		return errors.New("author (OperatorID) doesn't exist in Committee")
	}

	if author != signedMsg.GetSigners()[0] {
		return errors.New("author of VCBCSend differs from sender of the message")
	}

	// priority
	priority := VCBCSendData.Priority
	if state.VCBCState.HasM(author, priority) {
		if !state.VCBCState.EqualM(author, priority, VCBCSendData.Proposals) {
			return errors.New("existing (priority,author) with different proposals")
		}
	}

	return nil
}

func CreateVCBCSend(state *specalea.State, config alea.IConfig, proposals []*specalea.ProposalData, priority specalea.Priority, author types.OperatorID) (*specalea.SignedMessage, error) {
	vcbcSendData := &specalea.VCBCSendData{
		Proposals: proposals,
		Priority:  priority,
		Author:    author,
	}
	dataByts, err := vcbcSendData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCSend: could not encode vcbcSendData")
	}
	msg := &specalea.Message{
		MsgType:    specalea.VCBCSendMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateVCBCSend: failed signing filler msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
