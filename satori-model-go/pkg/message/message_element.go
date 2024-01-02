package message

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

type childrenMessageElement struct {
	Children []MessageElement
}

func (e *childrenMessageElement) stringifyChildren() string {
	var result string
	for _, e := range e.Children {
		result += e.Stringify()
	}
	return result
}

func (e *childrenMessageElement) stringifyByTag(tag string) string {
	if len(e.Children) == 0 {
		return "<" + tag + "/>"
	}
	return "<" + tag + ">" + e.stringifyChildren() + "</" + tag + ">"
}
