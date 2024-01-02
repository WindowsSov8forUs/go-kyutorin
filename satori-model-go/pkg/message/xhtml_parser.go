package message

import (
	"regexp"
	"strings"
)

var (
	tagPat  = regexp.MustCompile(`<!--[\s\S]*?-->|<(/?)([^!\s>/]*)([^>]*?)\s*(/?)>`)
	attrPat = regexp.MustCompile(`([^\s=]+)(?:="([^"]*)"|='([^']*)')?`)
)

// 消息元素节点
type Node struct {
	Type     string            // 节点类型
	Attrs    map[string]string // 节点属性
	Children []*Node           // 子节点
	Source   string            // 原始字符串
}

func (n *Node) String() string {
	if n.Source != "" {
		return n.Source
	}
	if n.Type == "text" {
		return Escape(n.Attrs["content"])
	}

	var attr = func(key, value string) string {
		if value == "true" {
			return key
		}
		if value == "false" {
			return "no-" + key
		}
		return key + "=" + Escape(value)
	}

	var attrs []string
	for key, value := range n.Attrs {
		attrs = append(attrs, attr(key, value))
	}
	attrStr := strings.Join(attrs, " ")
	if len(n.Children) == 0 {
		return "<" + n.Type + " " + attrStr + "/>"
	} else {
		var children []string
		for _, child := range n.Children {
			children = append(children, child.String())
		}
		return "<" + n.Type + " " + attrStr + ">" + strings.Join(children, "") + "</" + n.Type + ">"
	}
}

// 消息元素标签
type Token struct {
	Type   string            // 标签类型
	Close  bool              // 是否闭合
	Empty  bool              // 是否空标签
	Attrs  map[string]string // 标签属性
	Source string            // 原始字符串
}

// parse 将字符串解析为消息元素节点列表
func parse(source string) []*Node {
	var tokens []*Token

	// 负责将字符串转换为文本对象的函数
	var parseToText = func(text string) {
		text = Unescape(text)
		if text != "" {
			token := &Token{
				Type:  "text",
				Close: true,
				Attrs: map[string]string{
					"content": text,
				},
			}
			tokens = append(tokens, token)
		}
	}

	// 匹配标签并循环处理
	for {
		// 不断匹配直到不再能匹配到标签
		tagLoc := tagPat.FindStringIndex(source)
		if tagLoc == nil {
			break
		}
		matches := tagPat.FindStringSubmatch(source)

		// 将标签前的文本转换为文本对象
		parseToText(source[:tagLoc[0]])
		source = source[tagLoc[1]:]

		// 如果是注释则跳过
		if strings.HasPrefix(source, "<!--") {
			continue
		}

		// 根据匹配结果创建标签对象
		close, tag, attrStr, empty := matches[1], matches[2], matches[3], matches[4]
		if tag == "" {
			tag = "body"
		}
		token := &Token{
			Type:   tag,
			Close:  close != "",
			Empty:  empty != "",
			Attrs:  make(map[string]string),
			Source: matches[0],
		}

		for {
			// 匹配 HTML 属性并循环处理
			attrLoc := attrPat.FindStringIndex(attrStr)
			if attrLoc == nil {
				break
			}
			matches := attrPat.FindStringSubmatch(attrStr)

			// 获取有效属性值
			key, value1, value2 := matches[1], matches[2], matches[3]
			value := value1
			if value == "" {
				value = value2
			}
			if value != "" {
				token.Attrs[key] = Unescape(value)
			} else if strings.HasPrefix(key, "no-") {
				token.Attrs[key] = "false"
			} else {
				token.Attrs[key] = "true"
			}
			attrStr = attrStr[attrLoc[1]:]
		}

		// 将标签对象添加到标签列表
		tokens = append(tokens, token)
	}

	// 将最后的文本转换为文本对象
	parseToText(source)

	// 定义一个 Element 对象的栈，并创建一个类型为 body 的根元素
	var stack []*Node
	root := &Node{
		Type:  "body",
		Attrs: make(map[string]string),
	}
	stack = append(stack, root)

	// 一个回滚栈中一定元素数并将其添加到上一层元素子元素中的函数
	var rollback = func(count int) {
		for count > 0 {
			child := stack[0]
			stack = stack[1:]
			source := stack[0].Children[len(stack[0].Children)-1]
			stack[0].Children = stack[0].Children[:len(stack[0].Children)-1]
			stack[0].Children = append(stack[0].Children, &Node{
				Type:  "text",
				Attrs: map[string]string{"content": source.String()}})
			stack[0].Children = append(stack[0].Children, child.Children...)
			count--
		}
	}

	// 循环处理标签列表
	for _, token := range tokens {
		if token.Type == "text" {
			// 如果是文本标签则将其添加到栈顶元素的子元素中
			stack[0].Children = append(stack[0].Children, &Node{
				Type:  "text",
				Attrs: token.Attrs,
			})
		} else if token.Close {
			// 如果是闭合标签则查找对应的开放标签并记录索引
			var index int
			for index < len(stack) && stack[index].Type != token.Type {
				index++
			}
			// 如果没有找到
			if index == len(stack) {
				// 作为文本元素处理
				stack[0].Children = append(stack[0].Children, &Node{
					Type: "text",
					Attrs: map[string]string{
						"content": token.Source,
					},
				})
			} else {
				// 回滚处理
				rollback(index)
				// 弹出第一个元素并赋值
				node := stack[0]
				stack = stack[1:]
				node.Source = ""
			}
		} else {
			// 是 Token 且不是关闭标签，创建一个 Node 并添加为第一元素的子元素
			node := &Node{
				Type:  token.Type,
				Attrs: token.Attrs,
			}
			stack[0].Children = append(stack[0].Children, node)
			// 如果不是空标签则将其添加到栈中
			if !token.Empty {
				// 赋值，并将 node 设为第一元素
				node.Source = token.Source
				stack = append([]*Node{node}, stack...)
			}
		}
	}

	// 回滚除最后一个元素以外的所有元素
	rollback(len(stack) - 1)

	// 将根元素的子元素作为解析结果返回
	return root.Children
}
