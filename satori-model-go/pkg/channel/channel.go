package channel

type ChannelType uint64

const (
	CHANNEL_TYPE_TEXT     ChannelType = 0 // 文本频道
	CHANNEL_TYPE_VOICE    ChannelType = 1 // 语音频道
	CHANNEL_TYPE_CATEGORY ChannelType = 2 // 分类频道
	CHANNEL_TYPE_DIRECT   ChannelType = 3 // 私聊频道
)

type Channel struct {
	Id       string      `json:"id"`                  // 频道 ID
	Type     ChannelType `json:"type"`                // 频道类型
	Name     string      `json:"name,omitempty"`      // 频道名称
	ParentId string      `json:"parent_id,omitempty"` // 父频道 ID
}

type ChannelList struct {
	Data []Channel `json:"data"`
	Next string    `json:"next,omitempty"`
}
