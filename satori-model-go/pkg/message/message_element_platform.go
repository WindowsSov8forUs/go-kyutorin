package message

import (
	"fmt"
	"regexp"
)

var customTagPat = regexp.MustCompile(`(.+):(.+)$`)

type MessageElementCustom struct {
	*noAliasMessageElement
	*childrenMessageElement
	Platform  string                 // 平台名称
	CustomTag string                 // 标签名称
	Attrs     map[string]interface{} // 属性
}

func (e *MessageElementCustom) Tag() string {
	return fmt.Sprintf("%s:%s", e.Platform, e.CustomTag)
}

func (e *MessageElementCustom) Stringify() string {
	result := "<" + e.Tag()
	for k, v := range e.Attrs {
		switch _v := v.(type) {
		case string:
			result += fmt.Sprintf(" %s=\"%s\"", k, Escape(_v))
		case int:
			result += fmt.Sprintf(" %s=\"%d\"", k, _v)
		case bool:
			if _v {
				result += fmt.Sprintf(" %s", k)
			}
		default:
			result += fmt.Sprintf(" %s=%v", k, Escape(fmt.Sprint(v)))
		}
	}
	if len(e.Children) == 0 {
		return result + "/>"
	} else {
		return result + ">" + e.stringifyChildren() + "</" + e.Tag() + ">"
	}
}

func (e *MessageElementCustom) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementCustom) parse(n *Node) (MessageElement, error) {
	// 判断是否为自定义标签，自定义标签格式：platform:tag
	customTagMatch := customTagPat.FindStringSubmatch(n.Type)
	if customTagMatch != nil {
		e.Platform = customTagMatch[1]
		e.CustomTag = customTagMatch[2]
		for k, v := range n.Attrs {
			e.Attrs[k] = v
		}
		if len(n.Children) > 0 {
			var children []MessageElement
			err := parseChildrenNode(n, func(e MessageElement) {
				children = append(children, e)
			})
			if err != nil {
				return nil, err
			}
			e.SetChildren(children)
		}
		return e, nil
	} else {
		return nil, nil
	}
}

func parseCustomNode(n *Node) (MessageElement, error) {
	element := &MessageElementCustom{}
	return element.parse(n)
}
