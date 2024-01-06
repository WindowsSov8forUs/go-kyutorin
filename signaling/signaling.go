package signaling

import (
	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/dezhishen/satori-model-go/pkg/guildrole"
	"github.com/dezhishen/satori-model-go/pkg/interaction"
	"github.com/dezhishen/satori-model-go/pkg/login"
	"github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
)

// Signaling WebSocket 发送的信令的数据结构
type Signaling struct {
	Op   SignalingType `json:"op"`             // 信令类型
	Body interface{}   `json:"body,omitempty"` // 信令数据
}

// SignalingType 信令类型
type SignalingType int

const (
	SIGNALING_EVENT    SignalingType = iota // 事件
	SIGNALING_PING                          // 心跳
	SIGNALING_PONG                          // 心跳回复
	SIGNALING_IDENTIFY                      // 鉴权
	SIGNALING_READY                         // 鉴权回复
)

// READY 信令的信令数据
type ReadyBody struct {
	Logins []*login.Login `json:"logins"` // 登录信息
}

// 事件类型定义
type Event struct {
	Id        int64                    `json:"id"`                 // 事件 ID
	Type      EventType                `json:"type"`               // 事件类型
	Platform  string                   `json:"platform"`           // 接收者的平台名称
	SelfId    string                   `json:"self_id"`            // 接收者的平台账号
	Timestamp int64                    `json:"timestamp"`          // 事件的时间戳
	Argv      *interaction.Argv        `json:"argv,omitempty"`     // 交互指令
	Button    *interaction.Button      `json:"button,omitempty"`   // 交互按钮
	Channel   *channel.Channel         `json:"channel,omitempty"`  // 事件所属的频道
	Guild     *guild.Guild             `json:"guild,omitempty"`    // 事件所属的群组
	Login     *login.Login             `json:"login,omitempty"`    // 事件的登录信息
	Member    *guildmember.GuildMember `json:"member,omitempty"`   // 事件的目标成员
	Message   *message.Message         `json:"message,omitempty"`  // 事件的消息
	Operator  *user.User               `json:"operator,omitempty"` // 事件的操作者
	Role      *guildrole.GuildRole     `json:"role,omitempty"`     // 事件的目标角色
	User      *user.User               `json:"user,omitempty"`     // 事件的目标用户
	Type_     string                   `json:"_type,omitempty"`    // 原生事件类型
	Data_     interface{}              `json:"_data,omitempty"`    // 原生事件数据
}

// EventType 事件类型
type EventType string

const (
	// Guild 事件

	EVENT_TYPE_GUILD_ADDED   EventType = "guild-added"   // 加入群组时触发
	EVENT_TYPE_GUILD_UPDATED EventType = "guild-updated" // 群组被修改时触发
	EVENT_TYPE_GUILD_REMOVED EventType = "guild-removed" // 退出群组时触发
	EVENT_TYPE_GUILD_REQUEST EventType = "guild-request" // 接收到新的入群邀请时触发

	// GuildMember 事件

	EVENT_TYPE_GUILD_MEMBER_ADDED   EventType = "guild-member-added"   // 群组成员增加时触发
	EVENT_TYPE_GUILD_MEMBER_UPDATED EventType = "guild-member-updated" // 群组成员信息更新时触发
	EVENT_TYPE_GUILD_MEMBER_REMOVED EventType = "guild-member-removed" // 群组成员移除时触发
	EVENT_TYPE_GUILD_MEMBER_REQUEST EventType = "guild-member-request" // 接收到新的加群请求时触发

	// GuildRole 事件

	EVENT_TYPE_GUILD_ROLE_ADDED   EventType = "guild-role-added"   // 群组角色被创建时触发
	EVENT_TYPE_GUILD_ROLE_UPDATED EventType = "guild-role-updated" // 群组角色被修改时触发
	EVENT_TYPE_GUILD_ROLE_REMOVED EventType = "guild-role-removed" // 群组角色被删除时触发

	// Interaction 事件

	EVENT_TYPE_INTERACTION_BUTTON  EventType = "interaction/button"  // 类型为 action 的按钮被点击时触发
	EVENT_TYPE_INTERACTION_COMMAND EventType = "interaction/command" // 调用斜线指令时触发

	// Login 事件

	EVENT_TYPE_LOGIN_ADDED   EventType = "login-added"   // 登录被创建时触发
	EVENT_TYPE_LOGIN_REMOVED EventType = "login-removed" // 登录被删除时触发
	EVENT_TYPE_LOGIN_UPDATED EventType = "login-updated" // 登录信息更新时触发

	// Message 事件

	EVENT_TYPE_MESSAGE_CREATED EventType = "message-created" // 当消息被创建时触发
	EVENT_TYPE_MESSAGE_UPDATED EventType = "message-updated" // 当消息被编辑时触发
	EVENT_TYPE_MESSAGE_DELETED EventType = "message-deleted" // 当消息被删除时触发

	// Reaction 事件

	EVENT_TYPE_REACTION_ADDED   EventType = "reaction-added"   // 当表态被添加时触发
	EVENT_TYPE_REACTION_REMOVED EventType = "reaction-removed" // 当表态被移除时触发

	// User 事件

	EVENT_TYPE_FRIEND_REQUEST EventType = "friend-request" // 接收到新的好友申请时触发

	// Internal 事件

	EVENT_TYPE_INTERNAL EventType = "internal" // 内部事件
)

// EVENT 信令的信令数据
type EventBody Event

// IDENTIFY 信令的信令数据
type IdentifyBody struct {
	Token    string `json:"token,omitempty"`    // 鉴权令牌
	Sequence int64  `json:"sequence,omitempty"` // 序列号
}
