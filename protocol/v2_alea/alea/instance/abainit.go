package instance

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/pkg/errors"
)

func (i *Instance) uponABAInit(signedABAInit *specalea.SignedMessage) error {

	// get data
	abaInitData, err := signedABAInit.Message.GetABAInitData()
	if err != nil {
		return errors.Wrap(err, "uponABAInit: could not get abainitdata from signedABAInit")
	}

	// old message -> ignore
	if abaInitData.ACRound < i.State.ACState.ACRound {
		return nil
	}
	if abaInitData.Round < i.State.ACState.GetCurrentABAState().Round {
		return nil
	}
	// if future round -> intialize future state
	if abaInitData.ACRound > i.State.ACState.ACRound {
		i.State.ACState.InitializeRound(abaInitData.ACRound)
	}
	if abaInitData.Round > i.State.ACState.GetABAState(abaInitData.ACRound).Round {
		i.State.ACState.GetABAState(abaInitData.ACRound).InitializeRound(abaInitData.Round)
	}

	abaState := i.State.ACState.GetABAState(abaInitData.ACRound)

	// add the message to the container
	abaState.ABAInitContainer.AddMsg(signedABAInit)

	// sender
	senderID := signedABAInit.GetSigners()[0]

	if !abaState.HasInit(abaInitData.Round, senderID, abaInitData.Vote) {
		abaState.SetInit(abaInitData.Round, senderID, abaInitData.Vote)
	}

	// weak support -> send INIT
	// if never sent INIT(b) but reached PartialQuorum (i.e. f+1, weak support), send INIT(b)
	for _, vote := range []byte{0, 1} {
		if !abaState.SentInit(abaInitData.Round, vote) && abaState.CountInit(abaInitData.Round, vote) >= i.State.Share.PartialQuorum {

			// send INIT
			initMsg, err := CreateABAInit(i.State, i.config, vote, abaInitData.Round, abaInitData.ACRound)
			if err != nil {
				return errors.Wrap(err, "uponABAInit: failed to create ABA Init message after weak support")
			}
			i.Broadcast(initMsg)
			// update sent flag
			abaState.SetSentInit(abaInitData.Round, vote, true)
			// process own init msg
			i.uponABAInit(initMsg)
		}
	}

	// strong support -> send AUX
	// if never sent AUX(b) but reached Quorum (i.e. 2f+1, strong support), sends AUX(b) and add b to values
	for _, vote := range []byte{0, 1} {

		if !abaState.SentAux(abaInitData.Round, vote) && abaState.CountInit(abaInitData.Round, vote) >= i.State.Share.Quorum {

			// append vote

			abaState.AddToValues(abaInitData.Round, vote)

			// sends AUX(b)
			auxMsg, err := CreateABAAux(i.State, i.config, vote, abaInitData.Round, abaInitData.ACRound)
			if err != nil {
				return errors.Wrap(err, "uponABAInit: failed to create ABA Aux message after strong init support")
			}
			i.Broadcast(auxMsg)

			// update sent flag
			abaState.SetSentAux(abaInitData.Round, vote, true)
			// process own aux msg
			i.uponABAAux(auxMsg)
		}
	}

	return nil
}

func isValidABAInit(
	state *specalea.State,
	config alea.IConfig,
	signedMsg *specalea.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
) error {
	if signedMsg.Message.MsgType != specalea.ABAInitMsgType {
		return errors.New("msg type is not specalea.ABAInitMsgType")
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

	ABAInitData, err := signedMsg.Message.GetABAInitData()
	if err != nil {
		return errors.Wrap(err, "could not get ABAInitData data")
	}
	if err := ABAInitData.Validate(); err != nil {
		return errors.Wrap(err, "ABAInitData invalid")
	}

	// vote
	vote := ABAInitData.Vote
	if vote != 0 && vote != 1 {
		return errors.New("vote different than 0 and 1")
	}

	return nil
}

func CreateABAInit(state *specalea.State, config alea.IConfig, vote byte, round specalea.Round, acRound specalea.ACRound) (*specalea.SignedMessage, error) {
	ABAInitData := &specalea.ABAInitData{
		Vote:    vote,
		Round:   round,
		ACRound: acRound,
	}
	dataByts, err := ABAInitData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAInit: could not encode abainit data")
	}
	msg := &specalea.Message{
		MsgType:    specalea.ABAInitMsgType,
		Height:     state.Height,
		Round:      state.Round,
		Identifier: state.ID,
		Data:       dataByts,
	}
	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
	if err != nil {
		return nil, errors.Wrap(err, "CreateABAInit: failed signing abainit msg")
	}

	signedMsg := &specalea.SignedMessage{
		Signature: sig,
		Signers:   []types.OperatorID{state.Share.OperatorID},
		Message:   msg,
	}
	return signedMsg, nil
}
