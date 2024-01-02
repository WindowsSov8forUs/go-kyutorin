package guild

type Guild struct {
	Id     string `json:"id"`               // 私聊频道
	Name   string `json:"name,omitempty"`   // 私聊频道
	Avatar string `json:"avatar,omitempty"` // 群组头像
}

type GuildList struct {
	Data []Guild `json:"data"`
	Next string  `json:"next,omitempty"`
}
