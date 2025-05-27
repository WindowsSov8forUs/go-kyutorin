package event

import (
	"github.com/tencent-connect/botgo/dto"
)

// DefaultHandlers 默认的 handler 结构，管理所有支持的 handler 类型
var DefaultHandlers struct {
	Ready       ReadyHandler
	ErrorNotify ErrorNotifyHandler
	Plain       PlainEventHandler

	Hello     HelloHandler
	Reconnect ReconnectHandler

	Guild       GuildEventHandler
	GuildMember GuildMemberEventHandler
	Channel     ChannelEventHandler

	Message             MessageEventHandler
	MessageReaction     MessageReactionEventHandler
	ATMessage           ATMessageEventHandler
	DirectMessage       DirectMessageEventHandler
	MessageAudit        MessageAuditEventHandler
	MessageDelete       MessageDeleteEventHandler
	PublicMessageDelete PublicMessageDeleteEventHandler
	DirectMessageDelete DirectMessageDeleteEventHandler

	Audio AudioEventHandler

	Thread     ThreadEventHandler
	Post       PostEventHandler
	Reply      ReplyEventHandler
	ForumAudit ForumAuditEventHandler

	Interaction InteractionEventHandler

	GroupATMessage  GroupATMessageEventHandler
	C2CMessage      C2CMessageEventHandler
	GroupAddbot     GroupAddRobotEventHandler
	GroupDelbot     GroupDelRobotEventHandler
	GroupMsgReject  GroupMsgRejectHandler
	GroupMsgReceive GroupMsgReceiveHandler
}

// ReadyHandler 可以处理 ws 的 ready 事件
type ReadyHandler func(event *dto.Payload, data *dto.WSReadyData)

// ErrorNotifyHandler 当 ws 连接发生错误的时候，会回调，方便使用方监控相关错误
// 比如 reconnect invalidSession 等错误，错误可以转换为 bot.Err
type ErrorNotifyHandler func(err error)

// PlainEventHandler 透传handler
type PlainEventHandler func(event *dto.Payload, message []byte) error

// HelloHandler 当 ws 接收到 hello 事件的时候会回调
type HelloHandler func(event *dto.Payload)

// ReconnectHandler 当 ws 重新连接的时候会回调
type ReconnectHandler func(event *dto.Payload)

// GuildEventHandler 频道事件handler
type GuildEventHandler func(event *dto.Payload, data *dto.GuildData) error

// GuildMemberEventHandler 频道成员事件 handler
type GuildMemberEventHandler func(event *dto.Payload, data *dto.GuildMemberData) error

// ChannelEventHandler 子频道事件 handler
type ChannelEventHandler func(event *dto.Payload, data *dto.ChannelData) error

// MessageEventHandler 消息事件 handler
type MessageEventHandler func(event *dto.Payload, data *dto.MessageData) error

// MessageDeleteEventHandler 消息事件 handler
type MessageDeleteEventHandler func(event *dto.Payload, data *dto.MessageDeleteData) error

// PublicMessageDeleteEventHandler 消息事件 handler
type PublicMessageDeleteEventHandler func(event *dto.Payload, data *dto.PublicMessageDeleteData) error

// DirectMessageDeleteEventHandler 消息事件 handler
type DirectMessageDeleteEventHandler func(event *dto.Payload, data *dto.DirectMessageDeleteData) error

// MessageReactionEventHandler 表情表态事件 handler
type MessageReactionEventHandler func(event *dto.Payload, data *dto.MessageReactionData) error

// ATMessageEventHandler at 机器人消息事件 handler
type ATMessageEventHandler func(event *dto.Payload, data *dto.ATMessageData) error

// DirectMessageEventHandler 私信消息事件 handler
type DirectMessageEventHandler func(event *dto.Payload, data *dto.DirectMessageData) error

// AudioEventHandler 音频机器人事件 handler
type AudioEventHandler func(event *dto.Payload, data *dto.AudioData) error

