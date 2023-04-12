package instance

import (
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	logging "github.com/ipfs/go-log"
	"go.uber.org/zap"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

var logger = logging.Logger("ssv/protocol/alea/instance").Desugar()

// Instance is a single alea instance that starts with a Start call (including a value).
// Every new msg the ProcessMsg function needs to be called
type Instance struct {
	State  *specalea.State
	config alea.IConfig

	processMsgF *spectypes.ThreadSafeF
	startOnce   sync.Once
	StartValue  []byte

	logger    *zap.Logger
	initTime  int64
	finalTime int64
}

func NewInstance(
	config alea.IConfig,
	share *spectypes.Share,
	identifier []byte,
	height specalea.Height,
) *Instance {
	msgId := spectypes.MessageIDFromBytes(identifier)
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
		processMsgF: spectypes.NewThreadSafeF(),
		logger: logger.With(zap.String("publicKey", hex.EncodeToString(msgId.GetPubKey())), zap.String("role", msgId.GetRoleType().String()),
			zap.Uint64("height", uint64(height))),
		initTime:  -1,
		finalTime: -1,
	}
}

// Start is an interface implementation
func (i *Instance) Start(value []byte, height specalea.Height) {
	i.startOnce.Do(func() {
		i.StartValue = value
		i.State.Round = specalea.FirstRound
		i.State.Height = height

		i.initTime = makeTimestamp()

		i.config.GetTimer().TimeoutForRound(specalea.FirstRound)

		//funciton identifier
		functionID := uuid.New().String()

		// logger
		log := func(str string) {
			i.logger.Debug("$$$$$$ UponStart "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
		}

		log("start alea instance")

		// i.logger.Debug("$$$$$$ starting alea instance. time(micro):", zap.Int64("time(micro)", makeTimestamp()))

		// propose if this node is the proposer
		if proposer(i.State, i.GetConfig(), specalea.FirstRound) == i.State.Share.OperatorID {
			log("create proposal")

			proposal, err := CreateProposal(i.State, i.config, i.StartValue, nil, nil)
			// nolint
			if err != nil {
				i.logger.Warn("failed to create proposal", zap.Error(err))
				// TODO align spec to add else to avoid broadcast errored proposal
			} else {
				// nolint
				log("broadcast start")

				if err := i.Broadcast(proposal); err != nil {
					i.logger.Warn("failed to broadcast proposal", zap.Error(err))
				}
			}
		}
		log("finish")

	})
}

func (i *Instance) Broadcast(msg *messages.SignedMessage) error {
	byts, err := msg.Encode()
	if err != nil {
		return errors.Wrap(err, "could not encode message")
	}

	msgID := spectypes.MessageID{}
	copy(msgID[:], msg.Message.Identifier)

	msgToBroadcast := &spectypes.SSVMessage{
		MsgType: spectypes.SSVConsensusMsgType,
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

	res := i.processMsgF.Run(func() interface{} {
		switch msg.Message.MsgType {
		case specalea.ProposalMsgType:
			return i.uponProposal(msg, i.State.ProposeContainer)
		case specalea.PrepareMsgType:
			return i.uponPrepare(msg, i.State.PrepareContainer, i.State.CommitContainer)
		case specalea.CommitMsgType:
			decided, decidedValue, aggregatedCommit, err = i.UponCommit(msg, i.State.CommitContainer)
			if decided {
				i.State.Decided = decided
				i.State.DecidedValue = decidedValue
				// i.logger.Debug("$$$$$$ Decided on value with commit.", zap.Int64("time(micro)", makeTimestamp()))
			}
			return err
		case specalea.RoundChangeMsgType:
			return i.uponRoundChange(i.StartValue, msg, i.State.RoundChangeContainer, i.config.GetValueCheckF())
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

	switch msg.Message.MsgType {
	case specalea.ProposalMsgType:
		return isValidProposal(
			i.State,
			i.config,
			msg,
			i.config.GetValueCheckF(),
			i.State.Share.Committee,
		)
	case specalea.PrepareMsgType:
		proposedMsg := i.State.ProposalAcceptedForCurrentRound
		if proposedMsg == nil {
			return errors.New("did not receive proposal for this round")
		}
		acceptedProposalData, err := proposedMsg.Message.GetCommitData()
		if err != nil {
			return errors.Wrap(err, "could not get accepted proposal data")
		}
		return validSignedPrepareForHeightRoundAndValue(
			i.config,
			msg,
			i.State.Height,
			i.State.Round,
			acceptedProposalData.Data,
			i.State.Share.Committee,
		)
	case specalea.CommitMsgType:
		proposedMsg := i.State.ProposalAcceptedForCurrentRound
		if proposedMsg == nil {
			return errors.New("did not receive proposal for this round")
		}
		return validateCommit(
			i.config,
			msg,
			i.State.Height,
			i.State.Round,
			i.State.ProposalAcceptedForCurrentRound,
			i.State.Share.Committee,
		)
	case specalea.RoundChangeMsgType:
		return validRoundChange(i.State, i.config, msg, i.State.Height, msg.Message.Round)
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
