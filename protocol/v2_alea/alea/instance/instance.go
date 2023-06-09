package instance

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	// "time"

	"github.com/google/uuid"
	logging "github.com/ipfs/go-log"
	"go.uber.org/zap"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	// specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	// spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"

	spectypesalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"
)

var logger = logging.Logger("ssv/protocol/alea/instance").Desugar()

// func makeTimestamp() int64 {
// 	return time.Now().UnixNano() / int64(time.Microsecond)
// }

// Instance is a single QBFT instance that starts with a Start call (including a value).
// Every new msg the ProcessMsg function needs to be called
type Instance struct {
	State  *messages.State
	config alea.IConfig

	processMsgF *spectypesalea.ThreadSafeF
	startOnce   sync.Once
	StartValue  []byte

	logger    *zap.Logger
	vcbcNum   int
	priority  specalea.Priority
	initTime  int64
	finalTime int64
}

func NewInstance(
	config alea.IConfig,
	share *spectypesalea.Share,
	identifier []byte,
	height specalea.Height,
) *Instance {
	msgId := spectypesalea.MessageIDFromBytes(identifier)

	nodeIDs := make([]types.OperatorID, len(share.Committee))
	for i, op := range share.Committee {
		nodeIDs[i] = types.OperatorID(op.OperatorID)
	}

	inst := &Instance{
		State: &messages.State{
			Share:             share,
			ID:                identifier,
			Round:             specalea.FirstRound,
			Height:            height,
			LastPreparedRound: specalea.NoRound,
			ProposeContainer:  specalea.NewMsgContainer(),
			BatchSize:         1,
			// VCBCState:         specalea.NewVCBCState(),
			FillGapContainer: specalea.NewMsgContainer(),
			FillerContainer:  specalea.NewMsgContainer(),
			AleaDefaultRound: specalea.FirstRound,
			Delivered:        specalea.NewVCBCQueue(),
			StopAgreement:    false,
			// ACState:           specalea.NewACState(),
			VCBCState:         messages.NewVCBCState(nodeIDs),
			ReceivedReadys:    messages.NewReceivedReadys(),
			SentReadys:        messages.NewSentReadys(),
			ACState:           messages.NewACState(nodeIDs),
			FillerMsgReceived: 0,
			StartedCV: false,
			CVState: messages.NewCVState(),
			WaitForVCBCAfterDecided: false,
			WaitForVCBCAfterDecided_Author: types.OperatorID(0),
		},
		priority:    specalea.FirstPriority,
		config:      config,
		processMsgF: spectypesalea.NewThreadSafeF(),
		logger: logger.With(zap.String("publicKey", hex.EncodeToString(msgId.GetPubKey())), zap.String("role", msgId.GetRoleType().String()),
			zap.Uint64("height", uint64(height))),
		initTime:  -1,
		finalTime: -1,
	}
	return inst
}

// Start is an interface implementation
func (i *Instance) Start(value []byte, height specalea.Height) {
	i.startOnce.Do(func() {

		//funciton identifier
		functionID := uuid.New().String()

		// logger
		log := func(str string) {
			i.logger.Debug("$$$$$$ UponStart "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("own operator id", int(i.State.Share.OperatorID)))
		}

		log("start")

		log("set values")
		i.StartValue = value
		i.State.Round = specalea.FirstRound
		i.State.Height = height

		i.vcbcNum = 0

		i.initTime = makeTimestamp()
		log(fmt.Sprintf("i.initTime: %v", i.initTime))

		log("call vcbc")
		i.StartVCBC(value)

		log("finish vcbc")
	})
}

func (i *Instance) Deliver(proposals []*specalea.ProposalData) int {
	// FIX ME : to be adjusted according to the QBFT implementation
	return 1
}

func (i *Instance) Broadcast(msg *messages.SignedMessage) error {
	byts, err := msg.Encode()
	if err != nil {
		return errors.Wrap(err, "could not encode message")
	}

	msgID := spectypesalea.MessageID{}
	copy(msgID[:], msg.Message.Identifier)

	msgToBroadcast := &spectypesalea.SSVMessage{
		MsgType: spectypesalea.SSVConsensusMsgType,
		MsgID:   msgID,
		Data:    byts,
	}
	return i.config.GetNetwork().Broadcast(msgToBroadcast)
}

