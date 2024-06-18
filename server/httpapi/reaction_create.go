package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("reaction.create", HandleReactionCreate)
}

// RequestReactionCreate 添加表态请求
type RequestReactionCreate struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
	Emoji     string `json:"emoji"`      // 表态名称
}

// HandleReactionCreate 处理添加表态请求
func HandleReactionCreate(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestReactionCreate
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		dtoEmoji := dto.Emoji{
			ID:   request.Emoji,
			Type: 1, // 统一为 QQ 系统表情
		}
		err = apiv2.CreateMessageReaction(context.TODO(), request.ChannelId, request.MessageId, dtoEmoji)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
	}

	return defaultResource(message)
}
