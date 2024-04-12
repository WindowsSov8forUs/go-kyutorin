package interaction

type Argv struct {
	Name      string                 `json:"name"`
	Arguments []interface{}          `json:"arguments"`
	Options   map[string]interface{} `json:"options"`
}

type Button struct {
	Id string `json:"id"`
}
