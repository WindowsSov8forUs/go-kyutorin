package message

import (
	"testing"
)

func Test(t *testing.T) {
	for group, messages := range _getRawMessage() {
		t.Logf("start test group: %s", group)
		for _, message := range messages {
			elements, err := Parse(message)
			if err != nil {
				t.Fatalf("%s Parse error: %s", message, err)
			}
			result, err := Stringify(elements)
			if err != nil {
				t.Fatalf("%s Stringify error: %s", elements, err)
			}
			if result != message {
				t.Fatalf("%s not eq %s", result, message)
			}
		}
	}
}
