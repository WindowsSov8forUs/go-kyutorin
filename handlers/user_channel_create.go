package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("user.channel", "create", HandleUserChannelCreate)
}

// UserChannelCreateRequest 创建私聊频道请求
type UserChannelCreateRequest struct {
	UserId  string `json:"user_id"`            // 用户 ID
	GuildId string `json:"guild_id,omitempty"` // 群组 ID
}

// UserChannelCreateResponse 创建私聊频道响应
type UserChannelCreateResponse channel.Channel

// HandleUserChannelCreate 处理创建私聊频道请求
func HandleUserChannelCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request UserChannelCreateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qq" {
		var response UserChannelCreateResponse
		// QQ 单聊没有相应的操作，因此只是在栈中存储一个临时的 ID

		// 查询栈中是否已有该用户的频道
		channelType := echo.GetOpenIdType(request.UserId)
		if channelType == "" {
			// 不存在则创建
			channelType = "private"
			echo.SetOpenIdType(request.UserId, channelType)
		}
		response.Id = request.UserId
		response.Type = channel.CHANNEL_TYPE_DIRECT

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qqguild" {
		var response UserChannelCreateResponse
		// QQ 频道需要调用 API

		var dtoDirectMessage *dto.DirectMessage
		dtoDirectMessage, err = apiv2.CreateDirectMessage(context.TODO(), createDirectMessageToCreate(request))
		if err != nil {
			return "", err
		} else {
			echo.SetDirectChannel(dtoDirectMessage.ChannelID, dtoDirectMessage.GuildID)
		}
		response.Id = dtoDirectMessage.ChannelID
		response.Type = channel.CHANNEL_TYPE_DIRECT

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// createDirectMessageToCreate 构造创建私信频道请求
func createDirectMessageToCreate(request UserChannelCreateRequest) *dto.DirectMessageToCreate {
	return &dto.DirectMessageToCreate{
		SourceGuildID: request.GuildId,
		RecipientID:   request.UserId,
	}
}
