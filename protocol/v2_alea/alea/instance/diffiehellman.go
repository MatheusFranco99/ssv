package instance

// import (
// 	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
// 	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
// 	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
// 	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
// 	"github.com/google/uuid"

// 	"github.com/pkg/errors"
// 	"go.uber.org/zap"
// )

// func (i *Instance) uponDiffieHellman(msg *messages.SignedMessage) error {

// 	// get data
// 	data, err := msg.Message.GetDiffieHellmanData()
// 	if err != nil {
// 		return errors.Wrap(err, "uponDiffieHellman: could not GetDiffieHellmanData from msg")
// 	}

// 	// sender
// 	senderID := msg.GetSigners()[0]
// 	public_share := data.PublicShare

// 	//funciton identifier
// 	functionID := uuid.New().String()

// 	// logger
// 	log := func(str string) {
// 		i.logger.Debug("$$$$$$UponDiffieHellman "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("sender", int(senderID)))
// 	}

// 	log("start")

// 	if i.State.DiffieHellmanContainer.HasPublicShare(senderID) {
// 		log("already has public share")
// 		return nil
// 	}

// 	i.State.DiffieHellmanContainer.AddPublicShare(senderID, public_share)
// 	log("added")

// 	i.StartVCBC(senderID)
// 	return nil
// }

// func (i *Instance) SendDiffieHellmanPublicShare() {

// 	//funciton identifier
// 	functionID := uuid.New().String()

// 	// logger
// 	log := func(str string) {
// 		i.logger.Debug("$$$$$$ SendDiffieHellman "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
// 	}

// 	log("start")

// 	msg, err := CreateDiffieHelman(i.State, i.config, i.State.DiffieHellmanContainer.GetPublicShare())
// 	if err != nil {
// 		log("failed to create msg")
// 	}
// 	log("created diffie hellman msg")

// 	i.Broadcast(msg)
// 	log("broadcasted diffie hellman public share")
// }

// func isValidDiffieHellman(
// 	state *messages.State,
// 	config alea.IConfig,
// 	signedMsg *messages.SignedMessage,
// 	valCheck specalea.ProposedValueCheckF,
// 	operators []*types.Operator,
// 	logger *zap.Logger,
// ) error {

// 	//funciton identifier
// 	functionID := uuid.New().String()

// 	// logger
// 	log := func(str string) {

// 		if (i.State.DecidedLogOnly) {
// 			return
// 		}
// 		logger.Debug("$$$$$$ UponMV_DiffieHellman "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
// 	}

// 	log("start")

// 	if signedMsg.Message.MsgType != messages.DiffieHellmanMsgType {
// 		return errors.New("msg type is not messages.DiffieHellmanMsgType")
// 	}
// 	log("checked msg type")
// 	if signedMsg.Message.Height != state.Height {
// 		return errors.New("wrong msg height")
// 	}
// 	log("checked height")
// 	if len(signedMsg.GetSigners()) != 1 {
// 		return errors.New("msg allows 1 signer")
// 	}
// 	log("checked signers == 1")
// 	if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
// 		return errors.Wrap(err, "msg signature invalid")
// 	}
// 	log("checked signature")

// 	data, err := signedMsg.Message.GetDiffieHellmanData()
// 	log("got data")
// 	if err != nil {
// 		return errors.Wrap(err, "could not get DiffieHellmanData")
// 	}
// 	if err := data.Validate(); err != nil {
// 		return errors.Wrap(err, "DiffieHellmanData invalid")
// 	}
// 	log("validated")

// 	return nil
// }

// func CreateDiffieHelman(state *messages.State, config alea.IConfig, PublicShare int) (*messages.SignedMessage, error) {
// 	DiffieHellmanData := &messages.DiffieHellmanData{
// 		PublicShare: PublicShare,
// 	}
// 	dataByts, err := DiffieHellmanData.Encode()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "CreateDiffieHelman: could not encode abainit data")
// 	}
// 	msg := &messages.Message{
// 		MsgType:    messages.DiffieHellmanMsgType,
// 		Height:     state.Height,
// 		Round:      state.Round,
// 		Identifier: state.ID,
// 		Data:       dataByts,
// 	}
// 	sig, err := config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "CreateDiffieHelman: failed signing abainit msg")
// 	}

// 	signedMsg := &messages.SignedMessage{
// 		Signature: sig,
// 		Signers:   []types.OperatorID{state.Share.OperatorID},
// 		Message:   msg,
// 	}
// 	return signedMsg, nil
// }