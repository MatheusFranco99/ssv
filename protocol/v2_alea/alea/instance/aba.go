package instance

import (
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/pkg/errors"
	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
)

func (i *Instance) StartAlea() error {


	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponStartAlea "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("ACRound", int(i.State.ACState.ACRound)))
	}
	

	log("start")

	acround := int(i.State.ACState.ACRound)
	opIDList := make([]types.OperatorID, len(i.State.Share.Committee))
	for idx, op := range i.State.Share.Committee {
		opIDList[idx] = op.OperatorID
	}
	leader := opIDList[(acround)%len(opIDList)]
	// i.config.GetProposerF()(i.State, specalea.Round(i.State.ACState.ACRound))

	log(fmt.Sprintf("leader: %v",int(leader)))

	vote := byte(0)
	if i.State.VCBCState.HasData(leader) {
		vote = byte(1)
	}

	log(fmt.Sprintf("vote: %v",int(vote)))


	// start ABA protocol
	err := i.StartABA(vote,leader)
	if err != nil {
		return errors.Wrap(err, "failed to start ABA and get result")
	}
	return nil
}



// import (
// 	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
// 	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"

// 	"fmt"
// 	"time"

// 	"github.com/pkg/errors"
// 	"go.uber.org/zap"
// )

// func (i *Instance) StartAgreementComponent() error {

// 	for {

// 		i.logger.Debug("$$$$$$ AgreementComponent start round", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)))

// 		// check if it should stop performing agreement
// 		if i.State.StopAgreement {
// 			break
// 		}

// 		// calculate the round leader (to get value to be decided on)
// 		// leader := i.config.GetProposerF()(i.State, specalea.Round(i.State.ACState.ACRound))
// 		leader := types.OperatorID(2)

// 		// get the local queue associated with the leader's id (create if there isn't one)
// 		if _, exists := i.State.VCBCState.Queues[leader]; !exists {
// 			i.State.VCBCState.Queues[leader] = specalea.NewVCBCQueue()
// 		}
// 		queue := i.State.VCBCState.Queues[leader]

// 		// get the value of the queue with the lowest priority value
// 		value, priority := queue.Peek()
// 		// value := nil
// 		// priority := alea.Priority(0)
// 		// for value == nil {
// 		// 	queue := i.State.VCBCState.Queues[leader]
// 		// 	value, priority = queue.Peek()
// 		// }
// 		for value == nil {
// 			value, priority = i.State.VCBCState.Queues[leader].Peek()

// 			if value != nil {
// 				// Do something with the value
// 				fmt.Println("Received value:", value)
// 				break // exit the loop
// 			} else {
// 				// No value in queue yet, wait for a short duration before checking again
// 				time.Sleep(time.Millisecond * 100)
// 			}
// 		}

// 		// decide own vote
// 		vote := byte(0)
// 		if value == nil {
// 			vote = byte(0)
// 		} else {
// 			vote = byte(1)
// 		}

// 		i.logger.Debug("$$$$$$ AgreementComponent round info", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)), zap.Int("vote", int(vote)), zap.Int("leader", int(leader)), zap.Int("priority", int(priority)))

// 		// start ABA protocol
// 		result, err := i.StartABA(vote)
// 		if err != nil {
// 			return errors.Wrap(err, "failed to start ABA and get result")
// 		}

// 		i.logger.Debug("$$$$$$ AgreementComponent aba result", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)), zap.Int("abaresult", int(result)))
// 		if i.State.StopAgreement {
// 			break
// 		}

// 		if result == 1 {
// 			// if the protocol agreed on the value of the leader replica, deliver it

// 			// if ABA decided 1 but own vote was 0, start recover mechanism to get VCBC messages not received from leader
// 			if vote == 0 {
// 				if !i.State.VCBCState.HasM(leader, priority) {
// 					// create FILLGAP message
// 					fillerContLen := i.State.FillerContainer.Len(i.State.AleaDefaultRound)
// 					fillGapMsg, err := CreateFillGap(i.State, i.config, leader, priority)
// 					if err != nil {
// 						return errors.Wrap(err, "StartAgreementComponent: failed to create FillGap message")
// 					}
// 					i.logger.Debug("$$$$$$ AgreementComponent fill gap broadcast start", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)))

// 					i.Broadcast(fillGapMsg)
// 					i.logger.Debug("$$$$$$ AgreementComponent fill gap broadcast finish", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)))

// 					i.logger.Debug("$$$$$$ AgreementComponent wait filler", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)))
// 					// wait for the value to be received
// 					i.WaitFillGapResponse(leader, priority, fillerContLen)
// 					i.logger.Debug("$$$$$$ AgreementComponent got filler", zap.Int64("time(micro)", makeTimestamp()), zap.Int("acround", int(i.State.ACState.ACRound)))

// 				}
// 			}

// 			// get decided value
// 			value, priority = queue.Peek()

// 			// remove the value from the queue and add it to S
// 			queue.Dequeue()

// 			i.State.Delivered.Enqueue(value, priority)

// 			// return the value to the client
// 			i.Deliver(value)
// 		}
// 		// increment the round number
// 		i.State.ACState.IncrementRound()
// 		i.logger.Debug("$$$$$$ AgreementComponent finish round", zap.Int64("time(micro)", makeTimestamp()), zap.Int("newround", int(i.State.ACState.ACRound)))

// 	}
// 	return nil
// }

// func (i *Instance) WaitFillGapResponse(leader types.OperatorID, priority specalea.Priority, fillerContLen int) {
// 	// gets the leader queue
// 	queue := i.State.VCBCState.Queues[leader]
// 	currentFillerNum := fillerContLen
// 	for {
// 		// if has the desired priority, returns
// 		_, localPriority := queue.Peek()
// 		if localPriority >= priority {
// 			return
// 		}

// 		// waits until a FILLER signal is received (actived on the uponFiller function)
// 		for {
// 			newLen := i.State.FillerContainer.Len(i.State.AleaDefaultRound)
// 			if newLen > currentFillerNum {
// 				currentFillerNum = newLen
// 				break
// 			}
// 		}
// 	}
// }

func (i *Instance) StartABA(vote byte, leader types.OperatorID) error {

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponStartABA "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("vote", int(vote)), zap.Int("leader", int(leader)))
	}
	

	log("start")

	priority := i.State.ACState.CurrentPriority(leader)
	log(fmt.Sprintf("priority: %v",int(priority)))


	initMsg, err := CreateABAInit(i.State, i.config, vote, specalea.FirstRound, leader, priority)
	if err != nil {
		return errors.Wrap(err, "UponStartABA: failed to create ABA Init message")
	}
	log("broadcast start abainit")
	i.Broadcast(initMsg)
	log("broadcast finish abainit")

	return nil
}
