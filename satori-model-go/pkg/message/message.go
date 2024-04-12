package message

import (
	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
)

type Message struct {
	Id       string                   `json:"id"`
	Content  string                   `json:"content"`
	Channel  *channel.Channel         `json:"channel,omitempty"`
	Guild    *guild.Guild             `json:"guild,omitempty"`
	Member   *guildmember.GuildMember `json:"member,omitempty"`
	User     *user.User               `json:"user,omitempty"`
	CreateAt int64                    `json:"create_at,omitempty"`
	UpdateAt int64                    `json:"update_at,omitempty"`
}

type MessageList struct {
	Data []Message `json:"data"`
	Next string    `json:"next,omitempty"`
}

func (m *Message) Decode(elements []MessageElement) error {
	raw := ""
	for _, e := range elements {
		raw += e.Stringify()
	}
	m.Content = raw
	return nil
}

func (m *Message) Encode() ([]MessageElement, error) {
	if m.Content == "" {
		return nil, nil
	}
	return Parse(m.Content)
}
