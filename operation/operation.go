package operation

import (
	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildrole"
	"github.com/satori-protocol-go/satori-model-go/pkg/interaction"
	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
)

// Operation WebSocket 发送的信令的数据结构
type Operation struct {
	Op   OpCode      `json:"op"`             // 信令类型
	Body interface{} `json:"body,omitempty"` // 信令数据
}

// OpCode 信令类型
type OpCode int

const (
	OpCodeEvent    OpCode = iota // 事件
	OpCodePing                   // 心跳
	OpCodePong                   // 心跳回复
	OpCodeIdentify               // 鉴权
	OpCodeReady                  // 鉴权成功
	OpCodeMeta                   // 元信息更新
)

// 事件类型定义
type Event struct {
	Sn        int64                    `json:"sn"`                 // 序列号
	Type      EventType                `json:"type"`               // 事件类型
	Timestamp int64                    `json:"timestamp"`          // 事件的时间戳
	Login     *login.Login             `json:"login"`              // 登录信息
	Argv      *interaction.Argv        `json:"argv,omitempty"`     // 交互指令
	Button    *interaction.Button      `json:"button,omitempty"`   // 交互按钮
	Channel   *channel.Channel         `json:"channel,omitempty"`  // 事件所属的频道
	Guild     *guild.Guild             `json:"guild,omitempty"`    // 事件所属的群组
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

	EventTypeGuildAdded   EventType = "guild-added"   // 加入群组时触发
	EventTypeGuildUpdated EventType = "guild-updated" // 群组被修改时触发
	EventTypeGuildRemoved EventType = "guild-removed" // 退出群组时触发
	EventTypeGuildRequest EventType = "guild-request" // 接收到新的入群邀请时触发

	// GuildMember 事件

	EventTypeGuildMemberAdded   EventType = "guild-member-added"   // 群组成员增加时触发
	EventTypeGuildMemberUpdated EventType = "guild-member-updated" // 群组成员信息更新时触发
	EventTypeGuildMemberRemoved EventType = "guild-member-removed" // 群组成员移除时触发
	EventTypeGuildMemberRequest EventType = "guild-member-request" // 接收到新的加群请求时触发

	// GuildRole 事件

	EventTypeGuildRoleCreated EventType = "guild-role-created" // 群组角色被创建时触发
	EventTypeGuildRoleUpdated EventType = "guild-role-updated" // 群组角色被修改时触发
	EventTypeGuildRoleDeleted EventType = "guild-role-deleted" // 群组角色被删除时触发

	// Interaction 事件

	EventTypeInteractionButton  EventType = "interaction/button"  // 类型为 action 的按钮被点击时触发
	EventTypeInteractionCommand EventType = "interaction/command" // 调用斜线指令时触发

	// Login 事件

	EventTypeLoginAdded   EventType = "login-added"   // 登录被创建时触发
	EventTypeLoginRemoved EventType = "login-removed" // 登录被删除时触发
	EventTypeLoginUpdated EventType = "login-updated" // 登录信息更新时触发

	// Message 事件

	EventTypeMessageCreated EventType = "message-created" // 当消息被创建时触发
	EventTypeMessageUpdated EventType = "message-updated" // 当消息被编辑时触发
	EventTypeMessageDeleted EventType = "message-deleted" // 当消息被删除时触发

	// Reaction 事件

	EventTypeReactionAdded   EventType = "reaction-added"   // 当表态被添加时触发
	EventTypeReactionRemoved EventType = "reaction-removed" // 当表态被移除时触发

	// User 事件

	EventTypeFriendRequest EventType = "friend-request" // 接收到新的好友申请时触发

	// Internal 事件

	EventTypeInternal EventType = "internal" // 内部事件
)

// EVENT 信令的信令数据
type EventBody Event

// IDENTIFY 信令的信令数据
type IdentifyBody struct {
	Token string `json:"token,omitempty"` // 鉴权令牌
	Sn    int64  `json:"sn,omitempty"`    // 序列号
}

// READY 信令的信令数据
type ReadyBody struct {
	Logins    []*login.Login `json:"logins"`     // 登录信息
	ProxyUrls []string       `json:"proxy_urls"` // 代理路由 列表
}

// META 信令的信令数据
type MetaBody struct {
	ProxyUrls []string `json:"proxy_urls"` // 代理路由 列表
}
