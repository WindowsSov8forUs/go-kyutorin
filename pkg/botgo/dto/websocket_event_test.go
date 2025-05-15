package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_transposeIntentEventMap(t *testing.T) {
	t.Run("transpose", func(t *testing.T) {
		re := transposeIntentEventMap(intentEventMap)
		assert.Equal(t, re[EventAudioFinish], IntentAudioAction)
		assert.Equal(t, re[EventAudioOffMic], IntentAudioAction)
		assert.Equal(t, re[EventChannelCreate], IntentGuilds)
	})
}
