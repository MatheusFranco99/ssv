package instance

import (
	"fmt"
	// "crypto/sha256"
	// "encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"github.com/google/uuid"
)

// uponCommonCoin process proposal message
// Assumes message is valid!
func (i *Instance) uponCommonCoin(signedMessage *messages.SignedMessage) error {

	// get Data
	commonCoinData, err := signedMessage.Message.GetCommonCoinData()
	if err != nil {
		return errors.Wrap(err, "uponProposal: could not get proposal data from signedProposal")
	}

	shareSig := commonCoinData.ShareSign
	senderID := signedMessage.GetSigners()[0]


	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponCommonCoinData "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)))
	}

	log("start")

	if i.initTime == -1 {
		i.initTime = makeTimestamp()
	}


	if i.State.CommonCoin.HasSeed() {
		log("has seed. quitting.")
		return nil
	}

	i.State.CommonCoinContainer.AddSignature(senderID,shareSig)
	log("added signature")


	if i.State.CommonCoinContainer.HasQuorum() {
		log("got quorum")

		root, err := i.GetCommonCoinRoot()
		if err != nil {
			return errors.Wrap(err,"UponCommonCoin: could not compute common coin root")
		}
		log("recalculated root")

		signature, err := i.State.CommonCoinContainer.ReconstructSignature(root.Value,i.State.Share.ValidatorPubKey)
		if err != nil {
			return errors.Wrap(err,"UponCommonCoin: error reconstructing signature")
		}
		log(fmt.Sprintf("generated threshold signature: %v",signature))
		

		var value int64
		for i := 0; i < 5; i++ {
			value = (value << 8) + int64(root.Value[i])
		}
		

		i.State.CommonCoin.SetSeed(value)
		log("setted seed")
	}

	log("finish")
	return nil
}


func isValidCommonCoin(
	state *messages.State,
	config alea.IConfig,
	signedMessage *messages.SignedMessage,
	valCheck specalea.ProposedValueCheckF,
	operators []*types.Operator,
	logger *zap.Logger,
) error {
	if signedMessage.Message.MsgType != messages.CommonCoinMsgType {
		return errors.New("msg type is not common coin")
	}
	if signedMessage.Message.Height != state.Height {
		return errors.New("wrong msg height")
	}
	if len(signedMessage.GetSigners()) != 1 {
		return errors.New("msg allows 1 signer")
	}
	if err := signedMessage.Signature.VerifyByOperators(signedMessage, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}

	msgData, err := signedMessage.Message.GetCommonCoinData()
	if err != nil {
		return errors.Wrap(err, "could not get common coin data")
	}
	if err := msgData.Validate(); err != nil {
		return errors.Wrap(err, "common coin data invalid")
	}

	return nil
}

func (i *Instance) SendCommonCoinShare() error {

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponSendCommonCoinShare "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	shareSign, err := i.GetCommonCoinShare()
	if err != nil {
		return err
	}
	log("get common coin share")

	msg, err := CreateCommonCoin(i.State, i.config, shareSign)
	if err != nil {
		return errors.Wrap(err, "SendCommonCoinShare: failed to create common coin message")
	}
	log("created common coin message")

	i.Broadcast(msg)
	log("broadcasted")

	return nil
}

func (i *Instance) GetCommonCoinRoot() (*messages.ByteRoot, error)  {

	data := fmt.Sprintf("AleaCommonCoin%v%v",i.State.ID, i.State.Height)

	root, err := types.ComputeSigningRoot(messages.NewByteRoot([]byte(data)), types.ComputeSignatureDomain(i.config.GetSignatureDomainType(), types.QBFTSignatureType))
	if err != nil {
		return messages.NewByteRoot([]byte{}), errors.Wrap(err, "GetCommonCoinShare: could not compute signing root")
	}
	return messages.NewByteRoot(root), nil
}

func (i *Instance) GetCommonCoinShare() (types.Signature, error)  {

	root, err := i.GetCommonCoinRoot()
	if err != nil {
		return nil, err
	}

	return i.config.GetSigner().SignRoot(root ,types.QBFTSignatureType, i.State.Share.SharePubKey)
}


// CreateCommonCoin
func CreateCommonCoin(state *messages.State, config alea.IConfig, shareSign types.Signature) (*messages.SignedMessage, error) {
	msgData := &messages.CommonCoinData{
		ShareSign: shareSign,
	}
	dataByts, err := msgData.Encode()
	if err != nil {
		return nil, errors.Wrap(err, "could not encode common coin data")
	}
	msg := &messages.Message{
		MsgType:    messages.CommonCoinMsgType,
		Height:     state.Height,
		Round:      state.Round,
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
