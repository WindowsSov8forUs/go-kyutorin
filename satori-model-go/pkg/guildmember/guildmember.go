package guildmember

import "github.com/satori-protocol-go/satori-model-go/pkg/user"

type GuildMember struct {
	User     *user.User `json:"user"`
	Nick     string     `json:"nick"`
	Avatar   string     `json:"avatar"`
	JoinedAt int64      `json:"joined_at"`
}

type GuildMemberList struct {
	Data []GuildMember `json:"data"`
	Next string        `json:"next,omitempty"`
}
