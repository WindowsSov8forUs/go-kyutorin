package httpapi

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("message.delete", HandleMessageDelete)
}

// MessageDeleteRequest 撤回消息请求
type MessageDeleteRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
}

// HandleMessageDelete 处理撤回消息请求
func HandleMessageDelete(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request MessageDeleteRequest
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		// 尝试获取私聊频道，若没有则视为群组频道
		guildId := processor.GetDirectChannelGuild(request.ChannelId)
		if guildId == "" {
			// 群组频道
			err = apiv2.RetractMessage(context.TODO(), request.ChannelId, request.MessageId)
			if err != nil {
				return gin.H{}, &InternalServerError{err}
			}
			return gin.H{}, nil
		} else {
			// 私聊频道
			err = apiv2.RetractDMMessage(context.TODO(), guildId, request.MessageId)
			if err != nil {
				return gin.H{}, &InternalServerError{err}
			}
			return gin.H{}, nil
		}
	}

	return defaultResource(message)
}
