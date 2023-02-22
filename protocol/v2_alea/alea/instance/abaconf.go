package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/pkg/errors"
)

func (i *Instance) uponABAConf(signedABAConf *specalea.SignedMessage) error {

	// get data
	ABAConfData, err := signedABAConf.Message.GetABAConfData()
	if err != nil {
		return errors.Wrap(err, "uponABAConf:could not get ABAConfData from signedABAConf")
	}

	// old message -> ignore
	if ABAConfData.ACRound < i.State.ACState.ACRound {
		return nil
	}
	if ABAConfData.Round < i.State.ACState.GetCurrentABAState().Round {
		return nil
	}
	// if future round -> intialize future state
	if ABAConfData.ACRound > i.State.ACState.ACRound {
		i.State.ACState.InitializeRound(ABAConfData.ACRound)
	}
	if ABAConfData.Round > i.State.ACState.GetABAState(ABAConfData.ACRound).Round {
		i.State.ACState.GetABAState(ABAConfData.ACRound).InitializeRound(ABAConfData.Round)
	}

	abaState := i.State.ACState.GetABAState(ABAConfData.ACRound)

	// add the message to the containers
	abaState.ABAConfContainer.AddMsg(signedABAConf)

	// sender
	senderID := signedABAConf.GetSigners()[0]

	alreadyReceived := abaState.HasConf(ABAConfData.Round, senderID)

	// if never received this msg, update
	if !alreadyReceived {
		abaState.SetConf(ABAConfData.Round, senderID, ABAConfData.Votes)

	}

	// reached strong support -> try to decide value
	if abaState.CountConf(ABAConfData.Round) >= i.State.Share.Quorum {

		q := abaState.CountConfContainedInValues(ABAConfData.Round)
		if q < i.State.Share.Quorum {
			return nil
		}

		// get common coin
		s := i.config.GetCoinF()(abaState.Round)

		// if values = {0,1}, choose randomly (i.e. coin) value for next round
		if len(abaState.Values[ABAConfData.Round]) == 2 {

			abaState.SetVInput(ABAConfData.Round+1, s)

		} else {
			abaState.SetVInput(ABAConfData.Round+1, abaState.GetValues(ABAConfData.Round)[0])

			// if value has only one value, sends FINISH
			if abaState.GetValues(ABAConfData.Round)[0] == s {
				// check if indeed never sent FINISH message for this vote
				if !abaState.SentFinish(s) {
					finishMsg, err := CreateABAFinish(i.State, i.config, s, ABAConfData.ACRound)
					if err != nil {
						return errors.Wrap(err, "uponABAConf: failed to create ABA Finish message")
					}
					i.Broadcast(finishMsg)
					// update sent finish flag
					abaState.SetSentFinish(s, true)
					// process own finish msg
					i.uponABAFinish(finishMsg)
				}
			}
		}

		// increment round
		abaState.IncrementRound()

		// start new round sending INIT message with vote
		initMsg, err := CreateABAInit(i.State, i.config, abaState.GetVInput(abaState.Round), abaState.Round, ABAConfData.ACRound)
		if err != nil {
			return errors.Wrap(err, "uponABAConf: failed to create ABA Init message")
		}
		i.Broadcast(initMsg)
		// update sent init flag
		abaState.SetSentInit(abaState.Round, abaState.GetVInput(abaState.Round), true)
		// process own aux msg
		i.uponABAInit(initMsg)
	}

	return nil
}

func isValidABAConf(
	state *specalea.State,
	config alea.IConfig,
	signedMsg *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != specalea.ABAConfMsgType {
		return errors.New("msg type is not ABAConfMsgType")
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

	ABAConfData, err := signedMsg.Message.GetABAConfData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAConfData data")
	}
	if err := ABAConfData.Validate(); err != nil {
		return errors.Wrap(err, "ABAConfData invalid")
	}

	// vote
	votes := ABAConfData.Votes
	for _, vote := range votes {
		if vote != 0 && vote != 1 {
			return errors.New("vote different than 0 and 1")
		}
	}

	return nil
}

func CreateABAConf(state *specalea.State, config alea.IConfig, votes []byte, round specalea.Round, acRound specalea.ACRound) (*specalea.SignedMessage, error) {
	ABAConfData := &specalea.ABAConfData{
		Votes:   votes,
		Round:   round,
		ACRound: acRound,
	}
	dataByts, err := ABAConfData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAConf: could not encode abaconf data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.ABAConfMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAConf: failed signing abaconf msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
