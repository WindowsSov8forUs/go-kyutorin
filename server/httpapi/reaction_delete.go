package httpapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("reaction.delete", HandleReactionDelete)
}

// RequestReactionDelete 删除表态请求
type RequestReactionDelete struct {
	ChannelId string `json:"channel_id"`        // 频道 ID
	MessageId string `json:"message_id"`        // 消息 ID
	Emoji     string `json:"emoji"`             // 表态名称
	UserId    string `json:"user_id,omitempty"` // 用户 ID
}

// HandleReactionDelete 处理删除表态请求
func HandleReactionDelete(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestReactionDelete
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		// 不允许指定用户 ID
		if request.UserId != "" {
			return gin.H{}, &BadRequestError{fmt.Errorf(`"user_id" is not allowed in this request on platform "qqguild".`)}
		}

		dtoEmoji := dto.Emoji{
			ID:   request.Emoji,
			Type: 1, // 统一为 QQ 系统表情
		}
		err = apiv2.DeleteOwnMessageReaction(context.TODO(), request.ChannelId, request.MessageId, dtoEmoji)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
	}

	return defaultResource(message)
}
