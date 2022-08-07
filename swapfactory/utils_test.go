package swapfactory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStage_StageToString(t *testing.T) {
	expectedValues := []string{
		"Invalid",
		"Pending",
		"Ready",
		"Completed",
		"UnknownStageValue(4)",
	}
	for s := byte(0); s < byte(5); s++ {
		require.Equal(t, expectedValues[s], StageToString(s))
	}
}
