package message

import (
	"golang.org/x/net/html"
)

type MessageElmentBr struct {
	*noAliasMessageElement
}

func (e *MessageElmentBr) Tag() string {
	return "br"
}

func (e *MessageElmentBr) Stringify() string {
	return "<br/>"
}

func (e *MessageElmentBr) Parse(n *html.Node) (MessageElement, error) {
	return &MessageElmentBr{}, nil
}

type MessageElmentP struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElmentP) Tag() string {
	return "p"
}

func (e *MessageElmentP) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return "<" + e.Tag() + result + "/>"
	}
	return "<" + e.Tag() + result + ">" + childrenStr + "</" + e.Tag() + ">"
}

func (e *MessageElmentP) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElmentP{}
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

type MessageElementMessage struct {
	Id      string
	Forward bool
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementMessage) Tag() string {
	return "message"
}

func (e *MessageElementMessage) Stringify() string {
	result := ""
	if e.Id != "" {
		result += ` id="` + escape(e.Id, true) + `"`
	}
	if e.Forward {
		result += ` forward`
	}
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return "<" + e.Tag() + result + "/>"
	}
	return "<" + e.Tag() + result + ">" + childrenStr + "</" + e.Tag() + ">"
}

func (e *MessageElementMessage) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementMessage{
		Forward: false,
	}
	if id, ok := attrMap["id"]; ok {
		result.Id = id
	}
	if forwardAttr, ok := attrMap["forward"]; ok {
		result.Forward = forwardAttr == "" || forwardAttr == "true" || forwardAttr == "1"
	}
	for key, value := range attrMap {
		if key != "id" && key != "forward" {
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
	RegsiterParserElement(&MessageElmentBr{})
	RegsiterParserElement(&MessageElmentP{})
	RegsiterParserElement(&MessageElementMessage{})
}
