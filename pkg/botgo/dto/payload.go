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

// GuildData 频道 payload
type GuildData Guild

// GuildMemberData 频道成员 payload
type GuildMemberData Member

// ChannelData 子频道 payload
type ChannelData Channel

// MessageData 消息 payload
type MessageData Message

// ATMessageData only at 机器人的消息 payload
type ATMessageData Message

// DirectMessageData 私信消息 payload
type DirectMessageData Message

// MessageDeleteData 消息 payload
type MessageDeleteData MessageDelete

// PublicMessageDeleteData 公域机器人的消息删除 payload
type PublicMessageDeleteData MessageDelete

// DirectMessageDeleteData 私信消息 payload
type DirectMessageDeleteData MessageDelete

// AudioData 音频机器人的音频流事件
type AudioData AudioAction

// MessageReactionData 表情表态事件
type MessageReactionData MessageReaction

// MessageAuditData 消息审核事件
type MessageAuditData MessageAudit

// ThreadData 主题事件
type ThreadData Thread

// PostData 帖子事件
type PostData Post

// ReplyData 帖子回复事件
type ReplyData Reply

// ForumAuditData 帖子审核事件
type ForumAuditData ForumAuditResult

// InteractionEventData 互动事件
type InteractionEventData Interaction

// ***************** 群消息/C2C消息  *****************

// GroupATMessageData 群@机器人的事件
type GroupATMessageData Message

// C2CMessageData  c2c消息事件
type C2CMessageData Message

// ************************************************
