package guild

type Guild struct {
	Id     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

type GuildList struct {
	Data []Guild `json:"data"`
	Next string  `json:"next,omitempty"`
}
