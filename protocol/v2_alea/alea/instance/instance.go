package instance

import (
	"encoding/hex"
	"encoding/json"
	"sync"

	logging "github.com/ipfs/go-log"
	"go.uber.org/zap"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"

	// specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	// spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	spectypesalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"
)

var logger = logging.Logger("ssv/protocol/qbft/instance").Desugar()

// Instance is a single QBFT instance that starts with a Start call (including a value).
// Every new msg the ProcessMsg function needs to be called
type Instance struct {
	State  *specalea.State
	config alea.IConfig

	processMsgF *spectypesalea.ThreadSafeF
	startOnce   sync.Once
	StartValue  []byte

	logger *zap.Logger
}

func NewInstance(
	config alea.IConfig,
	share *spectypesalea.Share,
	identifier []byte,
	height specalea.Height,
) *Instance {
	msgId := spectypesalea.MessageIDFromBytes(identifier)
	return &Instance{
		State: &specalea.State{
			Share:                share,
			ID:                   identifier,
			Round:                specalea.FirstRound,
			Height:               height,
			LastPreparedRound:    specalea.NoRound,
			ProposeContainer:     specalea.NewMsgContainer(),
			PrepareContainer:     specalea.NewMsgContainer(),
			CommitContainer:      specalea.NewMsgContainer(),
			RoundChangeContainer: specalea.NewMsgContainer(),
		},
		config:      config,
		processMsgF: spectypesalea.NewThreadSafeF(),
		logger: logger.With(zap.String("publicKey", hex.EncodeToString(msgId.GetPubKey())), zap.String("role", msgId.GetRoleType().String()),
			zap.Uint64("height", uint64(height))),
	}
}

// Start is an interface implementation
func (i *Instance) Start(value []byte, height specalea.Height) {
	i.startOnce.Do(func() {
		i.StartValue = value
		i.State.Round = specalea.FirstRound
		i.State.Height = height

		i.config.GetTimer().TimeoutForRound(specalea.FirstRound)

		i.logger.Debug("starting QBFT instance")

	})
}

func (i *Instance) Deliver(proposals []*specalea.ProposalData) int {
	// FIX ME : to be adjusted according to the QBFT implementation
	return 1
}

func (i *Instance) Broadcast(msg *specalea.SignedMessage) error {
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
func (i *Instance) ProcessMsg(msg *specalea.SignedMessage) (decided bool, decidedValue []byte, aggregatedCommit *specalea.SignedMessage, err error) {
	if err := i.BaseMsgValidation(msg); err != nil {
		return false, nil, nil, errors.Wrap(err, "invalid signed message")
	}

	res := i.processMsgF.Run(func() interface{} {
		switch msg.Message.MsgType {
		case specalea.ProposalMsgType:
			return i.uponProposal(msg, i.State.ProposeContainer)
		case specalea.FillGapMsgType:
			return i.uponFillGap(msg, i.State.FillGapContainer)
		case specalea.FillerMsgType:
			return i.uponFiller(msg, i.State.FillerContainer)
		case specalea.ABAInitMsgType:
			return i.uponABAInit(msg)
		case specalea.ABAAuxMsgType:
			return i.uponABAAux(msg)
		case specalea.ABAConfMsgType:
			return i.uponABAConf(msg)
		case specalea.ABAFinishMsgType:
			return i.uponABAFinish(msg)
		case specalea.VCBCSendMsgType:
			return i.uponVCBCSend(msg)
		case specalea.VCBCReadyMsgType:
			return i.uponVCBCReady(msg)
		case specalea.VCBCFinalMsgType:
			return i.uponVCBCFinal(msg)
		default:
			return errors.New("signed message type not supported")
		}
	})
	if res != nil {
		return false, nil, nil, res.(error)
	}
	return i.State.Decided, i.State.DecidedValue, aggregatedCommit, nil
}

func (i *Instance) BaseMsgValidation(msg *specalea.SignedMessage) error {
	if err := msg.Validate(); err != nil {
		return errors.Wrap(err, "invalid signed message")
	}

	if msg.Message.Round < i.State.Round {
		return errors.New("past round")
	}

	switch msg.Message.MsgType {
	case specalea.ProposalMsgType:
		return isValidProposal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.FillGapMsgType:
		return isValidFillGap(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.FillerMsgType:
		return isValidFiller(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.VCBCSendMsgType:
		return isValidVCBCSend(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.VCBCReadyMsgType:
		return isValidVCBCReady(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.VCBCFinalMsgType:
		return isValidVCBCFinal(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.ABAInitMsgType:
		return isValidABAInit(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.ABAAuxMsgType:
		return isValidABAAux(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.ABAConfMsgType:
		return isValidABAConf(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	case specalea.ABAFinishMsgType:
		return isValidABAFinish(i.State, i.config, msg, i.config.GetValueCheckF(), i.State.Share.Committee)
	default:
		return errors.New("signed message type not supported")
	}
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
