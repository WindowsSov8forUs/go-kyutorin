package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("channel", "get", HandleChannelGet)
}

// ChannelGetRequest 获取群组频道请求
type ChannelGetRequest struct {
	ChannelID string `json:"channel_id"` // 频道 ID
}

// ChannelGetResponse 获取群组频道响应
type ChannelGetResponse channel.Channel

// HandleChannelGet 处理获取群组频道请求
func HandleChannelGet(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ChannelGetRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response ChannelGetResponse

		// 获取栈中的私聊频道信息（如果存在）
		channelType := echo.GetDirectChannelGuild(request.ChannelID)
		if channelType != "" {
			response.Id = request.ChannelID
			response.Type = channel.CHANNEL_TYPE_DIRECT
		} else {
			var dtoChannel *dto.Channel
			dtoChannel, err = apiv2.Channel(context.TODO(), request.ChannelID)
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
		}
		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qq" {
		// 只是通过缓存模拟而已

		var response ChannelGetResponse
		channelType := echo.GetOpenIdType(request.ChannelID)
		response.Id = request.ChannelID
		switch channelType {
		case "private":
			response.Type = channel.CHANNEL_TYPE_DIRECT
		case "group":
			response.Type = channel.CHANNEL_TYPE_TEXT
		default:
			return "", fmt.Errorf("频道未存储于缓存中: %s", request.ChannelID)
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
