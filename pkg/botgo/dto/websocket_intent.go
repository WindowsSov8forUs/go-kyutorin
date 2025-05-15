package dto

// Intent 类型
type Intent int

// websocket intent 声明
const (
	// IntentGuilds 包含
	// - GUILD_CREATE
	// - GUILD_UPDATE
	// - GUILD_DELETE
	// - GUILD_ROLE_CREATE
	// - GUILD_ROLE_UPDATE
	// - GUILD_ROLE_DELETE
	// - CHANNEL_CREATE
	// - CHANNEL_UPDATE
	// - CHANNEL_DELETE
	// - CHANNEL_PINS_UPDATE
	IntentGuilds Intent = 1 << iota

	// IntentGuildMembers 包含
	// - GUILD_MEMBER_ADD
	// - GUILD_MEMBER_UPDATE
	// - GUILD_MEMBER_REMOVE
	IntentGuildMembers

	IntentGuildBans
	IntentGuildEmojis
	IntentGuildIntegrations
	IntentGuildWebhooks
	IntentGuildInvites
	IntentGuildVoiceStates
	IntentGuildPresences
	IntentGuildMessages

	// IntentGuildMessageReactions 包含
	// - MESSAGE_REACTION_ADD
	// - MESSAGE_REACTION_REMOVE
	IntentGuildMessageReactions

	IntentGuildMessageTyping
	IntentDirectMessages
	IntentDirectMessageReactions
	IntentDirectMessageTyping

	// IntentGroupAndC2CEvent 群消息事件
	// - GROUP_AT_MESSAGE_CREATE // 群中@机器人时的消息
	IntentGroupAndC2CEvent Intent = 1 << 25 // 群消息事件

	IntentInteraction  Intent = 1 << 26 // 互动事件
	IntentMessageAudit Intent = 1 << 27 // 审核事件
	// IntentForumEvent 论坛事件
	//  - THREAD_CREATE     // 当用户创建主题时
	//  - THREAD_UPDATE     // 当用户更新主题时
	//  - THREAD_DELETE     // 当用户删除主题时
	//  - POST_CREATE       // 当用户创建帖子时
	//  - POST_DELETE       // 当用户删除帖子时
	//  - REPLY_CREATE      // 当用户回复评论时
	//  - REPLY_DELETE      // 当用户回复评论时
	//  - FORUM_PUBLISH_AUDIT_RESULT      // 当用户发表审核通过时
	IntentForumEvent Intent = 1 << 28 // 论坛事件

	// IntentAudioAction
	//  - AUDIO_START           // 音频开始播放时
	//  - AUDIO_FINISH          // 音频播放结束时
	IntentAudioAction         Intent = 1 << 29 // 音频机器人事件
	IntentPublicGuildMessages Intent = 1 << 30 // 只接收@消息事件

	IntentNone Intent = 0
)
