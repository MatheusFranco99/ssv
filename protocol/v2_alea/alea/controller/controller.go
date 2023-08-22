package controller

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/instance"
	logging "github.com/ipfs/go-log"

	"fmt"
	"math"
	"time"
)


const (

    reset = "\033[0m"
    bold = "\033[1m"
    underline = "\033[4m"
    strike = "\033[9m"
    italic = "\033[3m"

    cRed = "\033[31m"
    cGreen = "\033[32m"
    cYellow = "\033[33m"
    cBlue = "\033[34m"
    cPurple = "\033[35m"
    cCyan = "\033[36m"
    cWhite = "\033[37m"
)

var logger = logging.Logger("ssv/protocol/alea/controller").Desugar()

// NewDecidedHandler handles newly saved decided messages.
// it will be called in a new goroutine to avoid concurrency issues
type NewDecidedHandler func(msg *messages.SignedMessage)

// Controller is a alea coordinator responsible for starting and following the entire life cycle of multiple alea InstanceContainer
type Controller struct {
	Identifier []byte
	Height     specalea.Height // incremental Height for InstanceContainer
	// StoredInstances stores the last HistoricalInstanceCapacity in an array for message processing purposes.
	StoredInstances InstanceContainer
	// FutureMsgsContainer holds all msgs from a higher height
	FutureMsgsContainer map[spectypes.OperatorID]specalea.Height // maps msg signer to height of higher height received msgs
	Domain              spectypes.DomainType
	Share               *spectypes.Share
	NewDecidedHandler   NewDecidedHandler `json:"-"`
	config              alea.IConfig
	fullNode            bool
	logger              *zap.Logger
	HeightCountMap		map[int]int
	CurrSlot			int
	Latencies			map[int]map[specalea.Height]int64
	SlotStartTime		map[int]int64
	ThroughputHistogram map[int][]int
	SlotDivision		int
	HeightsStored		map[int][]*instance.Instance
}

func NewController(
	identifier []byte,
	share *spectypes.Share,
	domain spectypes.DomainType,
	config alea.IConfig,
	fullNode bool,
) *Controller {
	msgId := spectypes.MessageIDFromBytes(identifier)
	return &Controller{
		Identifier:          identifier,
		Height:              specalea.FirstHeight,
		Domain:              domain,
		Share:               share,
		StoredInstances:     make(InstanceContainer, 0, InstanceContainerDefaultCapacity),
		FutureMsgsContainer: make(map[spectypes.OperatorID]specalea.Height),
		config:              config,
		fullNode:            fullNode,
		logger: logger.With(zap.String("publicKey", hex.EncodeToString(msgId.GetPubKey())),
			zap.String("role", msgId.GetRoleType().String())),
		HeightCountMap:  make(map[int]int),
		CurrSlot: 0,
		Latencies: make(map[int]map[specalea.Height]int64),
		SlotStartTime: make(map[int]int64),
		ThroughputHistogram: make(map[int][]int),
		SlotDivision: 12,
		HeightsStored: make(map[int][]*instance.Instance),
	}
}

