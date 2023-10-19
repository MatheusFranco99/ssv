package validator

import (
	"context"
	"encoding/hex"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/message"

	"fmt"
	"time"
	"os"
	"strconv"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	specssv "github.com/MatheusFranco99/ssv-spec-AleaBFT/ssv"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/google/uuid"
	logging "github.com/ipfs/go-log"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/MatheusFranco99/ssv/ibft/storage"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/queue"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/runner"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
)

var logger = logging.Logger("ssv/protocol/ssv/validator").Desugar()

// Validator represents an SSV ETH consensus validator Share assigned, coordinates duty execution and more.
// Every validator has a validatorID which is validator's public key.
// Each validator has multiple DutyRunners, for each duty type.
type Validator struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger

	DutyRunners runner.DutyRunners
	Network     specalea.Network
	Beacon      specssv.BeaconNode
	Share       *types.SSVShare
	Signer      spectypes.KeyManager

	Storage *storage.ALEAStores
	Queues  map[spectypes.BeaconRole]queueContainer

	state uint32

	// test variable -> do only one duty per slot
	DoneDutyForSlot map[int]bool
	// increment system load
	SystemLoad int
}

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	underline = "\033[4m"
	strike    = "\033[9m"
	italic    = "\033[3m"

	cRed    = "\033[31m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
	cBlue   = "\033[34m"
	cPurple = "\033[35m"
	cCyan   = "\033[36m"
	cWhite  = "\033[37m"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

// NewValidator creates a new instance of Validator.
func NewValidator(pctx context.Context, cancel func(), options Options) *Validator {
	options.defaults()

	logger := logger.With(zap.String("validator", hex.EncodeToString(options.SSVShare.ValidatorPubKey)))

	v := &Validator{
		ctx:             pctx,
		cancel:          cancel,
		logger:          logger,
		DutyRunners:     options.DutyRunners,
		Network:         options.Network,
		Beacon:          options.Beacon,
		Storage:         options.Storage,
		Share:           options.SSVShare,
		Signer:          options.Signer,
		Queues:          make(map[spectypes.BeaconRole]queueContainer),
		state:           uint32(NotStarted),
		DoneDutyForSlot: make(map[int]bool),
		SystemLoad:      0,
	}

	for _, dutyRunner := range options.DutyRunners {
		// set timeout F
		dutyRunner.GetBaseRunner().TimeoutF = v.onTimeout
		v.Queues[dutyRunner.GetBaseRunner().BeaconRoleType] = queueContainer{
			Q: queue.New(),
			queueState: &queue.State{
				HasRunningInstance: false,
				Height:             0,
				Slot:               0,
				//Quorum:             options.SSVShare.Share,// TODO
			},
		}
	}

	return v
}

// StartDuty starts a duty for the validator
func (v *Validator) StartDuty(duty *spectypes.Duty) error {
	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		v.logger.Debug("$$$$$$ UponValidatorStartDuty "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	if duty.Type.String() != "SYNC_COMMITTEE" {
		// log("duty not sync committee. Quitting.")
		return nil
	}

	slot := duty.Slot
	if _, ok := v.DoneDutyForSlot[int(slot)]; ok {

		log(fmt.Sprintf("already did duty in this slot. %v not going to start. Quitting.", duty.Type.String()))
		return nil
	} else {
		v.DutyRunners[duty.Type].GetBaseRunner().QBFTController.ShowStats(int(slot) - 1)
		v.DutyRunners[duty.Type].GetBaseRunner().QBFTController.SetCurrentSlot(int(slot))
	}

	dutyRunner := v.DutyRunners[duty.Type]
	if dutyRunner == nil {
		return errors.Errorf("duty type %s not supported", duty.Type.String())
	}

	log(fmt.Sprintf("Setting true for %vslot %v,%v due to duty %v", cBlue, int(slot), reset, duty.Type.String()))
	v.DoneDutyForSlot[int(slot)] = true
	if v.SystemLoad == 0 {
		sload,err := strconv.Atoi(os.Getenv("SLOAD"))
		if err != nil {
			sload = 1
		}
		v.SystemLoad = sload
	} else {
		// panic("QUITING")
		log("TERMINATING")
		os.Exit(0)
		if v.SystemLoad == 1 {
			v.SystemLoad = 0
		}
		v.SystemLoad += 20
	}
	// if v.SystemLoad == 20 {
	// 	log("TERMINATING")
	// 	os.Exit(0)
	// }
	log(fmt.Sprintf("%vSystem load: %v %v", cYellow, v.SystemLoad, reset))

	dutyRunner.SetSystemLoad(v.SystemLoad)

	v.Queues[dutyRunner.GetBaseRunner().BeaconRoleType].ClearQ()

	return dutyRunner.StartNewDuty(duty)
}

// ProcessMessage processes Network Message of all types
func (v *Validator) ProcessMessage(msg *queue.DecodedSSVMessage) error {
	dutyRunner := v.DutyRunners.DutyRunnerForMsgID(msg.GetID())
	if dutyRunner == nil {
		return errors.Errorf("could not get duty runner for msg ID")
	}

	if err := validateMessage(v.Share.Share, msg.SSVMessage); err != nil {
		return errors.Wrap(err, "Message invalid")
	}

	switch msg.GetType() {
	case spectypes.SSVConsensusMsgType:
		signedMsg, ok := msg.Body.(*messages.SignedMessage)
		if !ok {
			return errors.New("could not decode consensus message from network message")
		}
		return dutyRunner.ProcessConsensus(signedMsg)
	case spectypes.SSVPartialSignatureMsgType:
		signedMsg, ok := msg.Body.(*specssv.SignedPartialSignatureMessage)
		if !ok {
			return errors.New("could not decode post consensus message from network message")
		}
		if signedMsg.Message.Type == specssv.PostConsensusPartialSig {
			return dutyRunner.ProcessPostConsensus(signedMsg)
		}
		return dutyRunner.ProcessPreConsensus(signedMsg)
	case message.SSVEventMsgType:
		return v.handleEventMessage(msg, dutyRunner)
	default:
		return errors.New("unknown msg")
	}
}

func validateMessage(share spectypes.Share, msg *spectypes.SSVMessage) error {
	if !share.ValidatorPubKey.MessageIDBelongs(msg.GetID()) {
		return errors.New("msg ID doesn't match validator ID")
	}

	if len(msg.GetData()) == 0 {
		return errors.New("msg data is invalid")
	}

	return nil
}
