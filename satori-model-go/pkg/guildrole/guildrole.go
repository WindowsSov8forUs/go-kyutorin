package guildrole

type GuildRole struct {
	Id   string `json:"id"`             // 角色 ID
	Name string `json:"name,omitempty"` // 角色名称
}
type GuildRoleList struct {
	Data []GuildRole `json:"data"`
	Next string      `json:"next,omitempty"`
}
