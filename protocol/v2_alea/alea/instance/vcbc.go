package instance

import (
	"github.com/pkg/errors"

	// specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"go.uber.org/zap"
)

func (i *Instance) StartVCBC(data []byte) error {

	// logger
	log := func(str string) {

		if i.State.HideLogs || i.State.DecidedLogOnly {
			return
		}
		i.logger.Debug("$$$$$$ UponVCBCStart : "+str+"$$$$$$", zap.Int64("time(micro)", makeTimestamp()))
	}

	log("start")

	// create VCBCSend message and broadcasts
	msgToBroadcast, err := CreateVCBCSend(i.State, i.config, data)
	if err != nil {
		return errors.Wrap(err, "StartVCBC: failed to create VCBCSend message")
	}
	log("created vcbc send")

	i.Broadcast(msgToBroadcast)
	log("broadcasted")

	return nil
}
