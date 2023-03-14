package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/pkg/errors"
)

func (i *Instance) uponFillGap(signedFillGap *specalea.SignedMessage, fillgapMsgContainer *specalea.MsgContainer) error {

	// get data
	fillGapData, err := signedFillGap.Message.GetFillGapData()
	if err != nil {
		return errors.Wrap(err, "uponFillGap: could not get fillgap data from signedFillGap")
	}

	// Add message to container
	fillgapMsgContainer.AddMsg(signedFillGap)

	// get structure values
	operatorID := fillGapData.OperatorID
	priorityAsked := fillGapData.Priority

	// get the desired queue
	queue := i.State.VCBCState.Queues[operatorID]
	// get highest local priority
	_, priority := queue.PeekLast()

	// if has more entries than the asker (sender of the message), sends FILLER message with local entries
	if priority >= priorityAsked {
		// init values, priority list
		returnValues := make([][]*specalea.ProposalData, 0)
		returnPriorities := make([]specalea.Priority, 0)
		returnProofs := make([][]byte, 0)

		// get local values and priorities
		values := queue.GetValues()
		priorities := queue.GetPriorities()

		// for each, test if priority if above and, if so, adds to the FILLER list
		for idx, priority := range priorities {
			if priority >= priorityAsked {
				returnValues = append(returnValues, values[idx])
				returnPriorities = append(returnPriorities, priority)
				returnProofs = append(returnProofs, i.State.VCBCState.GetU(operatorID, priority))
			}
		}

		// sends FILLER message
		fillerMsg, err := CreateFiller(i.State, i.config, returnValues, returnPriorities, returnProofs, operatorID)
		if err != nil {
			return errors.Wrap(err, "uponFillGap: failed to create Filler message")
		}

		// FIX ME : send only to sender of fillGap msg
		i.Broadcast(fillerMsg)
	}

	return nil
}

func isValidFillGap(
	state *specalea.State,
	config alea.IConfig,
	signedMsg *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != specalea.FillGapMsgType {
		return errors.New("msg type is not FillGapMsgType")
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

	FillGapData, err := signedMsg.Message.GetFillGapData()
	if err != nil {
		return errors.Wrap(err, "could not get FillGapData data")
	}
	if err := FillGapData.Validate(); err != nil {
		return errors.Wrap(err, "FillGapData invalid")
	}

	// operatorID
	operatorID := FillGapData.OperatorID
	InCommittee := false
	for _, opID := range operators {
		if opID.OperatorID == operatorID {
			InCommittee = true
		}
	}
	if !InCommittee {
		return errors.New("author (OperatorID) doesn't exist in Committee")
	}

	return nil
}

func CreateFillGap(state *specalea.State, config alea.IConfig, operatorID types.OperatorID, priority specalea.Priority) (*specalea.SignedMessage, error) {
	fillgapData := &specalea.FillGapData{
		OperatorID: operatorID,
		Priority:   priority,
	}
	dataByts, err := fillgapData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateFillGap: could not encode fillgap data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.FillGapMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateFillGap: failed signing fillgap msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}