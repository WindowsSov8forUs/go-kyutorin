package message

import "golang.org/x/net/html"

type MessageElementButton struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	//	id	string?	发	按钮的 ID
	//
	// type	string?	发	按钮的类型
	// href	string?	发	按钮的链接
	// text	string?	发	待输入文本
	// theme	string?	发	按钮的样式
	Id    string
	Type  string
	Href  string
	Text  string
	Theme string
}

func (e *MessageElementButton) Tag() string {
	return "button"
}

func (e *MessageElementButton) Stringify() string {
	result := ""
	if e.Id != "" {
		result += ` id="` + escape(e.Id, true) + `"`
	}
	if e.Type != "" {
		result += ` type="` + escape(e.Type, true) + `"`
	}
	if e.Href != "" {
		result += ` href="` + escape(e.Href, true) + `"`
	}
	if e.Text != "" {
		result += ` text="` + escape(e.Text, true) + `"`
	}
	if e.Theme != "" {
		result += ` theme="` + escape(e.Theme, true) + `"`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementButton) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementButton{
		Id:    attrMap["id"],
		Type:  attrMap["type"],
		Href:  attrMap["href"],
		Text:  attrMap["text"],
		Theme: attrMap["theme"],
	}
	for key, value := range attrMap {
		if key != "id" && key != "type" && key != "href" && key != "text" && key != "theme" {
			result.ExtendAttributes = result.AddAttribute(key, value)
		}
	}
	children, err := result.parseChildren(n)
	if err != nil {
		return nil, err
	}
	result.ChildrenMessageElement = children
	return result, nil
}

func init() {
	RegsiterParserElement(&MessageElementButton{})
}
