package message

type MessageElementButton struct {
	*noAliasMessageElement
	Id    string // 按钮的 ID
	Type  string // 按钮的类型
	Href  string // 按钮的链接
	Text  string // 待输入文本
	Theme string // 按钮的样式
}

func (e *MessageElementButton) Tag() string {
	return "button"
}

func (e *MessageElementButton) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += ` id="` + Escape(e.Id) + `"`
	}
	if e.Type != "" {
		result += ` type="` + Escape(e.Type) + `"`
	}
	if e.Href != "" {
		result += ` href="` + Escape(e.Href) + `"`
	}
	if e.Text != "" {
		result += ` text="` + Escape(e.Text) + `"`
	}
	if e.Theme != "" {
		result += ` theme="` + Escape(e.Theme) + `"`
	}
	return result + " />"
}

func (e *MessageElementButton) parse(n *Node) (MessageElement, error) {
	return &MessageElementButton{
		Id:    n.Attrs["id"],
		Type:  n.Attrs["type"],
		Href:  n.Attrs["href"],
		Text:  n.Attrs["text"],
		Theme: n.Attrs["theme"],
	}, nil
}

func init() {
	regsiterParserElement(&MessageElementButton{})
}
