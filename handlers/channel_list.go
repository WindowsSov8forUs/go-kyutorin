package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("channel", "list", HandleChannelList)
}

// ChannelListRequest 获取群组频道列表请求
type ChannelListRequest struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// ChannelListResponse 获取群组频道列表响应
type ChannelListResponse channel.ChannelList

// HandleChannelList 处理获取群组频道列表请求
func HandleChannelList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ChannelListRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response ChannelListResponse

		var dtoChannels []*dto.Channel
		dtoChannels, err = apiv2.Channels(context.TODO(), request.GuildId)
		if err != nil {
			return "", err
		}
		for _, dtoChannel := range dtoChannels {
			var channelType channel.ChannelType
			switch dtoChannel.Type {
			case dto.ChannelTypeText:
				channelType = channel.CHANNEL_TYPE_TEXT
			case dto.ChannelTypeVoice:
				channelType = channel.CHANNEL_TYPE_VOICE
			case dto.ChannelTypeCategory:
				channelType = channel.CHANNEL_TYPE_CATEGORY
			default:
				channelType = channel.CHANNEL_TYPE_CATEGORY
			}
			response.Data = append(response.Data, channel.Channel{
				Id:       dtoChannel.ID,
				Name:     dtoChannel.Name,
				ParentId: dtoChannel.ParentID,
				Type:     channelType,
			})
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}
