package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("channel", "create", HandleChannelCreate)
}

// ChannelCreateRequest 创建群组频道请求
type ChannelCreateRequest struct {
	GuildId string           `json:"guild_id"` // 群组 ID
	Data    *channel.Channel `json:"data"`     // 频道数据
}

// ChannelCreateResponse 创建群组频道响应
type ChannelCreateResponse channel.Channel

// HandleChannelCreate 处理创建群组频道请求
func HandleChannelCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ChannelCreateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response ChannelCreateResponse

		// 不能通过这种方式创建私聊子频道
		if request.Data.Type == channel.CHANNEL_TYPE_DIRECT {
			return "", callapi.ErrBadRequest
		}

		var dtoChannel *dto.Channel
		dtoChannel, err = apiv2.PostChannel(context.TODO(), request.GuildId, createChannelValue(request.Data))
		if err != nil {
			return "", err
		}
		response.Id = dtoChannel.ID
		response.Name = dtoChannel.Name
		response.ParentId = dtoChannel.ParentID
		switch dtoChannel.Type {
		case dto.ChannelTypeText:
			response.Type = channel.CHANNEL_TYPE_TEXT
		case dto.ChannelTypeVoice:
			response.Type = channel.CHANNEL_TYPE_VOICE
		case dto.ChannelTypeCategory:
			response.Type = channel.CHANNEL_TYPE_CATEGORY
		default:
			response.Type = channel.CHANNEL_TYPE_CATEGORY
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

// createChannelValue 构建频道值对象
func createChannelValue(channelData *channel.Channel) *dto.ChannelValueObject {
	var dtoChannelValue dto.ChannelValueObject
	dtoChannelValue.Name = channelData.Name
	dtoChannelValue.ParentID = channelData.ParentId
	switch channelData.Type {
	case channel.CHANNEL_TYPE_TEXT:
		dtoChannelValue.Type = dto.ChannelTypeText
	case channel.CHANNEL_TYPE_VOICE:
		dtoChannelValue.Type = dto.ChannelTypeVoice
	case channel.CHANNEL_TYPE_CATEGORY:
		dtoChannelValue.Type = dto.ChannelTypeCategory
	default:
		dtoChannelValue.Type = dto.ChannelTypeText
	}
	dtoChannelValue.SubType = dto.ChannelSubTypeChat // TODO: 默认为闲聊子频道
	return &dtoChannelValue
}