func (c *Controller) ShowStats(slot_value int) {
	if v, ok := c.HeightCountMap[slot_value]; ok {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: Throughput:%v, slot:%v, $$$$$$",v, slot_value))
	} else {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: No Throughput, slot:%v, $$$$$$", slot_value))
	}

	if _,ok := c.Latencies[slot_value]; !ok {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: No Latency list, slot:%v, $$$$$$", slot_value))
	} else {
		latency_lst := make([]int64,len(c.Latencies[slot_value]))
		mean_value := float64(0)
		idx := 0
		for _,v := range c.Latencies[slot_value] {
			mean_value += float64(v)
			latency_lst[idx] = v
			idx += 1
		}
		num_points := len(c.Latencies[slot_value])
		mean_value = mean_value / float64(num_points)
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: Latency mean:%v, num_points:%v, slot:%v, $$$$$$", mean_value, num_points, slot_value))
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: Latency list:%v, slot:%v, $$$$$$", latency_lst, slot_value))
	}

	if _, ok := c.ThroughputHistogram[slot_value]; !ok {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: No Histogram list, slot:%v, $$$$$$", slot_value))
	} else {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: Histogram:%v, slot:%v, $$$$$$", c.ThroughputHistogram[slot_value], slot_value))
	}

	if _, ok := c.HeightsStored[c.CurrSlot]; !ok {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: No HeightsStored list, slot:%v, $$$$$$", slot_value))
	} else {
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: HeightsStored:%v, slot:%v, $$$$$$", c.HeightsStored[slot_value], slot_value))
		
		num_not_decided := 0
		for _,v := range c.HeightsStored[slot_value] {
			if !v.State.Decided {
				num_not_decided += 1
			}
		}
		c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: %vNumber of not decided:%v%v, slot:%v, $$$$$$", cYellow,num_not_decided,reset, slot_value))

		v_idx := 0
		for _,v := range c.HeightsStored[slot_value] {
			if v_idx > 2 {
				break
			}
			if !v.State.Decided {
				v_idx += 1
				i := v
				acround := i.State.ACState.CurrentACRound()

				aba := i.State.ACState.GetABA(acround)
				round := aba.GetRound()
				abaround := aba.GetABARound(round)

				stats := fmt.Sprintf("Stats for H:%v\n\tNumber of Vcbc finals: %v. Authors: %v.\n\tCurrent Agreement round: %v.\n\tCurrent aba round: %v.\n\t\tABA INITs 0 received: %v. 1s received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n\t\tABA AUXs received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n\t\tABA CONFs received: %v. From: %v. HasSent: %v.\n\t\tABA Finish 0 received: %v. Finish 1 received: %v. From: %v. HasSent 0: %v. HasSent 1: %v.\n",	
					i.State.Height,
					i.State.VCBCState.GetLen(), i.State.VCBCState.GetNodeIDs(), 
					acround,
					round,
					abaround.LenInit(byte(0)), abaround.LenInit(byte(1)), abaround.GetInit(), abaround.HasSentInit(byte(0)), abaround.HasSentInit(byte(1)),
					abaround.LenAux(), abaround.GetAux(), abaround.HasSentAux(byte(0)), abaround.HasSentAux(byte(1)),
					abaround.LenConf(), abaround.GetConf(), abaround.HasSentConf(),
					aba.LenFinish(byte(0)), aba.LenFinish(byte(1)), aba.GetFinish(), aba.HasSentFinish(byte(0)), aba.HasSentFinish(byte(1)),
					)
				c.logger.Debug(fmt.Sprintf("$$$$$$ Controller:ShowStats: Instance stats:%v, slot:%v, $$$$$$", stats, slot_value))
			}
		}
	}
	
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

func (c *Controller) SetCurrentSlot(slot_value int) {
	c.CurrSlot = slot_value
	c.SlotStartTime[slot_value] = makeTimestamp()
}



// StartNewInstance will start a new alea instance, if can't will return error
func (c *Controller) StartNewInstance(value []byte) error {
	// if err := c.canStartInstanceForValue(value); err != nil {
	// 	return errors.Wrap(err, "can't start new alea instance")
	// }

	// only if current height's instance exists (and decided since passed can start instance) bump
	if c.StoredInstances.FindInstance(c.Height) != nil {
		c.bumpHeight()
	}

	newInstance := c.addAndStoreNewInstance()
	newInstance.Start(value, c.Height)

	if _, ok := c.HeightsStored[c.CurrSlot]; !ok {
		c.HeightsStored[c.CurrSlot] = make([]*instance.Instance,0)
	}
	c.HeightsStored[c.CurrSlot] = append(c.HeightsStored[c.CurrSlot],newInstance)

	return nil
}

// ProcessMsg processes a new msg, returns decided message or error
func (c *Controller) ProcessMsg(msg *messages.SignedMessage) (*messages.SignedMessage, []byte, error) {
	if err := c.BaseMsgValidation(msg); err != nil {
		return nil, nil, errors.Wrap(err, "invalid msg")
	}

	/**
	Main controller processing flow
	_______________________________
	All decided msgs are processed the same, out of instance
	All valid future msgs are saved in a container and can trigger highest decided futuremsg
	All other msgs (not future or decided) are processed normally by an existing instance (if found)
	*/
	if IsDecidedMsg(c.Share, msg) {
		return c.UponDecided(msg)
	} else if msg.Message.Height > c.Height {
		return c.UponFutureMsg(msg)
	} else {
		return c.UponExistingInstanceMsg(msg)
	}
}

func (c *Controller) UponExistingInstanceMsg(msg *messages.SignedMessage) (*messages.SignedMessage, []byte, error) {
	inst := c.InstanceForHeight(msg.Message.Height)
	if inst == nil {
		return nil, nil, errors.New("instance not found")
	}

	prevDecided, _ := inst.IsDecided()

	decided, decideValue, decidedMsg, err := inst.ProcessMsg(msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not process msg")
	}

	// save the highest Decided
	if !decided {
		return nil, nil, nil
	}

	// ProcessMsg returns a nil decidedMsg when given a non-commit message
	// while the instance is decided. In this case, we have nothing new to broadcast.
	// if decidedMsg == nil {
	// 	return nil, nil
	// }

	// if err := c.broadcastDecided(decidedMsg); err != nil {
	// 	// no need to fail processing instance deciding if failed to save/ broadcast
	// 	c.logger.Debug("failed to broadcast decided message", zap.Error(err))
	// }

	if prevDecided {
		return nil, nil, err
	}

	if _, ok := c.HeightCountMap[c.CurrSlot]; !ok {
		c.HeightCountMap[c.CurrSlot] = 0
	}
	c.HeightCountMap[c.CurrSlot] += 1

	if _, ok := c.Latencies[c.CurrSlot]; !ok {
		c.Latencies[c.CurrSlot] = make(map[specalea.Height]int64)
	}
	c.Latencies[c.CurrSlot][inst.State.Height] = (inst.GetFinalTime() - inst.GetInitTime())

	if _, ok := c.ThroughputHistogram[c.CurrSlot]; !ok {
		c.ThroughputHistogram[c.CurrSlot] = make([]int,c.SlotDivision)
	}
	slot_init_time := c.SlotStartTime[c.CurrSlot]//c.config.GetNetwork().GetSlotStartTime()
	diff := inst.GetFinalTime()-slot_init_time
	dividor := int64(12_000_000) / int64(c.SlotDivision)
	c.ThroughputHistogram[c.CurrSlot][int(math.Floor( float64(diff) / float64(dividor)  ))] += 1

	return decidedMsg, decideValue, nil
}

// BaseMsgValidation returns error if msg is invalid (base validation)
func (c *Controller) BaseMsgValidation(msg *messages.SignedMessage) error {
	// verify msg belongs to controller
	if !bytes.Equal(c.Identifier, msg.Message.Identifier) {
		return errors.New("message doesn't belong to Identifier")
	}

	return nil
}

func (c *Controller) InstanceForHeight(height specalea.Height) *instance.Instance {
	// Search in memory.
	if inst := c.StoredInstances.FindInstance(height); inst != nil {
		return inst
	}

	// Search in storage, if full node.
	if !c.fullNode {
		return nil
	}
	storedInst, err := c.config.GetStorage().GetInstance(c.Identifier, height)
	if err != nil {
		c.logger.Debug("could not load instance from storage",
			zap.Uint64("height", uint64(height)),
			zap.Uint64("ctrl_height", uint64(c.Height)),
			zap.Error(err))
		return nil
	}
	if storedInst == nil {
		return nil
	}
	inst := instance.NewInstance(c.config, c.Share, c.Identifier, storedInst.State.Height)
	inst.State = storedInst.State
	return inst
}

func (c *Controller) bumpHeight() {
	c.Height++
}

// GetIdentifier returns alea Identifier, used to identify messages
func (c *Controller) GetIdentifier() []byte {
	return c.Identifier
}

// addAndStoreNewInstance returns creates a new alea instance, stores it in an array and returns it
func (c *Controller) addAndStoreNewInstance() *instance.Instance {
	i := instance.NewInstance(c.GetConfig(), c.Share, c.Identifier, c.Height)
	c.StoredInstances.addNewInstance(i)
	return i
}

func (c *Controller) canStartInstanceForValue(value []byte) error {
	// check value
	if err := c.GetConfig().GetValueCheckF()(value); err != nil {
		return errors.Wrap(err, "value invalid")
	}

	return c.CanStartInstance()
}

// CanStartInstance returns nil if controller can start a new instance
func (c *Controller) CanStartInstance() error {
	// check prev instance if prev instance is not the first instance
	inst := c.StoredInstances.FindInstance(c.Height)
	if inst == nil {
		return nil
	}
	// if decided, _ := inst.IsDecided(); !decided {
	// 	return errors.New("previous instance hasn't Decided")
	// }

	return nil
}

// GetRoot returns the state's deterministic root
func (c *Controller) GetRoot() ([]byte, error) {
	marshaledRoot, err := json.Marshal(c)
	if err != nil {
		return nil, errors.Wrap(err, "could not encode state")
	}
	ret := sha256.Sum256(marshaledRoot)
	return ret[:], nil
}

// Encode implementation
func (c *Controller) Encode() ([]byte, error) {
	return json.Marshal(c)
}

// Decode implementation
func (c *Controller) Decode(data []byte) error {
	err := json.Unmarshal(data, &c)
	if err != nil {
		return errors.Wrap(err, "could not decode controller")
	}

	config := c.GetConfig()
	for _, i := range c.StoredInstances {
		if i != nil {
			// TODO-spec-align changed due to instance and controller are not in same package as in spec, do we still need it for test?
			i.SetConfig(config)
		}
	}
	return nil
}

func (c *Controller) broadcastDecided(aggregatedCommit *messages.SignedMessage) error {
	// Broadcast Decided msg
	byts, err := aggregatedCommit.Encode()
	if err != nil {
		return errors.Wrap(err, "could not encode decided message")
	}

	msgToBroadcast := &spectypes.SSVMessage{
		MsgType: spectypes.SSVConsensusMsgType,
		MsgID:   specalea.ControllerIdToMessageID(c.Identifier),
		Data:    byts,
	}
	if err := c.GetConfig().GetNetwork().Broadcast(msgToBroadcast); err != nil {
		// We do not return error here, just Log broadcasting error.
		return errors.Wrap(err, "could not broadcast decided")
	}
	return nil
}

func (c *Controller) GetConfig() alea.IConfig {
	return c.config
}
