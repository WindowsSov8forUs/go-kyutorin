package message

import "golang.org/x/net/html"

type MessageElementQuote struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementQuote) Tag() string {
	return "quote"
}

func (e *MessageElementQuote) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementQuote) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementQuote{}
	for key, value := range attrMap {
		result.ExtendAttributes = result.AddAttribute(key, value)
	}
	children, err := result.parseChildren(n)
	if err != nil {
		return nil, err
	}
	result.ChildrenMessageElement = children
	return result, nil
}

type MessageElementAuthor struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Id     string
	Name   string
	Avatar string
}

func (e *MessageElementAuthor) Tag() string {
	return "author"
}

func (e *MessageElementAuthor) Stringify() string {
	result := ""
	if e.Id != "" {
		result += ` id="` + escape(e.Id, true) + `"`
	}
	if e.Name != "" {
		result += ` name="` + escape(e.Name, true) + `"`
	}
	if e.Avatar != "" {
		result += ` avatar="` + escape(e.Avatar, true) + `"`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementAuthor) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementAuthor{
		Id:     attrMap["id"],
		Name:   attrMap["name"],
		Avatar: attrMap["avatar"],
	}
	for key, value := range attrMap {
		if key != "id" && key != "name" && key != "avatar" {
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
	RegsiterParserElement(&MessageElementQuote{})
	RegsiterParserElement(&MessageElementAuthor{})
}
