package httpapi

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("user.channel.create", HandleUserChannelCreate)
}

// RequestUserChannelCreate 创建私聊频道请求
type RequestUserChannelCreate struct {
	UserId  string `json:"user_id"`            // 用户 ID
	GuildId string `json:"guild_id,omitempty"` // 群组 ID
}

// ResponseUserChannelCreate 创建私聊频道响应
type ResponseUserChannelCreate channel.Channel

// HandleUserChannelCreate 处理创建私聊频道请求
func HandleUserChannelCreate(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestUserChannelCreate
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qq" {
		var response ResponseUserChannelCreate
		// QQ 单聊没有相应的操作，因此只是在栈中存储一个临时的 ID

		// 查询栈中是否已有该用户的频道
		channelType := processor.GetOpenIdType(request.UserId)
		if channelType == "" {
			// 不存在则创建
			channelType = "private"
			processor.SetOpenIdType(request.UserId, channelType)
		}
		response.Id = request.UserId
		response.Type = channel.ChannelTypeDirect

		return response, nil
	} else if message.Platform == "qqguild" {
		var response ResponseUserChannelCreate
		// QQ 频道需要调用 API

		var dtoDirectMessage *dto.DirectMessage
		dtoDirectMessage, err = apiv2.CreateDirectMessage(context.TODO(), createDirectMessageToCreate(request))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		} else {
			processor.SetDirectChannel(dtoDirectMessage.ChannelID, dtoDirectMessage.GuildID)
		}
		response.Id = dtoDirectMessage.ChannelID
		response.Type = channel.ChannelTypeDirect

		return response, nil
	}

	return defaultResource(message)
}

// createDirectMessageToCreate 构造创建私信频道请求
func createDirectMessageToCreate(request RequestUserChannelCreate) *dto.DirectMessageToCreate {
	return &dto.DirectMessageToCreate{
		SourceGuildID: request.GuildId,
		RecipientID:   request.UserId,
	}
}
