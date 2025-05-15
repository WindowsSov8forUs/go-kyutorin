package websocket

import (
	"github.com/tencent-connect/botgo/dto"
)

// ATMessageEventHandler 定义了一个处理 频道 AT 消息的事件处理程序。
type ATMessageEventHandler func(event *dto.Payload, data *dto.WSATMessageData) error

// MessageEventHandler 定义了一个处理频道普通消息的事件处理程序。
type MessageEventHandler func(event *dto.Payload, data *dto.WSMessageData) error
