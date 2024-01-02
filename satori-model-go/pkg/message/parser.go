package message

import (
	"strings"
)

type messageElementParserFunc func(n *Node) (MessageElement, error)

type messageElementParser interface {
	Tag() string
	Alias() []string
	parse(n *Node) (MessageElement, error)
}

type parsersStruct struct {
	_storage map[string]messageElementParserFunc
}

func (parsers *parsersStruct) set(tag string, parseFunc messageElementParserFunc) {
	parsers._storage[tag] = parseFunc
}

func (parsers *parsersStruct) get(tag string) (messageElementParserFunc, bool) {
	val, ok := parsers._storage[tag]
	return val, ok
}

var factory = &parsersStruct{
	_storage: make(map[string]messageElementParserFunc),
}

func regsiterParserElement(parser messageElementParser) {
	factory.set(parser.Tag(), parser.parse)
	if len(parser.Alias()) > 0 {
		for _, tag := range parser.Alias() {
			factory.set(tag, parser.parse)
		}
	}

}

func parseNode(n *Node, callback func(e MessageElement)) error {
	// 获取节点解析函数
	parseFunc, ok := factory.get(n.Type)
	if ok {
		// 解析节点
		element, err := parseFunc(n)
		if err != nil {
			return err
		}
		callback(element)
	} else {
		// 尝试解析自定义节点
		element, err := parseCustomNode(n)
		if err != nil {
			return err
		}
		if element != nil {
			callback(element)
		}
	}
	return nil
}

func parseChildrenNode(n *Node, callback func(e MessageElement)) error {
	for _, c := range n.Children {
		err := parseNode(c, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func Parse(source string) ([]MessageElement, error) {
	nodes := parse(source)
	var result []MessageElement
	for _, node := range nodes {
		err := parseNode(node, func(e MessageElement) {
			if e != nil {
				result = append(result, e)
			}
		})
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func Stringify([]MessageElement) (string, error) {
	return "", nil
}

func Escape(source string) string {
	result := strings.ReplaceAll(source, "&", "&amp;")
	result = strings.ReplaceAll(result, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")
	result = strings.ReplaceAll(result, "\"", "&quot;")
	return result
}

func Unescape(source string) string {
	result := strings.ReplaceAll(source, "&amp;", "&")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&quot;", "\"")
	return result
}
