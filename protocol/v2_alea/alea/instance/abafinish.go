package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/pkg/errors"
)

func (i *Instance) uponABAFinish(signedABAFinish *specalea.SignedMessage) error {

	// get data
	ABAFinishData, err := signedABAFinish.Message.GetABAFinishData()
	if err != nil {
		return errors.Wrap(err, "uponABAFinish: could not get ABAFinishData from signedABAConf")
	}

	// old message -> ignore
	if ABAFinishData.ACRound < i.State.ACState.ACRound {
		return nil
	}
	// if future round -> intialize future state
	if ABAFinishData.ACRound > i.State.ACState.ACRound {
		i.State.ACState.InitializeRound(ABAFinishData.ACRound)
	}

	abaState := i.State.ACState.GetABAState(ABAFinishData.ACRound)

	// add the message to the container
	abaState.ABAFinishContainer.AddMsg(signedABAFinish)

	// sender
	senderID := signedABAFinish.GetSigners()[0]

	alreadyReceived := abaState.HasFinish(senderID)
	// if never received this msg, update
	if !alreadyReceived {

		// get vote from FINISH message
		vote := ABAFinishData.Vote

		// increment counter
		abaState.SetFinish(senderID, vote)
	}

	// if FINISH(b) reached partial quorum and never broadcasted FINISH(b), broadcast
	if !abaState.SentFinish(byte(0)) && !abaState.SentFinish(byte(1)) {
		for _, vote := range []byte{0, 1} {

			if abaState.CountFinish(vote) >= i.State.Share.PartialQuorum {
				// broadcast FINISH
				finishMsg, err := CreateABAFinish(i.State, i.config, vote, ABAFinishData.ACRound)
				if err != nil {
					return errors.Wrap(err, "uponABAFinish: failed to create ABA Finish message")
				}
				i.Broadcast(finishMsg)

				// update sent flag
				abaState.SetSentFinish(vote, true)
				// process own finish msg
				i.uponABAFinish(finishMsg)
			}
		}
	}

	// if FINISH(b) reached Quorum, decide for b and send termination signal
	for _, vote := range []byte{0, 1} {
		if abaState.CountFinish(vote) >= i.State.Share.Quorum {
			abaState.SetDecided(vote)
			abaState.SetTerminate(true)
		}
	}

	return nil
}

func isValidABAFinish(
	state *specalea.State,
	config alea.IConfig,
	signedMsg *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != specalea.ABAFinishMsgType {
		return errors.New("msg type is not ABAFinishMsgType")
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

	ABAFinishData, err := signedMsg.Message.GetABAFinishData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAFinishData data")
	}
	if err := ABAFinishData.Validate(); err != nil {
		return errors.Wrap(err, "ABAFinishData invalid")
	}

	// vote
	vote := ABAFinishData.Vote
	if vote != 0 && vote != 1 {
		return errors.New("vote different than 0 and 1")
	}

	return nil
}

func CreateABAFinish(state *specalea.State, config alea.IConfig, vote byte, acRound specalea.ACRound) (*specalea.SignedMessage, error) {
	ABAFinishData := &specalea.ABAFinishData{
		Vote:    vote,
		ACRound: acRound,
	}
	dataByts, err := ABAFinishData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode abafinish data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.ABAFinishMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed signing abafinish msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
