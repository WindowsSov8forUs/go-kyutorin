package message

import (
	"fmt"
)

type MessageElmentBr struct {
	*noAliasMessageElement
}

func (e *MessageElmentBr) Tag() string {
	return "br"
}

func (e *MessageElmentBr) Stringify() string {
	return fmt.Sprintln()
}

func (e *MessageElmentBr) parse(n *Node) (MessageElement, error) {
	return &MessageElmentBr{}, nil
}

type MessageElmentP struct {
	*noAliasMessageElement
	*childrenMessageElement
}

func (e *MessageElmentP) Tag() string {
	return "p"
}

func (e *MessageElmentP) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElmentP) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElmentP) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElmentP{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

type MessageElementMessage struct {
	*noAliasMessageElement
	*childrenMessageElement
	Id      string // 消息的 ID
	Forward bool   // 是否为转发消息
}

func (e *MessageElementMessage) Tag() string {
	return "message"
}

func (e *MessageElementMessage) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += " id=\"" + Escape(e.Id) + "\""
	}
	if e.Forward {
		result += " forward"
	}
	if len(e.Children) == 0 {
		return result + "/>"
	} else {
		return result + ">" + e.stringifyChildren() + "</" + e.Tag() + ">"
	}
}

func (e *MessageElementMessage) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementMessage) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementMessage{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
		Id:      n.Attrs["id"],
		Forward: n.Attrs["forward"] == "" || n.Attrs["forward"] == "true" || n.Attrs["forward"] == "1",
	}, nil
}

func init() {
	regsiterParserElement(&MessageElmentBr{})
	regsiterParserElement(&MessageElmentP{})
	regsiterParserElement(&MessageElementMessage{})
}
