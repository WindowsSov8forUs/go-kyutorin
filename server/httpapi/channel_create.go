package httpapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("channel.create", HandleChannelCreate)
}

// RequestChannelCreate 创建群组频道请求
type RequestChannelCreate struct {
	GuildId string           `json:"guild_id"` // 群组 ID
	Data    *channel.Channel `json:"data"`     // 频道数据
}

// ResponseChannelCreate 创建群组频道响应
type ResponseChannelCreate channel.Channel

// HandleChannelCreate 处理创建群组频道请求
func HandleChannelCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestChannelCreate
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseChannelCreate

		// 不能通过这种方式创建私聊子频道
		if request.Data.Type == channel.ChannelTypeDirect {
			return gin.H{}, &BadRequestError{fmt.Errorf("cannot create direct channel using this api")}
		}

		var dtoChannel *dto.Channel
		dtoChannel, err = apiv2.PostChannel(context.TODO(), request.GuildId, createChannelValue(request.Data))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		response.Id = dtoChannel.ID
		response.Name = dtoChannel.Name
		response.ParentId = dtoChannel.ParentID
		switch dtoChannel.Type {
		case dto.ChannelTypeText:
			response.Type = channel.ChannelTypeText
		case dto.ChannelTypeVoice:
			response.Type = channel.ChannelTypeVoice
		case dto.ChannelTypeCategory:
			response.Type = channel.ChannelTypeCategory
		default:
			response.Type = channel.ChannelTypeCategory
		}

		return response, nil
	}

	return defaultResource(message)
}

// createChannelValue 构建频道值对象
func createChannelValue(channelData *channel.Channel) *dto.ChannelValueObject {
	var dtoChannelValue dto.ChannelValueObject
	dtoChannelValue.Name = channelData.Name
	dtoChannelValue.ParentID = channelData.ParentId
	switch channelData.Type {
	case channel.ChannelTypeText:
		dtoChannelValue.Type = dto.ChannelTypeText
	case channel.ChannelTypeVoice:
		dtoChannelValue.Type = dto.ChannelTypeVoice
	case channel.ChannelTypeCategory:
		dtoChannelValue.Type = dto.ChannelTypeCategory
	default:
		dtoChannelValue.Type = dto.ChannelTypeText
	}
	dtoChannelValue.SubType = dto.ChannelSubTypeChat // TODO: 默认为闲聊子频道
	return &dtoChannelValue
}
