package instance

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"

	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"go.uber.org/zap"
)

func (i *Instance) StartCV() error {

	//funciton identifier
	functionID := uuid.New().String()

	// logger
	log := func(str string) {
		i.logger.Debug("$$$$$$ UponStartCV "+functionID+": "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	data, _ := i.State.VCBCState.GetMaxValueOccurences()
	vcbcData := i.State.VCBCState.GetVCBCDataByData(data)

	dataAuthor := vcbcData.Author
	aggregatedSignature := vcbcData.AggregatedSignature
	nodeIDs := vcbcData.NodeIDs


	// author := i.State.Share.OperatorID

	log("create cv vote for round 0")
	// create VCBCSend message and broadcasts
	msgToBroadcast, err := CreateCVVote(i.State, i.config, data, dataAuthor,aggregatedSignature,nodeIDs, 0)
	if err != nil {
		return errors.Wrap(err, "StartVCBC: failed to create CVVote message")
	}


	log("broadcast start")
	i.Broadcast(msgToBroadcast)
	log("broadcast finish")

	
	return nil
}
