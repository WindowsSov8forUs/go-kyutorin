package event

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tencent-connect/botgo/dto"
)

func TestRegisterHandlers(t *testing.T) {
	var guild GuildEventHandler = func(event *dto.Payload, data *dto.GuildData) error {
		return nil
	}
	var message MessageEventHandler = func(event *dto.Payload, data *dto.MessageData) error {
		return nil
	}
	var audio AudioEventHandler = func(event *dto.Payload, data *dto.AudioData) error {
		return nil
	}

	t.Run(
		"test intent", func(t *testing.T) {
			i := RegisterHandlers(guild, message, audio)
			log.Println(i)
			assert.Equal(t, dto.IntentGuildMessages, i&dto.IntentGuildMessages)
			assert.Equal(t, dto.IntentGuilds, i&dto.IntentGuilds)
			assert.Equal(t, dto.IntentAudioAction, i&dto.IntentAudioAction)
		},
	)
}
