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

	"strings"
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
			VCBCState:         messages.NewVCBCState(nodeIDs),
			ReceivedReadys:    messages.NewReceivedReadys(uint64(share.Quorum)),
			SentReadys:        messages.NewSentReadys(),
			ACState:           messages.NewACState(),
			CommitContainer:      specalea.NewMsgContainer(),
			StartedABA: false,
			WaitForVCBCAfterDecided: false,
			WaitForVCBCAfterDecided_Author: types.OperatorID(0),
			CommonCoinContainer: messages.NewPSigContainer(uint64(share.Quorum)),
			CommonCoin: messages.NewCommonCoin(int64(0)),
			ABASpecialState: messages.NewABASpecialState(len(share.Committee)),
			FastABAOptimization: false,
			WaitVCBCQuorumOptimization: false,
			EqualVCBCOptimization: false,
			DecidedMessage: nil,
			DecidedLogOnly: true,
			SendCommonCoin: true,
			HasStarted: false,
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

		i.State.HasStarted = true

		//funciton identifier
		functionID := uuid.New().String()

		// logger
		log := func(str string) {
			if (i.State.DecidedLogOnly && !strings.Contains(str,"start")) {
				return
			}
			i.logger.Debug("$$$$$$ UponStart "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("own operator id", int(i.State.Share.OperatorID)))
		}

		log("starting alea instance")

		i.StartValue = value
		i.State.Round = specalea.FirstRound
		i.State.Height = height

		i.vcbcNum = 0
		log("set values")

		i.initTime = makeTimestamp()
		log(fmt.Sprintf("i.initTime: %v", i.initTime))


		i.StartVCBC(value)
	})
}



func (i *Instance) Decide(value []byte, msg *messages.SignedMessage) {
	i.State.Decided = true
	i.State.DecidedValue = value
	i.State.DecidedMessage = msg
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
	
	if !i.State.HasStarted {
		return false,nil,nil,nil
	}

	// special treatment to ready msg
	if msg.Message.MsgType == messages.VCBCReadyMsgType {

		// get Data
		vcbcReadyData, err := msg.Message.GetVCBCReadyData()
		if err != nil {
			return false, nil, nil, errors.Wrap(err, "UponProcessMsg: could not get vcbcReadyData data from signedMessage")
		}

		// get attributes
		author := vcbcReadyData.Author
		if author != i.State.Share.OperatorID {
			return false, nil, nil, nil
		}
	}

	if err := i.BaseMsgValidation(msg); err != nil {
		return false, nil, nil, errors.Wrap(err, "invalid signed message")
	}

	decided = false
	decidedValue = nil
	aggregatedCommit = nil
	res := i.processMsgF.Run(func() interface{} {
		switch msg.Message.MsgType {
		case messages.ABAInitMsgType:
			return i.uponABAInit(msg)
		case messages.ABAAuxMsgType:
			return i.uponABAAux(msg)
		case messages.ABAConfMsgType:
			return i.uponABAConf(msg)
		case messages.ABAFinishMsgType:
			return i.uponABAFinish(msg)
		case messages.VCBCSendMsgType:
			return i.uponVCBCSend(msg)
		case messages.VCBCReadyMsgType:
			return i.uponVCBCReady(msg)
		case messages.VCBCFinalMsgType:
			return i.uponVCBCFinal(msg)
		case messages.CommonCoinMsgType:
			return i.uponCommonCoin(msg)
		default:
			return errors.New("signed message type not supported")
		}
	})
	if res != nil {
		return false, nil, nil, res.(error)
	}

	decided = i.State.Decided
	decidedValue = i.State.DecidedValue
	aggregatedCommit = i.State.DecidedMessage
	return i.State.Decided, i.State.DecidedValue, i.State.DecidedMessage, nil
}

func (i *Instance) BaseMsgValidation(msg *messages.SignedMessage) error {

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		return
		i.logger.Debug("$$$$$$ UponMessageValidation "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if err := msg.Validate(); err != nil {
		return errors.Wrap(err, "invalid signed message")
	}

	if msg.Message.Round < i.State.Round {
		return errors.New("past round")
	}

	var err error
	switch msg.Message.MsgType {
	case messages.VCBCSendMsgType:
		err = isValidVCBCSend(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.VCBCReadyMsgType:
		err = isValidVCBCReady(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.VCBCFinalMsgType:
		err = isValidVCBCFinal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAInitMsgType:
		err = isValidABAInit(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAAuxMsgType:
		err = isValidABAAux(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAConfMsgType:
		err = isValidABAConf(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.ABAFinishMsgType:
		err = isValidABAFinish(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	case messages.CommonCoinMsgType:
		err = isValidCommonCoin(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee, i.logger)
	default:
		err = errors.New(fmt.Sprintf("signed message type not supported: %v",msg.Message.MsgType))
	}

	log("finish")
	return err
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
