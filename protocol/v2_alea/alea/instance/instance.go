package instance

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"sync"

	// "time"

	logging "github.com/ipfs/go-log"
	"go.uber.org/zap"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	spectypesalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/pkg/errors"

	"strings"
)

var logger = logging.Logger("ssv/protocol/alea/instance").Desugar()

type Instance struct {
	State  *messages.State
	config alea.IConfig

	processMsgF *spectypesalea.ThreadSafeF
	startOnce   sync.Once
	StartValue  []byte

	logger *zap.Logger

	priority  specalea.Priority
	initTime  int64
	finalTime int64

	GenerateCPUProfile bool
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

	N := len(share.Committee)
	f := (N - 1) / 3
	share.Quorum = uint64(2*f + 1)
	share.PartialQuorum = uint64(f + 1)

	bls, err := strconv.Atoi(os.Getenv("BLS"))
	if err != nil {
		bls = 1
	}
	blsagg, err := strconv.Atoi(os.Getenv("BLSAGG"))
	if err != nil {
		blsagg = 0
	}
	dh, err := strconv.Atoi(os.Getenv("DH"))
	if err != nil {
		dh = 0
	}
	eddsa, err := strconv.Atoi(os.Getenv("EDDSA"))
	if err != nil {
		eddsa = 0
	}
	rsa, err := strconv.Atoi(os.Getenv("RSA"))
	if err != nil {
		rsa = 0
	}

	bls_f := (bls == 1)
	blsagg_f := (blsagg == 1)
	dh_f := (dh == 1)
	eddsa_f := (eddsa == 1)
	rsa_f := (rsa == 1)

	inst := &Instance{
		State: &messages.State{
			Share:             share,
			ID:                identifier,
			Round:             specalea.FirstRound,
			Height:            height,
			LastPreparedRound: specalea.NoRound,
			CommitContainer:   specalea.NewMsgContainer(),
			DecidedMessage:    nil,
			// Alea State
			VCBCState:       messages.NewVCBCState(nodeIDs),
			ReceivedReadys:  messages.NewReceivedReadys(uint64(share.Quorum)),
			SentReadys:      messages.NewSentReadys(),
			ACState:         messages.NewACState(),
			StartedABA:      false,
			ABASpecialState: messages.NewABASpecialState(len(share.Committee)),
			HasStarted:      false,
			// Flag for case: Aggrement decided one but doesnt' have leader's VCBC
			WaitForVCBCAfterDecided:        false,
			WaitForVCBCAfterDecided_Author: types.OperatorID(0),
			// Common Coin
			CommonCoinContainer: messages.NewPSigContainer(uint64(share.Quorum)),
			CommonCoin:          messages.NewCommonCoin(int64(0)),
			SendCommonCoin:      true,
			// Optimization
			FastABAOptimization:        true,
			WaitVCBCQuorumOptimization: true,
			EqualVCBCOptimization:      true,
			UseBLS:                     bls_f,
			AggregateVerify:            blsagg_f,
			UseDiffieHellman:           dh_f,
			UseEDDSA:                   eddsa_f,
			UseRSA:                     rsa_f,
			// logs
			DecidedLogOnly:     false,
			HideLogs:           false,
			HideValidationLogs: false,
			// Diffie Hellman
			DiffieHellmanContainer:            messages.NewDiffieHellmanContainer(),
			DiffieHellmanContainerOneTimeCost: messages.NewDiffieHellmanContainerOneTimeCost(int(share.OperatorID), nodeIDs),
			// Message Containers and counters
			ReadyContainer:         messages.NewMsgContainer(),
			FinalContainer:         messages.NewMsgContainer(),
			AbaInitContainer:       make(map[specalea.ACRound]map[specalea.Round]*messages.MessageContainer),
			AbaAuxContainer:        make(map[specalea.ACRound]map[specalea.Round]*messages.MessageContainer),
			AbaConfContainer:       make(map[specalea.ACRound]map[specalea.Round]*messages.MessageContainer),
			AbaFinishContainer:     make(map[specalea.ACRound]*messages.MessageContainer),
			CommonCoinMsgContainer: messages.NewMsgContainer(),
			ReadyCounter:           make(map[types.OperatorID]uint64),
			FinalCounter:           make(map[types.OperatorID]uint64),
			AbaInitCounter:         make(map[specalea.ACRound]map[specalea.Round]uint64),
			AbaAuxCounter:          make(map[specalea.ACRound]map[specalea.Round]uint64),
			AbaConfCounter:         make(map[specalea.ACRound]map[specalea.Round]uint64),
			AbaFinishCounter:       make(map[specalea.ACRound]uint64),
			CommonCoinCounter:      0,
			// Log Tags
			VCBCSendLogTag:   0,
			VCBCReadyLogTag:  0,
			VCBCFinalLogTag:  0,
			CommonCoinLogTag: 0,
			AbaInitLogTag:    0,
			AbaAuxLogTag:     0,
			AbaConfLogTag:    0,
			AbaFinishLogTag:  0,
			AbaLogTag:        0,
		},
		priority:    specalea.FirstPriority,
		config:      config,
		processMsgF: spectypesalea.NewThreadSafeF(),
		logger: logger.With(zap.String("publicKey", hex.EncodeToString(msgId.GetPubKey())), zap.String("role", msgId.GetRoleType().String()),
			zap.Uint64("height", uint64(height))),
		initTime:           -1,
		finalTime:          -1,
		GenerateCPUProfile: false,
	}
	return inst
}

