package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
)

func init() {
	RegisterHandler("channel.list", HandleChannelList)
}

// RequestChannelList 获取群组频道列表请求
type RequestChannelList struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// ResponseChannelList 获取群组频道列表响应
type ResponseChannelList channel.ChannelList

// HandleChannelList 处理获取群组频道列表请求
func HandleChannelList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestChannelList
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseChannelList

		var dtoChannels []*dto.Channel
		dtoChannels, err = apiv2.Channels(context.TODO(), request.GuildId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		for _, dtoChannel := range dtoChannels {
			var channelType channel.ChannelType
			switch dtoChannel.Type {
			case dto.ChannelTypeText:
				channelType = channel.ChannelTypeText
			case dto.ChannelTypeVoice:
				channelType = channel.ChannelTypeVoice
			case dto.ChannelTypeCategory:
				channelType = channel.ChannelTypeCategory
			default:
				channelType = channel.ChannelTypeCategory
			}
			response.Data = append(response.Data, &channel.Channel{
				Id:       dtoChannel.ID,
				Name:     dtoChannel.Name,
				ParentId: dtoChannel.ParentID,
				Type:     channelType,
			})
		}

		return response, nil
	} else if message.Platform == "qq" {
		// 只是通过缓存模拟而已

		var response ResponseChannelList
		guildType := processor.GetOpenIdType(request.GuildId)

		// 因为一群一频道，所以不存在多个频道的可能性
		if guildType == "group" {
			var c channel.Channel
			c.Id = request.GuildId
			c.Type = channel.ChannelTypeText
			response.Data = append(response.Data, &c)
		}

		return response, nil
	}

	return defaultResource(message)
}
