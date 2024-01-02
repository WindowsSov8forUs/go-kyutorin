package processor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"

	"github.com/dezhishen/satori-model-go/pkg/login"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

// Processor 消息处理器
type Processor struct {
	Api        openapi.OpenAPI
	Apiv2      openapi.OpenAPI
	WebSocket  callapi.WebSocketServer
	EventQueue *EventQueue
}

// EventQueue 事件队列
type EventQueue struct {
	Events []*signaling.Event
	mu     sync.Mutex
}

// NewEventQueue 创建事件队列
func NewEventQueue() *EventQueue {
	return &EventQueue{
		Events: make([]*signaling.Event, 0),
	}
}

// PushEvent 推送事件
func (q *EventQueue) PushEvent(event *signaling.Event) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Events = append(q.Events, event)
}

// PopEvent 弹出事件
func (q *EventQueue) PopEvent() *signaling.Event {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.Events) == 0 {
		return nil
	}
	event := q.Events[0]
	q.Events = q.Events[1:]
	return event
}

// Clear 清空事件队列
func (q *EventQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Events = make([]*signaling.Event, 0)
}

// GetReadyBody 创建 READY 信令的信令数据
func GetReadyBody() *signaling.ReadyBody {
	var logins []*login.Login
	for platform, bot := range handlers.GetBots() {
		login := &login.Login{
			User:     bot,
			SelfId:   handlers.SelfId,
			Platform: platform,
			Status:   handlers.GetStatus(platform),
		}
		logins = append(logins, login)
	}
	return &signaling.ReadyBody{
		Logins: logins,
	}
}

// NewProcessor 创建消息处理器
func NewProcessor(api openapi.OpenAPI, apiv2 openapi.OpenAPI) *Processor {
	return &Processor{
		Api:        api,
		Apiv2:      apiv2,
		EventQueue: NewEventQueue(),
	}
}

// BroadcastEvent 向 Satori 应用发送事件
func (p *Processor) BroadcastEvent(event *signaling.Event) error {
	var errors []string

	// 构建 WebSocket 信令
	sgnl := &signaling.Signaling{
		Op:   signaling.SIGNALING_EVENT,
		Body: (*signaling.EventBody)(event),
	}
	// 转换为 []byte
	data, err := json.Marshal(sgnl)
	if err != nil {
		errors = append(errors, fmt.Sprintf("转换信令时出错: %v", err))
	} else {
		// 判断 WebSocket 服务器是否已建立
		if p.WebSocket != nil {
			// 发送
			if err := p.WebSocket.SendMessage(data); err != nil {
				errors = append(errors, fmt.Sprintf("发送信令时出错: %v", err))
				p.EventQueue.PushEvent(event)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}

	return nil
}

// ProcessInternal 将 qq 事件转换为 Satori 的 Internal 事件
func (p *Processor) ProcessQQInternal(payload *dto.WSPayload, data interface{}) error {
	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := payload.S

	// 将当前时间转换为时间戳
	t := time.Now().Unix()

	// 如果 data 是 []byte ，将其转换为 json.RawMessage
	var data_ json.RawMessage
	switch v := data.(type) {
	case []byte:
		data_ = v
	default:
		data_, _ = json.Marshal(data)
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      signaling.EVENT_TYPE_INTERNAL,
		Platform:  "qq",
		SelfId:    handlers.SelfId,
		Timestamp: t,
		Type_:     string(payload.Type),
		Data_:     data_,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

// ProcessInternal 将 qqguild 事件转换为 Satori 的 Internal 事件
func (p *Processor) ProcessQQGuildInternal(payload *dto.WSPayload, data interface{}) error {
	// 构建事件数据
	var event *signaling.Event

	// 获取 s
	id := payload.S

	// 将当前时间转换为时间戳
	t := time.Now().Unix()

	// 如果 data 是 []byte ，将其转换为 json.RawMessage
	var data_ json.RawMessage
	switch v := data.(type) {
	case []byte:
		data_ = v
	default:
		data_, _ = json.Marshal(data)
	}

	// 填充事件数据
	event = &signaling.Event{
		Id:        strconv.FormatInt(id, 10),
		Type:      signaling.EVENT_TYPE_INTERNAL,
		Platform:  "qqguild",
		SelfId:    handlers.SelfId,
		Timestamp: t,
		Type_:     string(payload.Type),
		Data_:     data_,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}

// ProcessInteractionEvent 处理交互事件
func (p *Processor) ProcessInteractionEvent(data *dto.WSInteractionData) error {
	// TODO: 目前无法将这个事件与 interaction/button 事件适配

	// 构建事件数据
	var event *signaling.Event

	// 以当前时间作为时间戳
	t := time.Now()

	// 根据不同的 设置不同的 platform
	var platform string
	if data.ChatType == 0 {
		platform = "qqguild"
	} else {
		platform = "qq"
	}

	event = &signaling.Event{
		Id:        data.ID,
		Type:      signaling.EVENT_TYPE_INTERNAL,
		Platform:  platform,
		SelfId:    handlers.SelfId,
		Timestamp: t.Unix(),
		Type_:     string(dto.EventInteractionCreate),
		Data_:     data,
	}

	// 上报消息到 Satori 应用
	return p.BroadcastEvent(event)
}
