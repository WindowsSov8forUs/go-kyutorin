package processor

import (
	"time"

	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/operation"
	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
)

// ReadyHandler 处理 Ready 事件
func ReadyHandler(p *Processor) event.ReadyHandler {
	return func(event *dto.Payload, data *dto.WSReadyData) {
		log.Info("连接成功！")
		SetStatus("qq", login.StatusOnline)
		SetStatus("qqguild", login.StatusOnline)

		// 构建 qq 事件
		id := SaveEventID(data.SessionID)

		satoriEvent := &operation.Event{
			Sn:        id,
			Type:      operation.EventTypeLoginUpdated,
			Timestamp: time.Now().UnixMilli(),
			Login:     buildLoginEventLogin("qq"),
		}

		// 构建 qqguild 事件
		id = SaveEventID(data.SessionID)

		satoriEventGuild := &operation.Event{
			Sn:        id,
			Type:      operation.EventTypeLoginUpdated,
			Timestamp: time.Now().UnixMilli(),
			Login:     buildLoginEventLogin("qqguild"),
		}

		p.BroadcastEvent(satoriEvent)
		p.BroadcastEvent(satoriEventGuild)
	}
}

// ErrorNotifyHandler 处理错误通知事件
func ErrorNotifyHandler(p *Processor) event.ErrorNotifyHandler {
	return func(err error) {
		log.Errorf("QQ 开放平台连接出现错误：%v", err)
		SetStatus("qq", login.StatusOffline)
		SetStatus("qqguild", login.StatusOffline)

		// 构建 qq 事件
		id := SaveEventID(err.Error())

		satoriEvent := &operation.Event{
			Sn:        id,
			Type:      operation.EventTypeLoginUpdated,
			Timestamp: time.Now().UnixMilli(),
			Login:     buildLoginEventLogin("qq"),
		}

		// 构建 qqguild 事件
		id = SaveEventID(err.Error())

		satoriEventGuild := &operation.Event{
			Sn:        id,
			Type:      operation.EventTypeLoginUpdated,
			Timestamp: time.Now().UnixMilli(),
			Login:     buildLoginEventLogin("qqguild"),
		}

		p.BroadcastEvent(satoriEvent)
		p.BroadcastEvent(satoriEventGuild)
	}
}

// HelloHandler 处理 Hello 事件
func HelloHandler(p *Processor) event.HelloHandler {
	return func(event *dto.Payload) {
		data := event.Data.(*dto.WSHelloData)
		log.Infof("成功与 QQ 开放平台建立 WebSocket 连接，心跳周期：%v", data.HeartbeatInterval)
	}
}

// ReconnectHandler 处理重新连接事件
func ReconnectHandler(p *Processor) event.ReconnectHandler {
	return func(event *dto.Payload) {
		log.Info("正在尝试重新连接 QQ 开放平台...")
		SetStatus("qq", login.StatusReconnect)
		SetStatus("qqguild", login.StatusReconnect)
	}
}

// PlainEventHandler 处理透传 handler
func PlainEventHandler(p *Processor) event.PlainEventHandler {
	return func(event *dto.Payload, message []byte) error {
		// 默认为 qqguild
		return p.ProcessQQGuildInternal(event, message)
	}
}

