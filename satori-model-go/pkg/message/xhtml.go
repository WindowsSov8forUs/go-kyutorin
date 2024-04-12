package message

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

var (
	tagPat  = regexp.MustCompile(`(<!--[\s\S]*?-->)|(<(/?)([^!\s>/]*)([^>]*?)\s*(/?)>)`)
	attrPat = regexp.MustCompile(`([^\s=]+)(?:="([^"]*)"|='([^']*)')?`)
)

func escape(text string, inLine bool) string {
	text = strings.Replace(text, "&", "&amp;", -1)
	text = strings.Replace(text, "<", "&lt;", -1)
	text = strings.Replace(text, ">", "&gt;", -1)
	if inLine {
		text = strings.Replace(text, "\"", "&quot;", -1)
	}
	return text
}

func unescape(text string) string {
	text = strings.Replace(text, "&lt;", "<", -1)
	text = strings.Replace(text, "&gt;", ">", -1)
	text = strings.Replace(text, "&quot;", "\"", -1)

	re := regexp.MustCompile(`&#(\d+);`)
	text = re.ReplaceAllStringFunc(text, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if matches[1] == "38" {
			return s
		}
		i, _ := strconv.Atoi(matches[1])
		return fmt.Sprint(i)
	})

	re = regexp.MustCompile("&#x([0-9a-f]+);")
	text = re.ReplaceAllStringFunc(text, func(s string) string {
		matches := re.FindStringSubmatch(s)
		if matches[1] == "26" {
			return s
		}
		i, _ := strconv.ParseInt(matches[1], 16, 32)
		return fmt.Sprint(i)
	})

	re = regexp.MustCompile("&(amp|#38|#x26);")
	text = re.ReplaceAllString(text, "&")

	return text
}

type Token struct {
	*html.Token
	extra string
}

func (t *Token) parseAttributes() []html.Attribute {
	if t.extra == "" {
		return nil
	}
	t.Attr = []html.Attribute{}
	for {
		attrLoc := attrPat.FindStringSubmatchIndex(t.extra)
		if attrLoc == nil {
			break
		}

		matches := attrPat.FindStringSubmatch(t.extra)
		t.extra = t.extra[attrLoc[1]:]

		key := matches[1]
		var value string
		if matches[2] != "" {
			value = matches[2]
		} else {
			value = matches[3]
		}
		if value != "" {
			t.Attr = append(t.Attr, html.Attribute{
				Key: key,
				Val: unescape(value),
			})
		} else if strings.HasPrefix(key, "no-") {
			t.Attr = append(t.Attr, html.Attribute{
				Key: key[3:],
				Val: "false",
			})
		} else {
			t.Attr = append(t.Attr, html.Attribute{
				Key: key,
				Val: "",
			})
		}
	}
	return t.Attr
}

func parseTokens(tokens []Token) *html.Node {
	var stack = []*html.Node{}
	var root *html.Node = &html.Node{
		Type: html.DocumentNode,
		Data: "body",
	}
	stack = append(stack, root)

	for _, token := range tokens {
		switch token.Type {
		case html.TextToken:
			if len(stack) > 0 {
				node := &html.Node{
					Type: html.TextNode,
					Data: token.Data,
				}
				stack[len(stack)-1].AppendChild(node)
			}
		case html.StartTagToken:
			node := &html.Node{
				Type: html.ElementNode,
				Data: token.Data,
				Attr: token.parseAttributes(),
			}
			if len(stack) > 0 {
				stack[len(stack)-1].AppendChild(node)
			} else {
				root = node
			}
			stack = append(stack, node)
		case html.EndTagToken:
			if token.Data == stack[len(stack)-1].Data {
				stack = stack[:len(stack)-1]
			}
		case html.SelfClosingTagToken:
			node := &html.Node{
				Type: html.ElementNode,
				Data: token.Data,
				Attr: token.parseAttributes(),
			}
			if len(stack) > 0 {
				stack[len(stack)-1].AppendChild(node)
			} else {
				root = node
			}
		}
	}
	return root
}

func xhtmlParse(source string) *html.Node {
	var tokens = []Token{}

	var pushText = func(text string) {
		if text != "" {
			tokens = append(tokens, Token{
				Token: &html.Token{
					Type: html.TextToken,
					Data: text,
				},
			})
		}
	}

	var parseContent = func(source string, start, end bool) {
		source = unescape(source)
		if start {
			re := regexp.MustCompile(`^\s*\n\s*`)
			source = re.ReplaceAllString(source, "")
		}
		if end {
			re := regexp.MustCompile(`\s*\n\s*$`)
			source = re.ReplaceAllString(source, "")
		}
		pushText(source)
	}

	for {
		tagLoc := tagPat.FindStringSubmatchIndex(source)
		if tagLoc == nil {
			break
		}

		parseContent(source[:tagLoc[0]], true, true)
		matches := tagPat.FindStringSubmatch(source)
		source = source[tagLoc[1]:]
		close, type_, extra, empty := matches[3], matches[4], matches[5], matches[6]
		if matches[1] != "" { // comment
			continue
		}
		var token Token
		if close == "" && empty == "" { // 开始标记
			token = Token{
				Token: &html.Token{
					Type: html.StartTagToken,
					Data: type_,
				},
				extra: extra,
			}
		} else if close != "" { // 结束标记
			token = Token{
				Token: &html.Token{
					Type: html.EndTagToken,
					Data: type_,
				},
				extra: extra,
			}
		} else if empty != "" { // 自闭合标记
			token = Token{
				Token: &html.Token{
					Type: html.SelfClosingTagToken,
					Data: type_,
				},
				extra: extra,
			}
		}
		tokens = append(tokens, token)
	}

	parseContent(source, true, true)
	return parseTokens(tokens)
}
