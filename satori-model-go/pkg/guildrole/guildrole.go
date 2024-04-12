package guildrole

type GuildRole struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type GuildRoleList struct {
	Data []GuildRole `json:"data"`
	Next string      `json:"next,omitempty"`
}
