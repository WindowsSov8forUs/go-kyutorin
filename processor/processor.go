package processor

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
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

// ResumeEvents 恢复事件
func (q *EventQueue) ResumeEvents(Sequence int64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	var events []*signaling.Event
	var isFound bool
	for _, event := range q.Events {
		if event.Id == Sequence {
			isFound = true
		}
		if isFound {
			events = append(events, event)
		}
	}
	q.Events = events
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

	// 获取 id
	id, err := HashEventID(payload.ID)
	if err != nil {
		log.Errorf("计算事件 ID 时出错: %v", err)
		return err
	}

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
		Id:        id,
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

	// 获取 id
	id, err := HashEventID(payload.ID)
	if err != nil {
		log.Errorf("计算事件 ID 时出错: %v", err)
		return err
	}

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
		Id:        id,
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

	// 获取 id
	id, err := HashEventID(data.ID)
	if err != nil {
		log.Errorf("计算事件 ID 时出错: %v", err)
		return err
	}

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
		Id:        id,
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

// getMessageLog 获取消息日志
func getMessageLog(data interface{}) string {
	// 强制类型转换获取 Message 结构
	var msg *dto.Message
	var isAt bool = false // 是否为 at 消息
	switch v := data.(type) {
	case *dto.WSGroupATMessageData:
		msg = (*dto.Message)(v)
		isAt = true
	case *dto.WSATMessageData:
		msg = (*dto.Message)(v)
	case *dto.WSMessageData:
		msg = (*dto.Message)(v)
	case *dto.WSDirectMessageData:
		msg = (*dto.Message)(v)
	case *dto.WSC2CMessageData:
		msg = (*dto.Message)(v)
	case *dto.Message:
		msg = v
	default:
		return ""
	}
	var messageString string

	// 使用正则表达式查找特殊格式字符
	re := regexp.MustCompile(`(@everyone|<@!\d+>|<#\d+>|<emoji:\d+>)`)

	// 获取所有匹配项的位置
	indexes := re.FindAllStringIndex(msg.Content, -1)

	// 根据匹配项的位置分割字符串
	var result []string
	start := 0
	for _, index := range indexes {
		if start != index[0] {
			part := msg.Content[start:index[0]]
			if part != "" {
				result = append(result, part)
			}
		}
		result = append(result, msg.Content[index[0]:index[1]])
		start = index[1]
	}
	if start != len(msg.Content) {
		part := msg.Content[start:]
		if part != "" {
			result = append(result, part)
		}
	}

	// 匹配检查每个结果
	for _, r := range result {
		if r == "@everyone" {
			if msg.MentionEveryone {
				messageString += "@全体成员"
			}
		} else if strings.HasPrefix(r, "<@!") && strings.HasSuffix(r, ">") {
			// 提取 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<@!")
			for _, mention := range msg.Mentions {
				if mention.ID == id {
					messageString += "@" + mention.Username
					break
				}
			}
		} else if strings.HasPrefix(r, "<#") && strings.HasSuffix(r, ">") {
			// 提取频道 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<#")
			messageString += "#" + id
		} else if strings.HasPrefix(r, "<emoji:") && strings.HasSuffix(r, ">") {
			// 提取 emoji ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<emoji:")
			messageString += ":emoji:" + id + ":"
		} else {
			// 普通文本
			messageString += r
		}
	}

	// 处理 Attachments 字段
	for _, attachment := range msg.Attachments {
		if attachment == nil {
			continue
		}
		// 根据 ContentType 前缀判断文件类型
		switch {
		case strings.HasPrefix(attachment.ContentType, "image"):
			image := "[图片]"
			if strings.HasPrefix(attachment.URL, "http") {
				image += "(" + attachment.URL + ")"
			} else {
				image += "(https://" + attachment.URL + ")"
			}
			messageString += image
		case strings.HasPrefix(attachment.ContentType, "audio"):
			audio := "[语音]"
			if strings.HasPrefix(attachment.URL, "http") {
				audio += "(" + attachment.URL + ")"
			} else {
				audio += "(https://" + attachment.URL + ")"
			}
			messageString += audio
		case strings.HasPrefix(attachment.ContentType, "video"):
			video := "[视频]"
			if strings.HasPrefix(attachment.URL, "http") {
				video += "(" + attachment.URL + ")"
			} else {
				video += "(https://" + attachment.URL + ")"
			}
			messageString += video
		default:
			file := "[文件]"
			if strings.HasPrefix(attachment.URL, "http") {
				file += "(" + attachment.URL + ")"
			} else {
				file += "(https://" + attachment.URL + ")"
			}
			messageString += file
		}
	}

	// 添加 embed 消息
	for _, embed := range msg.Embeds {
		if embed == nil {
			continue
		}
		messageString += fmt.Sprintf("[embed](%s)", embed.Title)
	}

	// 添加 ark 消息
	if msg.Ark != nil {
		messageString += fmt.Sprintf("[ark](%d)", msg.Ark.TemplateID)
	}

	// 添加消息回复
	if msg.MessageReference != nil {
		messageString = "[回复消息]" + "(" + msg.MessageReference.MessageID + ")" + messageString
	}

	// 添加消息前 at
	if isAt {
		bot := handlers.GetBot("qq") // 获取 qq 平台机器人实例
		if bot != nil {
			messageString = "@" + bot.Name + messageString
		}
	}

	return messageString
}

// HashEventID 计算事件 ID
func HashEventID(payloadId string) (int64, error) {
	h := fnv.New64a()
	_, err := io.WriteString(h, payloadId)
	if err != nil {
		return 0, err
	}
	return int64(h.Sum64()), nil
}
