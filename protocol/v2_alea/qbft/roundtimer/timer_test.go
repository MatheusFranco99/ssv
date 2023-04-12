package roundtimer

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRoundTimer_TimeoutForRound(t *testing.T) {
	t.Run("TimeoutForRound", func(t *testing.T) {
		count := int32(0)
		onTimeout := func() {
			atomic.AddInt32(&count, 1)
		}
		timer := New(context.Background(), zap.L(), onTimeout)
		timer.roundTimeout = func(round specalea.Round) time.Duration {
			return 1100 * time.Millisecond
		}
		timer.TimeoutForRound(specalea.Round(1))
		require.Equal(t, int32(0), atomic.LoadInt32(&count))
		<-time.After(timer.roundTimeout(specalea.Round(1)) + time.Millisecond*10)
		require.Equal(t, int32(1), atomic.LoadInt32(&count))
	})

	t.Run("timeout round before elapsed", func(t *testing.T) {
		count := int32(0)
		onTimeout := func() {
			atomic.AddInt32(&count, 1)
		}
		timer := New(context.Background(), zap.L(), onTimeout)
		timer.roundTimeout = func(round specalea.Round) time.Duration {
			return 1100 * time.Millisecond
		}

		timer.TimeoutForRound(specalea.Round(1))
		<-time.After(timer.roundTimeout(specalea.Round(1)) / 2)
		timer.TimeoutForRound(specalea.Round(2)) // reset before elapsed
		require.Equal(t, int32(0), atomic.LoadInt32(&count))
		<-time.After(timer.roundTimeout(specalea.Round(2)) + time.Millisecond*10)
		require.Equal(t, int32(1), atomic.LoadInt32(&count))
	})
}
