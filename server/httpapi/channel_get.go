package httpapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WindowsSov8forUs/glyccat/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("channel.get", HandleChannelGet)
}

// RequestChannelGet 获取群组频道请求
type RequestChannelGet struct {
	ChannelID string `json:"channel_id"` // 频道 ID
}

// ResponseChannelGet 获取群组频道响应
type ResponseChannelGet channel.Channel

// HandleChannelGet 处理获取群组频道请求
func HandleChannelGet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestChannelGet
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseChannelGet

		// 获取栈中的私聊频道信息（如果存在）
		channelType := processor.GetDirectChannelGuild(request.ChannelID)
		if channelType != "" {
			response.Id = request.ChannelID
			response.Type = channel.ChannelTypeDirect

		} else {
			var dtoChannel *dto.Channel
			dtoChannel, err = apiv2.Channel(context.TODO(), request.ChannelID)
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
		}

		return response, nil

	} else if message.Platform == "qq" {
		// 只是通过缓存模拟而已
		var response ResponseChannelGet

		channelType := processor.GetOpenIdType(request.ChannelID)
		response.Id = request.ChannelID
		switch channelType {
		case "private":
			response.Type = channel.ChannelTypeDirect
		case "group":
			response.Type = channel.ChannelTypeText
		default:
			return gin.H{}, &InternalServerError{fmt.Errorf("no such channel stored in cache: %s", request.ChannelID)}
		}

		return response, nil
	}

	return defaultResource(message)
}
