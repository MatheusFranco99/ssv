package runner

import (
	specqbft "github.com/MatheusFranco99/ssv-spec-AleaBFT/qbft"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2/qbft/instance"
	"github.com/MatheusFranco99/ssv/protocol/v2/qbft/roundtimer"
)

type TimeoutF func(identifier spectypes.MessageID, height specqbft.Height) func()

func (b *BaseRunner) registerTimeoutHandler(instance *instance.Instance, height specqbft.Height) {
	identifier := spectypes.MessageIDFromBytes(instance.State.ID)
	timer, ok := instance.GetConfig().GetTimer().(*roundtimer.RoundTimer)
	if ok {
		timer.OnTimeout(b.TimeoutF(identifier, height))
	}
}
