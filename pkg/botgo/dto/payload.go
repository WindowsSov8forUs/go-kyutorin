package dto

// EventType 事件类型
type EventType string

// Payload websocket 消息结构
type Payload struct {
	PayloadBase
	Data       interface{} `json:"d,omitempty"`
	S          int64       `json:"s,omitempty"`
	ID         string      `json:"id,omitempty"`
	RawMessage []byte      `json:"-"` // 原始的 message 数据
}

// PayloadBase 基础消息结构，排除了 data
type PayloadBase struct {
	OPCode OPCode    `json:"op"`
	Seq    uint32    `json:"s,omitempty"`
	Type   EventType `json:"t,omitempty"`
}
