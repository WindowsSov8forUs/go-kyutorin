package processor

import (
	"encoding/json"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/operation"
	"github.com/tencent-connect/botgo/dto"
)

// ProcessInternal 将 qq 事件转换为 Satori 的 Internal 事件
func (p *Processor) ProcessQQInternal(payload *dto.Payload, data interface{}) error {
	// 构建事件数据
	var event *operation.Event

	// 获取 id
	id := SaveEventID(payload.ID)

	// 将当前时间转换为时间戳
	t := time.Now().UnixMilli()

	// 如果 data 是 []byte ，将其转换为 json.RawMessage
	var data_ json.RawMessage
	switch v := data.(type) {
	case []byte:
		data_ = v
	default:
		data_, _ = json.Marshal(data)
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeInternal,
		Timestamp: t,
		Login:     buildNonLoginEventLogin("qq"),
		Type_:     string(payload.Type),
		Data_:     data_,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

// ProcessInternal 将 qqguild 事件转换为 Satori 的 Internal 事件
func (p *Processor) ProcessQQGuildInternal(payload *dto.Payload, data interface{}) error {
	// 构建事件数据
	var event *operation.Event

	// 获取 id
	id := SaveEventID(payload.ID)

	// 将当前时间转换为时间戳
	t := time.Now().UnixMilli()

	// 如果 data 是 []byte ，将其转换为 json.RawMessage
	var data_ json.RawMessage
	switch v := data.(type) {
	case []byte:
		data_ = v
	default:
		data_, _ = json.Marshal(data)
	}

	// 填充事件数据
	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeInternal,
		Timestamp: t,
		Login:     buildNonLoginEventLogin("qqguild"),
		Type_:     string(payload.Type),
		Data_:     data_,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

// ProcessInteractionEvent 处理交互事件
func (p *Processor) ProcessInteractionEvent(data *dto.InteractionEventData) error {
	// TODO: 目前无法将这个事件与 interaction/button 事件适配

	// 构建事件数据
	var event *operation.Event

	// 获取 id
	id := SaveEventID(data.ID)

	// 以当前时间作为时间戳
	t := time.Now().UnixMilli()

	// 根据不同的 设置不同的 platform
	var platform string
	if data.ChatType == 0 {
		platform = "qqguild"
	} else {
		platform = "qq"
	}

	event = &operation.Event{
		Sn:        id,
		Type:      operation.EventTypeInternal,
		Timestamp: t,
		Login:     buildNonLoginEventLogin(platform),
		Type_:     string(dto.EventInteractionCreate),
		Data_:     data,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}
