package dto

import (
	"fmt"

	"github.com/tencent-connect/botgo/token"
)

// WebsocketAP wss 接入点信息
type WebsocketAP struct {
	URL               string            `json:"url"`
	Shards            uint32            `json:"shards"`
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// WebsocketAP wss 单个接入点信息
type WebsocketAPSingle struct {
	URL               string            `json:"url"`
	ShardCount        uint32            `json:"shards"`   //最大值 比如是4个分片就是4
	ShardID           uint32            `json:"shard_id"` //从0开始的 0 1 2 3 对应上面的
	SessionStartLimit SessionStartLimit `json:"session_start_limit"`
}

// SessionStartLimit 链接频控信息
type SessionStartLimit struct {
	Total          uint32 `json:"total"`
	Remaining      uint32 `json:"remaining"`
	ResetAfter     uint32 `json:"reset_after"`
	MaxConcurrency uint32 `json:"max_concurrency"`
}

// ShardConfig 连接的 shard 配置，ShardID 从 0 开始，ShardCount 最小为 1
type ShardConfig struct {
	ShardID    uint32
	ShardCount uint32
}

// Session 连接的 session 结构，包括链接的所有必要字段
type Session struct {
	ID      string
	URL     string
	Token   token.Token
	Intent  Intent
	LastSeq uint32
	Shards  ShardConfig
}

// String 输出session字符串
func (s *Session) String() string {
	return fmt.Sprintf("[ws][ID:%s][Shard:(%d/%d)][Intent:%d]",
		s.ID, s.Shards.ShardID, s.Shards.ShardCount, s.Intent)
}

// WSUser 当前连接的用户信息
type WSUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

// WSIdentityData 鉴权数据
type WSIdentityData struct {
	Token      string   `json:"token"`
	Intents    Intent   `json:"intents"`
	Shard      []uint32 `json:"shard"` // array of two integers (shard_id, num_shards)
	Properties struct {
		Os      string `json:"$os,omitempty"`
		Browser string `json:"$browser,omitempty"`
		Device  string `json:"$device,omitempty"`
	} `json:"properties,omitempty"`
}

// WSResumeData 重连数据
type WSResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       uint32 `json:"seq"`
}

// 以下为会收到的事件data

// WSHelloData hello 返回
type WSHelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// WSReadyData ready，鉴权后返回
type WSReadyData struct {
	Version   int    `json:"version"`
	SessionID string `json:"session_id"`
	User      struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Bot      bool   `json:"bot"`
	} `json:"user"`
	Shard []uint32 `json:"shard"`
}

// WSGuildData 频道 payload
type WSGuildData Guild

// WSGuildMemberData 频道成员 payload
type WSGuildMemberData Member

// WSChannelData 子频道 payload
type WSChannelData Channel

// WSMessageData 消息 payload
type WSMessageData Message

// WSATMessageData only at 机器人的消息 payload
type WSATMessageData Message

// WSDirectMessageData 私信消息 payload
type WSDirectMessageData Message

// WSMessageDeleteData 消息 payload
type WSMessageDeleteData MessageDelete

// WSPublicMessageDeleteData 公域机器人的消息删除 payload
type WSPublicMessageDeleteData MessageDelete

// WSDirectMessageDeleteData 私信消息 payload
type WSDirectMessageDeleteData MessageDelete

// WSAudioData 音频机器人的音频流事件
type WSAudioData AudioAction

// WSMessageReactionData 表情表态事件
type WSMessageReactionData MessageReaction

// WSMessageAuditData 消息审核事件
type WSMessageAuditData MessageAudit

// WSThreadData 主题事件
type WSThreadData Thread

// WSPostData 帖子事件
type WSPostData Post

// WSReplyData 帖子回复事件
type WSReplyData Reply

// WSForumAuditData 帖子审核事件
type WSForumAuditData ForumAuditResult

// WSInteractionData 互动事件
type WSInteractionData Interaction

// ***************** 群消息/C2C消息  *****************

// WSGroupATMessageData 群@机器人的事件
type WSGroupATMessageData Message

// WSC2CMessageData  c2c消息事件
type WSC2CMessageData Message

// ************************************************