// AudioEventHandler 音频机器人事件 handler
func AudioEventHandler(p *Processor) event.AudioEventHandler {
	return func(event *dto.Payload, data *dto.AudioData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// InteractionHandler 处理内联交互事件
func InteractionHandler(p *Processor) event.InteractionEventHandler {
	return func(event *dto.Payload, data *dto.InteractionEventData) error {
		return p.ProcessInteractionEvent(data)
	}
}

// ThreadEventHandler 处理论坛主题事件
func ThreadEventHandler(p *Processor) event.ThreadEventHandler {
	return func(event *dto.Payload, data *dto.ThreadData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// PostEventHandler 处理论坛回帖事件
func PostEventHandler(p *Processor) event.PostEventHandler {
	return func(event *dto.Payload, data *dto.PostData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// ReplyEventHandler 处理论坛帖子回复事件
func ReplyEventHandler(p *Processor) event.ReplyEventHandler {
	return func(event *dto.Payload, data *dto.ReplyData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// ForumAuditEventHandler 处理论坛帖子审核事件
func ForumAuditEventHandler(p *Processor) event.ForumAuditEventHandler {
	return func(event *dto.Payload, data *dto.ForumAuditData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// GuildEventHandler 处理频道事件
func GuildEventHandler(p *Processor) event.GuildEventHandler {
	return func(event *dto.Payload, data *dto.GuildData) error {
		return p.ProcessGuildEvent(event, data)
	}
}

// MemberEventHandler 处理成员变更事件
func MemberEventHandler(p *Processor) event.GuildMemberEventHandler {
	return func(event *dto.Payload, data *dto.GuildMemberData) error {
		return p.ProcessMemberEvent(event, data)
	}
}

// ChannelEventHandler 处理子频道事件
func ChannelEventHandler(p *Processor) event.ChannelEventHandler {
	return func(event *dto.Payload, data *dto.ChannelData) error {
		return p.ProcessChannelEvent(event, data)
	}
}

// CreateMessageHandler 处理消息事件 私域的事件 不 at 信息
func CreateMessageHandler(p *Processor) event.MessageEventHandler {
	return func(event *dto.Payload, data *dto.MessageData) error {
		return p.ProcessGuildNormalMessage(event, data)
	}
}

// ATMessageEventHandler 实现处理 频道 at 消息的回调
func ATMessageEventHandler(p *Processor) event.ATMessageEventHandler {
	return func(event *dto.Payload, data *dto.ATMessageData) error {
		return p.ProcessGuildATMessage(event, data)
	}
}

// DirectMessageHandler 处理私信事件
func DirectMessageHandler(p *Processor) event.DirectMessageEventHandler {
	return func(event *dto.Payload, data *dto.DirectMessageData) error {
		return p.ProcessChannelDirectMessage(event, data)
	}
}

// MessageDeleteEventHandler 处理私域消息删除事件
func MessageDeleteEventHandler(p *Processor) event.MessageDeleteEventHandler {
	return func(event *dto.Payload, data *dto.MessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// PublicMessageDeleteEventHandler 处理公域消息删除事件
func PublicMessageDeleteEventHandler(p *Processor) event.PublicMessageDeleteEventHandler {
	return func(event *dto.Payload, data *dto.PublicMessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// DirectMessageDeleteEventHandler 处理私聊消息删除事件
func DirectMessageDeleteEventHandler(p *Processor) event.DirectMessageDeleteEventHandler {
	return func(event *dto.Payload, data *dto.DirectMessageDeleteData) error {
		return p.ProcessMessageDelete(event, data)
	}
}

// MessageReactionEventHandler 处理表情表态事件
func MessageReactionEventHandler(p *Processor) event.MessageReactionEventHandler {
	return func(event *dto.Payload, data *dto.MessageReactionData) error {
		return p.ProcessMessageReaction(event, data)
	}
}

// MessageAuditEventHandler 处理消息审核事件
func MessageAuditEventHandler(p *Processor) event.MessageAuditEventHandler {
	return func(event *dto.Payload, data *dto.MessageAuditData) error {
		// TODO: 专门的处理函数
		return p.ProcessQQGuildInternal(event, data)
	}
}

// GroupATMessageEventHandler 实现处理 群 at 消息的回调
func GroupATMessageEventHandler(p *Processor) event.GroupATMessageEventHandler {
	return func(event *dto.Payload, data *dto.GroupATMessageData) error {
		return p.ProcessGroupMessage(event, data)
	}
}

// GroupAddRobotEventHandler 实现处理 群添加机器人的回调
func GroupAddRobotEventHandler(p *Processor) event.GroupAddRobotEventHandler {
	return func(event *dto.Payload, data *dto.GroupAddBotEvent) error {
		return p.ProcessGroupAddRobot(event, data)
	}
}

// GroupDelRobotEventHandler 实现处理 群删除机器人的回调
func GroupDelRobotEventHandler(p *Processor) event.GroupDelRobotEventHandler {
	return func(event *dto.Payload, data *dto.GroupAddBotEvent) error {
		return p.ProcessGroupDelRobot(event, data)
	}
}

// C2CMessageEventHandler 实现处理私聊消息的回调
func C2CMessageEventHandler(p *Processor) event.C2CMessageEventHandler {
	return func(event *dto.Payload, data *dto.C2CMessageData) error {
		return p.ProcessC2CMessage(event, data)
	}
}

func (p *Processor) getHandlersByName(intentName string) ([]interface{}, bool) {
	switch intentName {
	case "DEFAULT": // 默认处理函数
		handlers := []interface{}{
			ReadyHandler(p),
			ErrorNotifyHandler(p),
			PlainEventHandler(p),
		}
		return handlers, true
	case "GUILDS": // 频道事件
		handlers := []interface{}{
			GuildEventHandler(p),
			ChannelEventHandler(p),
		}
		return handlers, true
	case "GUILD_MEMBERS": // 频道成员事件
		handlers := []interface{}{MemberEventHandler(p)}
		return handlers, true
	case "GUILD_MESSAGES": // 私域频道消息事件
		handlers := []interface{}{
			CreateMessageHandler(p),
			MessageDeleteEventHandler(p),
		}
		return handlers, true
	case "GUILD_MESSAGE_REACTIONS": // 频道消息表情表态事件
		handlers := []interface{}{MessageReactionEventHandler(p)}
		return handlers, true
	case "DIRECT_MESSAGE": // 频道私信事件
		handlers := []interface{}{
			DirectMessageHandler(p),
			DirectMessageDeleteEventHandler(p),
		}
		return handlers, true
	case "OPEN_FORUMS_EVENT": // 公域论坛事件
		return nil, true
	case "AUDIO_OR_LIVE_CHANNEL_MEMBER": // 音频或直播频道成员事件
		return nil, true
	case "USER_MESSAGES": // 单聊/群聊消息事件
		handlers := []interface{}{
			GroupATMessageEventHandler(p),
			GroupAddRobotEventHandler(p),
			GroupDelRobotEventHandler(p),
			C2CMessageEventHandler(p),
		}
		return handlers, true
	case "INTERACTION": // 互动事件
		handlers := []interface{}{InteractionHandler(p)}
		return handlers, true
	case "MESSAGE_AUDIT": // 消息审核事件
		handlers := []interface{}{MessageAuditEventHandler(p)}
		return handlers, true
	case "FORUMS_EVENT": // 私域论坛事件
		handlers := []interface{}{
			ThreadEventHandler(p),
			PostEventHandler(p),
			ReplyEventHandler(p),
			ForumAuditEventHandler(p),
		}
		return handlers, true
	case "AUDIO_ACTION": // 音频机器人事件
		handlers := []interface{}{AudioEventHandler(p)}
		return handlers, true
	case "PUBLIC_GUILD_MESSAGES": // 公域频道消息事件
		handlers := []interface{}{
			ATMessageEventHandler(p),
			PublicMessageDeleteEventHandler(p),
		}
		return handlers, true
	default:
		log.Warnf("未知的 Intents : %s\n", intentName)
		return nil, false
	}
}

func (p *Processor) getWebHookAvailableHandlers() ([]interface{}, bool) {
	handlers := []interface{}{
		ErrorNotifyHandler(p),
		PlainEventHandler(p),
		GuildEventHandler(p),
		ChannelEventHandler(p),
		MemberEventHandler(p),
		DirectMessageHandler(p),
		DirectMessageDeleteEventHandler(p),
		GroupATMessageEventHandler(p),
		GroupAddRobotEventHandler(p),
		GroupDelRobotEventHandler(p),
		C2CMessageEventHandler(p),
		InteractionHandler(p),
		MessageAuditEventHandler(p),
		ThreadEventHandler(p),
		PostEventHandler(p),
		ReplyEventHandler(p),
		ForumAuditEventHandler(p),
		AudioEventHandler(p),
		ATMessageEventHandler(p),
		PublicMessageDeleteEventHandler(p),
	}
	return handlers, true
}
