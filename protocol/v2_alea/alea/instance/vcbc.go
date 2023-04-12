package instance

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"

	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"go.uber.org/zap"
)

func (i *Instance) StartVCBC(data []byte) error {

	//funciton identifier
	functionID := uuid.New().String()
	priority := i.priority
	vcbcNum := i.vcbcNum

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponVCBCStart "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("priority", int(priority)), zap.Int("vcbcNum", int(vcbcNum)))
	}

	log("start")

	author := i.State.Share.OperatorID

	log("create vcbc send")
	// create VCBCSend message and broadcasts
	msgToBroadcast, err := CreateVCBCSend(i.State, i.config, data, i.priority, author)
	if err != nil {
		return errors.Wrap(err, "StartVCBC: failed to create VCBCSend message")
	}

	i.priority += 1

	log("broadcast start")
	i.Broadcast(msgToBroadcast)
	log("broadcast finish")

	i.vcbcNum += 1

	log("finish")

	return nil
}

// func (i *Instance) AddOwnVCBCReady(proposals []*specalea.ProposalData, priorioty specalea.Priority) error {

// 	hash, err := GetProposalsHash(proposals)
// 	if err != nil {
// 		return errors.Wrap(err, "AddOwnVCBCReady: could not get hash of proposals")
// 	}
// 	// create VCBCReady message with proof
// 	vcbcReadyMsg, err := CreateVCBCReady(i.State, i.config, hash, priorioty, i.State.Share.OperatorID)
// 	if err != nil {
// 		return errors.Wrap(err, "AddOwnVCBCReady: failed to create VCBCReady message with proof")
// 	}
// 	i.uponVCBCReady(vcbcReadyMsg)
// 	return nil
// }

// func (i *Instance) AddVCBCOutput(proposals []*specalea.ProposalData, priority specalea.Priority, author types.OperatorID) {

// 	// initializes queue of the author if it doesn't exist
// 	if _, exists := i.State.VCBCState.Queues[author]; !exists {
// 		i.State.VCBCState.Queues[author] = specalea.NewVCBCQueue()
// 	}

// 	// gets the sender's associated queue
// 	queue := i.State.VCBCState.Queues[author]

// 	// check if it was already delivered
// 	if i.State.Delivered.HasProposalList(proposals) {
// 		return
// 	}

// 	// check if queue alreasy has proposals and priority
// 	if queue.HasPriority(priority) {
// 		return
// 	}

// 	// store proposals and priorioty value
// 	queue.Enqueue(proposals, priority)
// }
