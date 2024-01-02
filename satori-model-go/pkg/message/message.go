package message

import (
	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/user"
)

type Message struct {
	Id       string                   `json:"id"`                  // 消息 ID
	Content  string                   `json:"content"`             // 消息内容
	Channel  *channel.Channel         `json:"channel,omitempty"`   // 频道对象
	Guild    *guild.Guild             `json:"guild,omitempty"`     // 群组对象
	Member   *guildmember.GuildMember `json:"member,omitempty"`    // 成员对象
	User     *user.User               `json:"user,omitempty"`      // 用户对象
	CreateAt int64                    `json:"create_at,omitempty"` // 消息发送的时间戳
	UpdateAt int64                    `json:"update_at,omitempty"` // 消息修改的时间戳
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
