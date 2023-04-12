package instance

import (
	"testing"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/stretchr/testify/require"
)

func TestInstance_Marshaling(t *testing.T) {
	i := alea.TestingInstanceStruct

	byts, err := i.Encode()
	require.NoError(t, err)

	decoded := &Instance{}
	require.NoError(t, decoded.Decode(byts))

	bytsDecoded, err := decoded.Encode()
	require.NoError(t, err)
	require.EqualValues(t, byts, bytsDecoded)
}
