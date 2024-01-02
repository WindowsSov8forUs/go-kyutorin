package interaction

type Argv struct {
	Name      string        `json:"name"`      // 指令名称
	Arguments []interface{} `json:"arguments"` // 参数
	Options   []interface{} `json:"options"`   // 选项
}

type Button struct {
	Id string `json:"id"` // 按钮 ID
}
