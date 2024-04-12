package message

import (
	"strings"

	"golang.org/x/net/html"
)

type MessageElementText struct {
	*noAliasMessageElement
	Content string
}

func (e *MessageElementText) Tag() string {
	return "text"
}

func (e *MessageElementText) Stringify() string {
	return escape(e.Content, true)
}

func (e *MessageElementText) Parse(n *html.Node) (MessageElement, error) {
	if n.Type == html.TextNode {
		content := strings.TrimSpace(n.Data)
		if content != "" {
			return &MessageElementText{
				Content: content,
			}, nil
		}
	}
	return nil, nil
}

type MessageElementAt struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Id   string
	Name string //	收发	目标用户的名称
	Role string //	收发	目标角色
	Type string //	收发	特殊操作，例如 all 表示 @全体成员，here 表示 @在线成员
}

func (e *MessageElementAt) Tag() string {
	return "at"
}

func (e *MessageElementAt) Stringify() string {
	result := ""
	if e.Id != "" {
		result += ` id="` + escape(e.Id, true) + `"`
	}
	if e.Name != "" {
		result += ` name="` + escape(e.Name, true) + `"`
	}
	if e.Role != "" {
		result += ` role="` + escape(e.Role, true) + `"`
	}
	if e.Type != "" {
		result += ` type="` + escape(e.Type, true) + `"`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementAt) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementAt{
		Id:   attrMap["id"],
		Name: attrMap["name"],
		Role: attrMap["role"],
		Type: attrMap["type"],
	}
	for key, value := range attrMap {
		if key != "id" && key != "name" && key != "role" && key != "type" {
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

type MessageElementSharp struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Id   string //收发	目标频道的 ID
	Name string //收发	目标频道的名称
}

func (e *MessageElementSharp) Tag() string {
	return "sharp"
}

func (e *MessageElementSharp) Stringify() string {
	result := ""
	if e.Id != "" {
		result += ` id="` + escape(e.Id, true) + `"`
	}
	if e.Name != "" {
		result += ` name="` + escape(e.Name, true) + `"`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementSharp) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementSharp{
		Id:   attrMap["id"],
		Name: attrMap["name"],
	}
	for key, value := range attrMap {
		if key != "id" && key != "name" && key != "role" && key != "type" {
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

type MessageElementA struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Href string
}

func (e *MessageElementA) Tag() string {
	return "a"
}

func (e *MessageElementA) Stringify() string {
	result := ""
	if e.Href != "" {
		result += ` href="` + escape(e.Href, true) + `"`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}
func (e *MessageElementA) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementA{
		Href: attrMap["href"],
	}
	for key, value := range attrMap {
		if key != "href" {
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
	RegsiterParserElement(&MessageElementText{})
	RegsiterParserElement(&MessageElementAt{})
	RegsiterParserElement(&MessageElementSharp{})
	RegsiterParserElement(&MessageElementA{})
}
