package message

// 纯文本
type MessageElementText struct {
	*noAliasMessageElement
	Content string // 一段纯文本
}

func (e *MessageElementText) Tag() string {
	return "text"
}

func (e *MessageElementText) Stringify() string {
	return Escape(e.Content)
}

func (e *MessageElementText) parse(n *Node) (MessageElement, error) {
	return &MessageElementText{
		Content: n.Attrs["content"],
	}, nil
}

// 提及用户
type MessageElementAt struct {
	*noAliasMessageElement
	Id   string // 收发	目标用户的 ID
	Name string // 收发	目标用户的名称
	Role string // 收发	目标角色
	Type string // 收发	特殊操作，例如 all 表示 @全体成员，here 表示 @在线成员
}

func (e *MessageElementAt) Tag() string {
	return "at"
}

func (e *MessageElementAt) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += ` id="` + Escape(e.Id) + `"`
	}
	if e.Name != "" {
		result += ` name="` + Escape(e.Name) + `"`
	}
	if e.Role != "" {
		result += ` role="` + Escape(e.Role) + `"`
	}
	if e.Type != "" {
		result += ` type="` + e.Type + `"`
	}
	return result + "/>"
}

func (e *MessageElementAt) parse(n *Node) (MessageElement, error) {
	return &MessageElementAt{
		Id:   n.Attrs["id"],
		Name: n.Attrs["name"],
		Role: n.Attrs["role"],
		Type: n.Attrs["type"],
	}, nil
}

// 提及频道
type MessageElementSharp struct {
	*noAliasMessageElement
	Id   string // 收发 目标频道的 ID
	Name string // 收发 目标频道的名称
}

func (e *MessageElementSharp) Tag() string {
	return "sharp"
}

func (e *MessageElementSharp) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += ` id="` + Escape(e.Id) + `"`
	}
	if e.Name != "" {
		result += ` name="` + Escape(e.Name) + `"`
	}
	return result + "/>"
}

func (e *MessageElementSharp) parse(n *Node) (MessageElement, error) {
	return &MessageElementSharp{
		Id:   n.Attrs["id"],
		Name: n.Attrs["name"],
	}, nil
}

// 链接
type MessageElementA struct {
	*noAliasMessageElement
	Href string // 链接的 URL
}

func (e *MessageElementA) Tag() string {
	return "a"
}

func (e *MessageElementA) Stringify() string {
	result := "<" + e.Tag()
	if e.Href != "" {
		result += ` href="` + Escape(e.Href) + `"`
	}
	return result + "/>"
}

func (e *MessageElementA) parse(n *Node) (MessageElement, error) {
	return &MessageElementA{
		Href: n.Attrs["href"],
	}, nil
}

func init() {
	regsiterParserElement(&MessageElementText{})
	regsiterParserElement(&MessageElementAt{})
	regsiterParserElement(&MessageElementSharp{})
	regsiterParserElement(&MessageElementA{})
}
