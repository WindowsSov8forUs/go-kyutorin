package message

import (
	"golang.org/x/net/html"
)

type MessageElement interface {
	Tag() string
	Stringify() string
	Alias() []string
}

type noAliasMessageElement struct {
}

func (e *noAliasMessageElement) Alias() []string {
	return nil
}

type ChildrenMessageElement struct {
	Children []MessageElement
}

func (e *ChildrenMessageElement) stringifyChildren() string {
	if e == nil {
		return ""
	}
	if len(e.Children) == 0 {
		return ""
	}
	var result string
	for _, e := range e.Children {
		result += e.Stringify()
	}
	return result
}

func (e *ChildrenMessageElement) parseChildren(n *html.Node) (*ChildrenMessageElement, error) {
	var children []MessageElement
	err := parseHtmlChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	result := &ChildrenMessageElement{
		Children: children,
	}
	return result, nil
}

type ExtendAttributes struct {
	Attributes map[string]string
}

func (e *ExtendAttributes) AddAttribute(key, value string) *ExtendAttributes {
	result := e
	if result == nil {
		result = &ExtendAttributes{
			Attributes: make(map[string]string),
		}
	}
	result.Attributes[key] = value
	return result
}

func (e *ExtendAttributes) Get(key string) (string, bool) {
	if e == nil {
		return "", false
	}
	v, ok := e.Attributes[key]
	return v, ok
}

func (e *ExtendAttributes) stringifyAttributes() string {
	if e == nil || len(e.Attributes) == 0 {
		return ""
	}
	var result string
	for k, v := range e.Attributes {
		if v == "" {
			result += " " + k
		} else {
			result += " " + k + `="` + escape(v, true) + `"`
		}
	}
	return result
}
