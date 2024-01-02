package message

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	elements, _ := Parse(
		`<passive id="08f9e7ba8f86d39e251090b5b4ae02385a48d1bbc9ac06"/><at id="1862902977615371284"/><image src="https://gchat.qpic.cn/qmeetpic/635502894002160049/634198672-3078136204-EFA2D473526BC132850E2C486E481246/0"/>`)
	s := ""
	for _, e := range elements {
		s += e.Stringify()
	}
	fmt.Println(s)
}

func TestStringify(t *testing.T) {
	elements, _ := Parse(
		`我是纯文本<strong>我是加粗文本<b>套娃<b>套娃中的套娃</b></b>我是123</strong>`)
	s := ""
	for _, e := range elements {
		s += e.Stringify()
	}
	fmt.Println(s)
	var newS string
	newElements, _ := Parse(s)
	for _, e := range newElements {
		newS += e.Stringify()
	}
	fmt.Println(newS)
}
