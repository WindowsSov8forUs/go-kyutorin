package message

// 粗体
type MessageElementStrong struct {
	*childrenMessageElement
}

func (e *MessageElementStrong) Tag() string {
	return "b"
}

func (e *MessageElementStrong) Alias() []string {
	return []string{"strong"}
}

func (e *MessageElementStrong) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementStrong) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementStrong) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementStrong{
		&childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 斜体
type MessageElementEm struct {
	*childrenMessageElement
}

func (e *MessageElementEm) Tag() string {
	return "i"
}

func (e *MessageElementEm) Alias() []string {
	return []string{"em"}
}

func (e *MessageElementEm) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementEm) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementEm) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementEm{
		&childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 下划线
type MessageElementIns struct {
	*childrenMessageElement
}

func (e *MessageElementIns) Tag() string {
	return "s"
}

func (e *MessageElementIns) Alias() []string {
	return []string{"ins"}
}

func (e *MessageElementIns) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementIns) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementIns) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementIns{
		&childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 删除线
type MessageElementDel struct {
	*childrenMessageElement
}

func (e *MessageElementDel) Tag() string {
	return "s"
}

func (e *MessageElementDel) Alias() []string {
	return []string{"del"}
}

func (e *MessageElementDel) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementDel) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementDel) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementDel{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 剧透
type MessageElementSpl struct {
	*noAliasMessageElement
	*childrenMessageElement
}

func (e *MessageElementSpl) Tag() string {
	return "spl"
}

func (e *MessageElementSpl) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementSpl) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementSpl) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementSpl{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 代码
type MessageElementCode struct {
	*childrenMessageElement
	*noAliasMessageElement
}

func (e *MessageElementCode) Tag() string {
	return "code"
}

func (e *MessageElementCode) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementCode) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementCode) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementCode{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 上标
type MessageElementSup struct {
	*childrenMessageElement
	*noAliasMessageElement
}

func (e *MessageElementSup) Tag() string {
	return "sup"
}

func (e *MessageElementSup) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementSup) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementSup) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementSup{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

// 下标
type MessageElementSub struct {
	*childrenMessageElement
	*noAliasMessageElement
}

func (e *MessageElementSub) Tag() string {
	return "sub"
}

func (e *MessageElementSub) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementSub) SetChildren(children []MessageElement) {
	// 首先判断是否已被初始化
	if e.childrenMessageElement == nil {
		e.childrenMessageElement = &childrenMessageElement{}
	}
	e.childrenMessageElement.Children = children
}

func (e *MessageElementSub) parse(n *Node) (MessageElement, error) {
	var children []MessageElement
	err := parseChildrenNode(n, func(e MessageElement) {
		children = append(children, e)
	})
	if err != nil {
		return nil, err
	}
	return &MessageElementSub{
		childrenMessageElement: &childrenMessageElement{
			Children: children,
		},
	}, nil
}

func init() {
	regsiterParserElement(&MessageElementStrong{})
	regsiterParserElement(&MessageElementEm{})
	regsiterParserElement(&MessageElementIns{})
	regsiterParserElement(&MessageElementDel{})
	regsiterParserElement(&MessageElementSpl{})
	regsiterParserElement(&MessageElementCode{})
	regsiterParserElement(&MessageElementSup{})
	regsiterParserElement(&MessageElementSub{})
}
