package instance

import (
	"github.com/pkg/errors"

	"go.uber.org/zap"
)

func (i *Instance) StartVCBC(data []byte) error {

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$"+cBlue+" UponVCBCStart "+reset+"1: "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()), zap.Int("own operator id", int(i.State.Share.OperatorID)))
	}

	log("start")

	// Create VCBCSend message and broadcasts
	msgToBroadcast, err := i.CreateVCBCSend(data)
	if err != nil {
		return errors.Wrap(err, "StartVCBC: failed to create VCBCSend message")
	}

	i.Broadcast(msgToBroadcast)
	log("broadcasted")

	return nil
}
