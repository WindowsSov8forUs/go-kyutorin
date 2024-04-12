package message

import (
	"golang.org/x/net/html"
)

type MessageElementStrong struct {
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementStrong) Tag() string {
	return "b"
}
func (e *MessageElementStrong) Alias() []string {
	return []string{"strong"}
}

func (e *MessageElementStrong) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementStrong) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementStrong{}
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

type MessageElementEm struct {
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementEm) Tag() string {
	return "i"
}
func (e *MessageElementEm) Alias() []string {
	return []string{"em"}
}

func (e *MessageElementEm) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementEm) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementEm{}
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

type MessageElementIns struct {
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementIns) Tag() string {
	return "u"
}

func (e *MessageElementIns) Alias() []string {
	return []string{"ins"}
}

func (e *MessageElementIns) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementIns) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementIns{}
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

type MessageElementDel struct {
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementDel) Tag() string {
	return "s"
}

func (e *MessageElementDel) Alias() []string {
	return []string{"del"}
}

func (e *MessageElementDel) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementDel) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementDel{}
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

type MessageElementSpl struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
}

func (e *MessageElementSpl) Tag() string {
	return "spl"
}

func (e *MessageElementSpl) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementSpl) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementSpl{}
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

type MessageElementCode struct {
	*ChildrenMessageElement
	*noAliasMessageElement
	*ExtendAttributes
}

func (e *MessageElementCode) Tag() string {
	return "code"
}

func (e *MessageElementCode) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementCode) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementCode{}
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

type MessageElementSup struct {
	*ChildrenMessageElement
	*noAliasMessageElement
	*ExtendAttributes
}

func (e *MessageElementSup) Tag() string {
	return "sup"
}

func (e *MessageElementSup) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementSup) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementSup{}
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

type MessageElementSub struct {
	*ChildrenMessageElement
	*noAliasMessageElement
	*ExtendAttributes
}

func (e *MessageElementSub) Tag() string {
	return "sub"
}

func (e *MessageElementSub) Stringify() string {
	result := e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementSub) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementSub{}
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

func init() {
	RegsiterParserElement(&MessageElementStrong{})
	RegsiterParserElement(&MessageElementEm{})
	RegsiterParserElement(&MessageElementIns{})
	RegsiterParserElement(&MessageElementDel{})
	RegsiterParserElement(&MessageElementSpl{})
	RegsiterParserElement(&MessageElementCode{})
	RegsiterParserElement(&MessageElementSup{})
	RegsiterParserElement(&MessageElementSub{})
}
