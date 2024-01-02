package message

import (
	"fmt"
	"strconv"
)

type resourceRootMessageElement struct {
	Src     string // 资源的 URL
	Cache   bool   // 是否使用已缓存的文件
	Timeout string // 下载文件的最长时间 (毫秒)
}

func parseResourceRootMessageElement(attrMap map[string]string) *resourceRootMessageElement {
	result := &resourceRootMessageElement{
		Src:     attrMap["src"],
		Cache:   false,
		Timeout: attrMap["timeout"],
	}
	cacheAttr, ok := attrMap["cache"]
	if ok {
		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
	}
	return result
}

func (e *resourceRootMessageElement) attrString() string {
	result := ""
	if e.Src != "" {
		result += ` src="` + Escape(e.Src) + `"`
	}
	if e.Cache {
		result += ` cache`
	}
	if e.Timeout != "" {
		result += ` timeout="` + e.Timeout + `"`
	}
	return result
}

func (e *resourceRootMessageElement) stringifyByTag(tag string) string {
	return "<" + tag + e.attrString() + "/>"
}

type MessageElementImg struct {
	*resourceRootMessageElement
	Width  uint32 // 图片的宽度
	Height uint32 // 图片的高度
}

func (e *MessageElementImg) Tag() string {
	return "img"
}

func (e *MessageElementImg) Alias() []string {
	return []string{"image"}
}

func (e *MessageElementImg) Stringify() string {
	result := "<" + e.Tag()
	attrStr := e.attrString()
	if attrStr != "" {
		result += attrStr
	}
	if e.Width > 0 {
		attrStr += fmt.Sprintf(" width=%d", e.Width)
	}
	if e.Height > 0 {
		attrStr += fmt.Sprintf(" height=%d", e.Height)
	}
	return result + "/>"
}

func (e *MessageElementImg) SetSrc(src string) {
	// 首先判断是否已被初始化
	if e.resourceRootMessageElement == nil {
		e.resourceRootMessageElement = &resourceRootMessageElement{}
	}
	e.resourceRootMessageElement.Src = src
}

func (e *MessageElementImg) parse(n *Node) (MessageElement, error) {
	root := parseResourceRootMessageElement(n.Attrs)
	result := &MessageElementImg{
		resourceRootMessageElement: root,
	}
	if w, ok := n.Attrs["width"]; ok {
		width, e := strconv.Atoi(w)
		if e != nil {
			return nil, fmt.Errorf("width[%s] is illegal:%v", w, e)
		}
		result.Width = uint32(width)
	}
	if h, ok := n.Attrs["height"]; ok {
		height, e := strconv.Atoi(h)
		if e != nil {
			return nil, fmt.Errorf("height[%s] is illegal:%v", h, e)
		}
		result.Height = uint32(height)
	}
	return result, nil
}

type MessageElementAudio struct {
	*noAliasMessageElement
	*resourceRootMessageElement
}

func (e *MessageElementAudio) Tag() string {
	return "audio"
}

func (e *MessageElementAudio) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementAudio) SetSrc(src string) {
	// 首先判断是否已被初始化
	if e.resourceRootMessageElement == nil {
		e.resourceRootMessageElement = &resourceRootMessageElement{}
	}
	e.resourceRootMessageElement.Src = src
}

func (e *MessageElementAudio) parse(n *Node) (MessageElement, error) {
	return &MessageElementAudio{
		resourceRootMessageElement: parseResourceRootMessageElement(n.Attrs),
	}, nil
}

type MessageElementVideo struct {
	*noAliasMessageElement
	*resourceRootMessageElement
}

func (e *MessageElementVideo) Tag() string {
	return "video"
}

func (e *MessageElementVideo) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementVideo) SetSrc(src string) {
	// 首先判断是否已被初始化
	if e.resourceRootMessageElement == nil {
		e.resourceRootMessageElement = &resourceRootMessageElement{}
	}
	e.resourceRootMessageElement.Src = src
}

func (e *MessageElementVideo) parse(n *Node) (MessageElement, error) {
	return &MessageElementVideo{
		resourceRootMessageElement: parseResourceRootMessageElement(n.Attrs),
	}, nil
}

type MessageElementFile struct {
	*noAliasMessageElement
	*resourceRootMessageElement
}

func (e *MessageElementFile) Tag() string {
	return "file"
}

func (e *MessageElementFile) Stringify() string {
	return e.stringifyByTag(e.Tag())
}

func (e *MessageElementFile) SetSrc(src string) {
	// 首先判断是否已被初始化
	if e.resourceRootMessageElement == nil {
		e.resourceRootMessageElement = &resourceRootMessageElement{}
	}
	e.resourceRootMessageElement.Src = src
}

func (e *MessageElementFile) parse(n *Node) (MessageElement, error) {
	return &MessageElementFile{
		resourceRootMessageElement: parseResourceRootMessageElement(n.Attrs),
	}, nil
}

func init() {
	regsiterParserElement(&MessageElementImg{})
	regsiterParserElement(&MessageElementAudio{})
	regsiterParserElement(&MessageElementVideo{})
	regsiterParserElement(&MessageElementFile{})
}