// ProcessMsg processes a new QBFT msg, returns non nil error on msg processing error
func (i *Instance) ProcessMsg(msg *messages.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *messages.SignedMessage, err error) {
	if err := i.BaseMsgValidation(msg); err != nil {
		return false, nil, nil, errors.Wrap(err, "invalid signed message")
	}

	decided = false
	decidedValue = nil
	aggregatedCommit = nil
	res := i.processMsgF.Run(func() interface{} {
		switch msg.Message.MsgType {
		// case ProposalMsgType:
		// return i.uponProposal(msg, i.State.ProposeContainer)
		// case specalea.FillGapMsgType:
		// 	return i.uponFillGap(msg, i.State.FillGapContainer)
		// case specalea.FillerMsgType:
		// 	return i.uponFiller(msg, i.State.FillerContainer)
		case messages.ABAInitMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponABAInit")
			return i.uponABAInit(msg)
		case messages.ABAAuxMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponABAAux")
			return i.uponABAAux(msg)
		case messages.ABAConfMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponABAConf")
			return i.uponABAConf(msg)
		case messages.ABAFinishMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponABAFinish")
			// _, _, err := i.uponABAFinish(msg)
			// decided, decidedValue, err := i.uponABAFinish(msg)
			// if decided {
			// 	i.State.Decided = decided
			// 	i.State.DecidedValue = decidedValue
			// 	aggregatedCommit = msg
			// }
			// return err
			return i.uponABAFinish(msg)
		case messages.VCBCSendMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponVCBCSend")
			return i.uponVCBCSend(msg)
		case messages.VCBCReadyMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponVCBCReady")
			return i.uponVCBCReady(msg)
		case messages.VCBCFinalMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponVCBCFinal")
			return i.uponVCBCFinal(msg)
		case messages.CVVoteMsgType:
			// i.logger.Debug("$$$$$$ ProcessMsg: uponVCBCFinal")
			return i.uponCVVote(msg)
		default:
			return errors.New("signed message type not supported")
		}
	})
	if res != nil {
		return false, nil, nil, res.(error)
	}
	return i.State.Decided, i.State.DecidedValue, aggregatedCommit, nil
}

func (i *Instance) BaseMsgValidation(msg *messages.SignedMessage) error {

	if err := msg.Validate(); err != nil {
		return errors.Wrap(err, "invalid signed message")
	}

	if msg.Message.Round < i.State.Round {
		return errors.New("past round")
	}

	// switch msg.Message.MsgType {
	// // case messages.ProposalMsgType:
	// // 	return isValidProposal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// // case messages.FillGapMsgType:
	// // 	return isValidFillGap(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// // case messages.FillerMsgType:
	// // 	return isValidFiller(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.VCBCSendMsgType:
	// 	return isValidVCBCSend(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.VCBCReadyMsgType:
	// 	return isValidVCBCReady(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.VCBCFinalMsgType:
	// 	return isValidVCBCFinal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.ABAInitMsgType:
	// 	return isValidABAInit(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.ABAAuxMsgType:
	// 	return isValidABAAux(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.ABAConfMsgType:
	// 	return isValidABAConf(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// case messages.ABAFinishMsgType:
	// 	return isValidABAFinish(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	// 	// default:
	// 	// 	return errors.New("signed message type not supported")
	// }
	return nil
}

// IsDecided interface implementation
func (i *Instance) IsDecided() (bool, []byte) {
	if state := i.State; state != nil {
		return state.Decided, state.DecidedValue
	}
	return false, nil
}

// GetConfig returns the instance config
func (i *Instance) GetConfig() alea.IConfig {
	return i.config
}

// SetConfig returns the instance config
func (i *Instance) SetConfig(config alea.IConfig) {
	i.config = config
}

// GetHeight interface implementation
func (i *Instance) GetHeight() specalea.Height {
	return i.State.Height
}

// GetRoot returns the state's deterministic root
func (i *Instance) GetRoot() ([]byte, error) {
	return i.State.GetRoot()
}

// Encode implementation
func (i *Instance) Encode() ([]byte, error) {
	return json.Marshal(i)
}

// Decode implementation
func (i *Instance) Decode(data []byte) error {
	return json.Unmarshal(data, &i)
}
