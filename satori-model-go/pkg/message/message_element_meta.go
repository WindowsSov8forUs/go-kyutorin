package message

import (
	"fmt"
	"strconv"
)

// 引用
type MessageElementQuote struct {
	*noAliasMessageElement
	*childrenMessageElement
}

func (e *MessageElementQuote) Tag() string {
	return "quote"
}

func (e *MessageElementQuote) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementQuote) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementQuote) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementQuote{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 作者
type MessageElementAuthor struct {
	*noAliasMessageElement
	Id     string // 用户 ID
	Name   string // 昵称
	Avatar string // 头像 URL
}

func (e *MessageElementAuthor) Tag() string {
	return "author"
}

func (e *MessageElementAuthor) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += ` id="` + Escape(e.Id) + `"`
	}
	if e.Name != "" {
		result += ` name="` + Escape(e.Name) + `"`
	}
	if e.Avatar != "" {
		result += ` avatar="` + Escape(e.Avatar) + `"`
	}
	return result + "/>"
}

func (e *MessageElementAuthor) parse(n *Node) (MessageElement, error) {
	result := &MessageElementAuthor{
		Id:     n.Attrs["id"],
		Name:   n.Attrs["name"],
		Avatar: n.Attrs["avatar"],
	}
	return result, nil
}

// 被动
type MessageElementPassive struct {
	*noAliasMessageElement
	Id  string // 被动消息 ID
	Seq int    // 被动消息序号
}

func (e *MessageElementPassive) Tag() string {
	return "passive"
}

func (e *MessageElementPassive) Stringify() string {
	result := "<" + e.Tag()
	if e.Id != "" {
		result += ` id="` + e.Id + `"`
	}
	if e.Seq != 0 {
		result += ` seq=` + fmt.Sprint(e.Seq)
	}
	return result + "/>"
}

func (e *MessageElementPassive) parse(n *Node) (MessageElement, error) {
	seq, _ := strconv.Atoi(n.Attrs["seq"])
	result := &MessageElementPassive{
		Id:  n.Attrs["id"],
		Seq: seq,
	}
	return result, nil
}

func init() {
	regsiterParserElement(&MessageElementQuote{})
	regsiterParserElement(&MessageElementAuthor{})
	regsiterParserElement(&MessageElementPassive{})
}