// MessageAuditEventHandler 消息审核事件 handler
type MessageAuditEventHandler func(event *dto.Payload, data *dto.MessageAuditData) error

// ThreadEventHandler 论坛主题事件 handler
type ThreadEventHandler func(event *dto.Payload, data *dto.ThreadData) error

// PostEventHandler 论坛回帖事件 handler
type PostEventHandler func(event *dto.Payload, data *dto.PostData) error

// ReplyEventHandler 论坛帖子回复事件 handler
type ReplyEventHandler func(event *dto.Payload, data *dto.ReplyData) error

// ForumAuditEventHandler 论坛帖子审核事件 handler
type ForumAuditEventHandler func(event *dto.Payload, data *dto.ForumAuditData) error

// InteractionEventHandler 互动事件 handler
type InteractionEventHandler func(event *dto.Payload, data *dto.InteractionEventData) error

// ***************** 群消息/C2C消息  *****************

// GroupATMessageEventHandler 群中at机器人消息事件 handler
type GroupATMessageEventHandler func(event *dto.Payload, data *dto.GroupATMessageData) error

// C2CMessageEventHandler 机器人消息事件 handler
type C2CMessageEventHandler func(event *dto.Payload, data *dto.C2CMessageData) error

// GroupAddRobot 机器人新增事件 handler
type GroupAddRobotEventHandler func(event *dto.Payload, data *dto.GroupAddBotEvent) error

// GroupDelRobot 机器人删除事件 handler
type GroupDelRobotEventHandler func(event *dto.Payload, data *dto.GroupAddBotEvent) error

// GroupMsgRejectHandler 机器人推送关闭事件 handler
type GroupMsgRejectHandler func(event *dto.Payload, data *dto.GroupMsgRejectEvent) error

// GroupMsgReceiveHandler 机器人推送开启事件 handler
type GroupMsgReceiveHandler func(event *dto.Payload, data *dto.GroupMsgReceiveEvent) error

// ************************************************

// RegisterHandlers 注册事件回调，并返回 intent 用于 websocket 的鉴权
func RegisterHandlers(handlers ...interface{}) dto.Intent {
	var i dto.Intent
	for _, h := range handlers {
		switch handle := h.(type) {
		case ReadyHandler:
			DefaultHandlers.Ready = handle
		case ErrorNotifyHandler:
			DefaultHandlers.ErrorNotify = handle
		case PlainEventHandler:
			DefaultHandlers.Plain = handle
		case HelloHandler:
			DefaultHandlers.Hello = handle
		case ReconnectHandler:
			DefaultHandlers.Reconnect = handle
		case AudioEventHandler:
			DefaultHandlers.Audio = handle
			i = i | dto.EventToIntent(
				dto.EventAudioStart, dto.EventAudioFinish,
				dto.EventAudioOnMic, dto.EventAudioOffMic,
			)
		case InteractionEventHandler:
			DefaultHandlers.Interaction = handle
			i = i | dto.EventToIntent(dto.EventInteractionCreate)
		case GroupAddRobotEventHandler:
			DefaultHandlers.GroupAddbot = handle
			i = i | dto.EventToIntent(dto.EventGroupAddRobot)
		case GroupDelRobotEventHandler:
			DefaultHandlers.GroupDelbot = handle
			i = i | dto.EventToIntent(dto.EventGroupDelRobot)
		case GroupMsgRejectHandler:
			DefaultHandlers.GroupMsgReject = handle
			i = i | dto.EventToIntent(dto.EventGroupMsgReject)
		case GroupMsgReceiveHandler:
			DefaultHandlers.GroupMsgReceive = handle
			i = i | dto.EventToIntent(dto.EventGroupMsgReceive)
		default:
		}
	}
	i = i | registerRelationHandlers(i, handlers...)
	i = i | registerMessageHandlers(i, handlers...)
	i = i | registerForumHandlers(i, handlers...)

	return i
}

