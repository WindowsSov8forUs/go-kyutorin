package processor

import (
	"strconv"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/tencent-connect/botgo/dto"
)

// ProcessChannelEvent 处理频道事件
func (p *Processor) ProcessChannelEvent(payload *dto.WSPayload, data *dto.WSChannelData) error {
	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := payload.S

	// 获取当前时间作为时间戳
	t := time.Now()

	// 填充事件数据
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      signaling.EVENT_TYPE_INTERNAL,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t.Unix(),
		Type_:     string(payload.Type),
		Data_:     data,
	}

	// 发送事件
	return p.BroadcastEvent(event)
}
