package channel

type ChannelType uint64

const (
	CHANNEL_TYPE_TEXT     ChannelType = 0
	CHANNEL_TYPE_VOICE    ChannelType = 1
	CHANNEL_TYPE_CATEGORY ChannelType = 2
	CHANNEL_TYPE_DIRECT   ChannelType = 3
)

type Channel struct {
	Id       string      `json:"id"`
	Type     ChannelType `json:"type"`
	Name     string      `json:"name"`
	ParentId string      `json:"parent_id"`
}

type ChannelList struct {
	Data []Channel `json:"data"`
	Next string    `json:"next,omitempty"`
}