func registerForumHandlers(i dto.Intent, handlers ...interface{}) dto.Intent {
	for _, h := range handlers {
		switch handle := h.(type) {
		case ThreadEventHandler:
			DefaultHandlers.Thread = handle
			i = i | dto.EventToIntent(
				dto.EventForumThreadCreate, dto.EventForumThreadUpdate, dto.EventForumThreadDelete,
			)
		case PostEventHandler:
			DefaultHandlers.Post = handle
			i = i | dto.EventToIntent(dto.EventForumPostCreate, dto.EventForumPostDelete)
		case ReplyEventHandler:
			DefaultHandlers.Reply = handle
			i = i | dto.EventToIntent(dto.EventForumReplyCreate, dto.EventForumReplyDelete)
		case ForumAuditEventHandler:
			DefaultHandlers.ForumAudit = handle
			i = i | dto.EventToIntent(dto.EventForumAuditResult)
		default:
		}
	}
	return i
}

// registerRelationHandlers 注册频道关系链相关handlers
func registerRelationHandlers(i dto.Intent, handlers ...interface{}) dto.Intent {
	for _, h := range handlers {
		switch handle := h.(type) {
		case GuildEventHandler:
			DefaultHandlers.Guild = handle
			i = i | dto.EventToIntent(dto.EventGuildCreate, dto.EventGuildDelete, dto.EventGuildUpdate)
		case GuildMemberEventHandler:
			DefaultHandlers.GuildMember = handle
			i = i | dto.EventToIntent(dto.EventGuildMemberAdd, dto.EventGuildMemberRemove, dto.EventGuildMemberUpdate)
		case ChannelEventHandler:
			DefaultHandlers.Channel = handle
			i = i | dto.EventToIntent(dto.EventChannelCreate, dto.EventChannelDelete, dto.EventChannelUpdate)
		default:
		}
	}
	return i
}

// registerMessageHandlers 注册消息相关的 handler
func registerMessageHandlers(i dto.Intent, handlers ...interface{}) dto.Intent {
	for _, h := range handlers {
		switch handle := h.(type) {
		case MessageEventHandler:
			DefaultHandlers.Message = handle
			i = i | dto.EventToIntent(dto.EventMessageCreate)
		case ATMessageEventHandler:
			DefaultHandlers.ATMessage = handle
			i = i | dto.EventToIntent(dto.EventAtMessageCreate)
		case DirectMessageEventHandler:
			DefaultHandlers.DirectMessage = handle
			i = i | dto.EventToIntent(dto.EventDirectMessageCreate)
		case MessageDeleteEventHandler:
			DefaultHandlers.MessageDelete = handle
			i = i | dto.EventToIntent(dto.EventMessageDelete)
		case PublicMessageDeleteEventHandler:
			DefaultHandlers.PublicMessageDelete = handle
			i = i | dto.EventToIntent(dto.EventPublicMessageDelete)
		case DirectMessageDeleteEventHandler:
			DefaultHandlers.DirectMessageDelete = handle
			i = i | dto.EventToIntent(dto.EventDirectMessageDelete)
		case MessageReactionEventHandler:
			DefaultHandlers.MessageReaction = handle
			i = i | dto.EventToIntent(dto.EventMessageReactionAdd, dto.EventMessageReactionRemove)
		case MessageAuditEventHandler:
			DefaultHandlers.MessageAudit = handle
			i = i | dto.EventToIntent(dto.EventMessageAuditPass, dto.EventMessageAuditReject)
		case GroupATMessageEventHandler:
			DefaultHandlers.GroupATMessage = handle
			i = i | dto.EventToIntent(dto.EventGroupAtMessageCreate)
		case C2CMessageEventHandler:
			DefaultHandlers.C2CMessage = handle
			i = i | dto.EventToIntent(dto.EventC2CMessageCreate)
		default:
		}
	}
	return i
}
