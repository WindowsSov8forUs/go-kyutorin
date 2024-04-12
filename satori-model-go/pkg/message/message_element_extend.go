package message

import "golang.org/x/net/html"

type MessageElementExtend struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Type string
}

func (e *MessageElementExtend) Tag() string {
	return e.Type
}

func (e *MessageElementExtend) Stringify() string {
	result := ""
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementExtend) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementExtend{
		Type: n.Data,
	}
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

var ExtendParser = &MessageElementExtend{}
