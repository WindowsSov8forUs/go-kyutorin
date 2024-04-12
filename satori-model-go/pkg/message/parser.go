package message

import (
	"strings"

	"golang.org/x/net/html"
)

type messageElementParserFunc func(n *html.Node) (MessageElement, error)

type MessageElementParser interface {
	Tag() string
	Alias() []string
	Parse(n *html.Node) (MessageElement, error)
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

func attrList2MapVal(attrs []html.Attribute) map[string]string {
	var result = make(map[string]string)
	for _, attr := range attrs {
		result[attr.Key] = attr.Val
	}
	return result
}

var factory = &parsersStruct{
	_storage: make(map[string]messageElementParserFunc),
}

func RegsiterParserElement(parser MessageElementParser) {
	factory.set(parser.Tag(), parser.Parse)
	if len(parser.Alias()) > 0 {
		for _, tag := range parser.Alias() {
			factory.set(tag, parser.Parse)
		}
	}

}

func parseHtmlNode(n *html.Node, callback func(e MessageElement)) error {
	parsed := false
	if n.Type == html.ElementNode {
		var parserOfTagFunc messageElementParserFunc
		parserOfTagFunc, parsed = factory.get(n.Data)
		if parsed {
			e, err := parserOfTagFunc(n)
			if err != nil {
				return err
			}
			callback(e)
		} else {
			e, err := ExtendParser.Parse(n)
			if err != nil {
				return err
			}
			callback(e)
			parsed = true
		}
	} else if n.Type == html.TextNode {
		content := strings.TrimSpace(n.Data)
		if content != "" {
			callback(&MessageElementText{
				Content: content,
			})
		}
		parsed = true
	}
	if !parsed {
		parseHtmlChildrenNode(n, callback)
	}
	return nil
}
func parseHtmlChildrenNode(n *html.Node, callback func(e MessageElement)) error {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		err := parseHtmlNode(c, callback)
		if err != nil {
			return err
		}
	}
	return nil
}

func Parse(source string) ([]MessageElement, error) {
	doc := xhtmlParse(source)
	var result []MessageElement
	err := parseHtmlNode(doc, func(e MessageElement) {
		if e != nil {
			result = append(result, e)
		}
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Stringify(elements []MessageElement) (string, error) {
	if len(elements) == 0 {
		return "", nil
	}
	result := ""
	for _, e := range elements {
		result += e.Stringify()
	}
	return result, nil

}
