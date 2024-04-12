package message

import (
	"fmt"
	"strconv"

	"golang.org/x/net/html"
)

// type ResourceRootMessageElement struct {
// 	Src     string
// 	Cache   bool
// 	Timeout string //ms
// }

// func parseResourceRootMessageElement(attrMap map[string]string) *ResourceRootMessageElement {
// 	result := &ResourceRootMessageElement{
// 		Src:     attrMap["src"],
// 		Cache:   false,
// 		Timeout: attrMap["timeout"],
// 	}
// 	cacheAttr, ok := attrMap["cache"]
// 	if ok {
// 		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
// 	}
// 	return result
// }

// func (e *ResourceRootMessageElement) attrString() string {
// 	if e == nil {
// 		return ""
// 	}
// 	result := ""
// 	if e.Src != "" {
// 		result += ` src="` + e.Src + `"`
// 	}
// 	if e.Cache {
// 		result += ` cache`
// 	}
// 	if e.Timeout != "" {
// 		result += ` timeout="` + e.Timeout + `"`
// 	}
// 	return result
// }

// func (e *ResourceRootMessageElement) stringifyByTag(tag string) string {
// 	return "<" + tag + e.attrString() + "/>"
// }

type MessageElementImg struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Src     string
	Title   string
	Cache   bool
	Timeout string //ms
	Width   uint32
	Height  uint32
}

func (e *MessageElementImg) Tag() string {
	return "img"
}

func (e *MessageElementImg) attrString() string {
	if e == nil {
		return ""
	}
	result := ""
	if e.Src != "" {
		result += ` src="` + escape(e.Src, true) + `"`
	}
	if e.Title != "" {
		result += ` title="` + escape(e.Title, true) + `"`
	}
	if e.Cache {
		result += ` cache`
	}
	if e.Timeout != "" {
		result += ` timeout="` + e.Timeout + `"`
	}
	return result
}