func (i *Instance) GetInitTime() int64 {
	return i.initTime
}

func (i *Instance) GetFinalTime() int64 {
	return i.finalTime
}

// Start is an interface implementation
func (i *Instance) Start(value []byte, height specalea.Height) {
	i.startOnce.Do(func() {

		if i.GenerateCPUProfile {
			cpuProfileFile, err := os.Create("profile.out")
			if err != nil {
				panic(err)
			}

			err = pprof.StartCPUProfile(cpuProfileFile)
			if err != nil {
				panic(err)
			}
		}

		i.State.HasStarted = true

		// logger
		log := func(str string) {
			if i.State.HideLogs || (i.State.DecidedLogOnly && !strings.Contains(str, "start")) {
				return
			}
			i.logger.Debug("$$$$$$"+cGreen+" UponStart "+reset+"1: "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("own operator id", int(i.State.Share.OperatorID)))
		}

		i.StartValue = value
		i.State.Round = specalea.FirstRound
		i.State.Height = height

		i.initTime = makeTimestamp()
		log(fmt.Sprintf("i.initTime: %v", i.initTime))

		i.StartVCBC(value)
	})
}

func (i *Instance) SendFinalForDecisionPurpose() {

	acround := i.State.ACState.ACRound
	finishMsg, err := i.CreateABAFinish(byte(1), acround)
	if err != nil {
		panic(err)
	}

	i.Broadcast(finishMsg)

	aba := i.State.ACState.GetABA(acround)
	aba.SetSentFinish(byte(1))
}

func (i *Instance) Decide(value []byte, msg *messages.SignedMessage) {
	i.State.Decided = true
	i.State.DecidedValue = value
	i.State.DecidedMessage = msg

	if i.GenerateCPUProfile {
		pprof.StopCPUProfile()
	}
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

func (i *Instance) GetStatsString() string {
	if i.State.Decided == false {
		return ""
	}
	acround := i.State.ACState.CurrentACRound()

	aba := i.State.ACState.GetABA(acround)
	round := aba.GetRound()
	abaround := aba.GetABARound(round)

	stats := fmt.Sprintf("Stats for H:%v\n\tNumber of Vcbc finals: %v. Has quorum: %v. Authors: %v.\n\tCurrent Agreement round: %v.\n\tCurrent aba round: %v.\n\t\tABA INITs 0 received: %v. 1s received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n\t\tABA AUXs received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n\t\tABA CONFs received: %v. From: %v. HasSent: %v.\n\t\tABA Finish 0 received: %v. Finish 1 received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n",
		i.State.Height,
		i.State.VCBCState.GetLen(), 0, i.State.VCBCState.GetNodeIDs(),
		acround,
		round,
		abaround.LenInit(byte(0)), abaround.LenInit(byte(1)), abaround.GetInit(), abaround.HasSentInit(byte(0)), abaround.HasSentInit(byte(1)),
		abaround.LenAux(), abaround.GetAux(), abaround.HasSentAux(byte(0)), abaround.HasSentAux(byte(1)),
		abaround.LenConf(), abaround.GetConf(), abaround.HasSentConf(),
		aba.LenFinish(byte(0)), aba.LenFinish(byte(1)), aba.GetFinish(), aba.HasSentFinish(byte(0)), aba.HasSentFinish(byte(1)),
	)

	i.logger.Debug(fmt.Sprintf("Instance: GetStats outputting: %v", stats))
	return stats
}
