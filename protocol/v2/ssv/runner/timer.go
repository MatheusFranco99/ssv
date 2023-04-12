package runner

import (
	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/instance"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/roundtimer"
)

type TimeoutF func(identifier spectypes.MessageID, height specalea.Height) func()

func (b *BaseRunner) registerTimeoutHandler(instance *instance.Instance, height specalea.Height) {
	identifier := spectypes.MessageIDFromBytes(instance.State.ID)
	timer, ok := instance.GetConfig().GetTimer().(*roundtimer.RoundTimer)
	if ok {
		timer.OnTimeout(b.TimeoutF(identifier, height))
	}
}