func (e *MessageElementImg) Stringify() string {
	result := ""
	attrStr := e.attrString()
	if e.Width > 0 {
		attrStr += fmt.Sprintf(` width="%d"`, e.Width)
	}
	if e.Height > 0 {
		attrStr += fmt.Sprintf(` height="%d"`, e.Height)
	}
	result += attrStr
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementImg) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementImg{
		Src:     attrMap["src"],
		Title:   attrMap["title"],
		Cache:   false,
		Timeout: attrMap["timeout"],
	}
	cacheAttr, ok := attrMap["cache"]
	if ok {
		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
	}
	if w, ok := attrMap["width"]; ok {
		width, e := strconv.Atoi(w)
		if e != nil {
			return nil, fmt.Errorf("width[%s] is illegal:%v", w, e)
		}
		result.Width = uint32(width)
	}
	if h, ok := attrMap["height"]; ok {
		height, e := strconv.Atoi(h)
		if e != nil {
			return nil, fmt.Errorf("height[%s] is illegal:%v", h, e)
		}
		result.Height = uint32(height)
	}
	for key, value := range attrMap {
		if key != "src" && key != "title" && key != "cache" && key != "timeout" && key != "width" && key != "height" {
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

type MessageElementAudio struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Src      string
	Title    string
	Cache    bool
	Timeout  string //ms
	Duration uint32
	Poster   string
}

func (e *MessageElementAudio) Tag() string {
	return "audio"
}

func (e *MessageElementAudio) attrString() string {
	if e == nil {
		return ""
	}
	result := ""
	if e.Src != "" {
		result += ` src="` + escape(e.Src, true) + `"`
	}
	if e.Title != "" {
		result += ` title="` + escape(e.Title, true) + `"`
	}
	if e.Cache {
		result += ` cache`
	}
	if e.Timeout != "" {
		result += ` timeout="` + e.Timeout + `"`
	}
	return result
}

func (e *MessageElementAudio) Stringify() string {
	result := ""
	attrStr := e.attrString()
	if e.Duration > 0 {
		attrStr += fmt.Sprintf(` duration="%d"`, e.Duration)

	}
	if e.Poster != "" {
		attrStr += ` poster="` + escape(e.Poster, true) + `"`
	}
	result += attrStr
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementAudio) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementAudio{
		Src:     attrMap["src"],
		Title:   attrMap["title"],
		Cache:   false,
		Timeout: attrMap["timeout"],
	}
	cacheAttr, ok := attrMap["cache"]
	if ok {
		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
	}
	if d, ok := attrMap["duration"]; ok {
		duration, e := strconv.Atoi(d)
		if e != nil {
			return nil, fmt.Errorf("duration[%s] is illegal:%v", d, e)
		}
		result.Duration = uint32(duration)
	}
	if p, ok := attrMap["poster"]; ok {
		result.Poster = p
	}
	for key, value := range attrMap {
		if key != "src" && key != "title" && key != "cache" && key != "timeout" && key != "duration" && key != "poster" {
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

type MessageElementVideo struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Src      string
	Title    string
	Cache    bool
	Timeout  string //ms
	Width    uint32
	Height   uint32
	Duration uint32
	Poster   string
}

func (e *MessageElementVideo) Tag() string {
	return "video"
}

func (e *MessageElementVideo) attrString() string {
	if e == nil {
		return ""
	}
	result := ""
	if e.Src != "" {
		result += ` src="` + escape(e.Src, true) + `"`
	}
	if e.Title != "" {
		result += ` title="` + escape(e.Title, true) + `"`
	}
	if e.Cache {
		result += ` cache`
	}
	if e.Timeout != "" {
		result += ` timeout="` + e.Timeout + `"`
	}
	return result
}

func (e *MessageElementVideo) Stringify() string {
	result := ""
	attrStr := e.attrString()
	if e.Width > 0 {
		attrStr += fmt.Sprintf(` width="%d"`, e.Width)
	}
	if e.Height > 0 {
		attrStr += fmt.Sprintf(` height="%d"`, e.Height)
	}
	if e.Duration > 0 {
		attrStr += fmt.Sprintf(` duration="%d"`, e.Duration)
	}
	if e.Poster != "" {
		attrStr += ` poster="` + escape(e.Poster, true) + `"`
	}
	result += attrStr
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementVideo) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementVideo{
		Src:     attrMap["src"],
		Title:   attrMap["title"],
		Cache:   false,
		Timeout: attrMap["timeout"],
	}
	cacheAttr, ok := attrMap["cache"]
	if ok {
		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
	}
	if w, ok := attrMap["width"]; ok {
		width, e := strconv.Atoi(w)
		if e != nil {
			return nil, fmt.Errorf("width[%s] is illegal:%v", w, e)
		}
		result.Width = uint32(width)
	}
	if h, ok := attrMap["height"]; ok {
		height, e := strconv.Atoi(h)
		if e != nil {
			return nil, fmt.Errorf("height[%s] is illegal:%v", h, e)
		}
		result.Height = uint32(height)
	}
	if d, ok := attrMap["duration"]; ok {
		duration, e := strconv.Atoi(d)
		if e != nil {
			return nil, fmt.Errorf("duration[%s] is illegal:%v", d, e)
		}
		result.Duration = uint32(duration)
	}
	if p, ok := attrMap["poster"]; ok {
		result.Poster = p
	}
	for key, value := range attrMap {
		if key != "src" && key != "title" && key != "cache" && key != "timeout" && key != "width" && key != "height" && key != "duration" && key != "poster" {
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

type MessageElementFile struct {
	*noAliasMessageElement
	*ChildrenMessageElement
	*ExtendAttributes
	Src     string
	Title   string
	Cache   bool
	Timeout string //ms
	Poster  string
}

func (e *MessageElementFile) Tag() string {
	return "file"
}
func (e *MessageElementFile) attrString() string {
	if e == nil {
		return ""
	}
	result := ""
	if e.Src != "" {
		result += ` src="` + escape(e.Src, true) + `"`
	}
	if e.Title != "" {
		result += ` title="` + escape(e.Title, true) + `"`
	}
	if e.Cache {
		result += ` cache`
	}
	if e.Timeout != "" {
		result += ` timeout="` + e.Timeout + `"`
	}
	return result
}
func (e *MessageElementFile) Stringify() string {
	result := ""
	attrStr := e.attrString()
	if e.Poster != "" {
		attrStr += ` poster="` + escape(e.Poster, true) + `"`
	}
	result += attrStr
	result += e.stringifyAttributes()
	childrenStr := e.stringifyChildren()
	if childrenStr == "" {
		return `<` + e.Tag() + result + `/>`
	}
	return `<` + e.Tag() + result + `>` + childrenStr + `</` + e.Tag() + `>`
}

func (e *MessageElementFile) Parse(n *html.Node) (MessageElement, error) {
	attrMap := attrList2MapVal(n.Attr)
	result := &MessageElementFile{
		Src:     attrMap["src"],
		Title:   attrMap["title"],
		Cache:   false,
		Timeout: attrMap["timeout"],
	}
	cacheAttr, ok := attrMap["cache"]
	if ok {
		result.Cache = cacheAttr == "" || cacheAttr == "true" || cacheAttr == "1"
	}
	if p, ok := attrMap["poster"]; ok {
		result.Poster = p
	}
	for key, value := range attrMap {
		if key != "src" && key != "title" && key != "cache" && key != "timeout" && key != "poster" {
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
	RegsiterParserElement(&MessageElementImg{})
	RegsiterParserElement(&MessageElementAudio{})
	RegsiterParserElement(&MessageElementVideo{})
	RegsiterParserElement(&MessageElementFile{})
}
