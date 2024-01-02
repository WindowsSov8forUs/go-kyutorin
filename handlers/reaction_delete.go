package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("reaction", "delete", HandleReactionDelete)
}

// ReactionDeleteRequest 删除表态请求
type ReactionDeleteRequest struct {
	ChannelId string `json:"channel_id"`        // 频道 ID
	MessageId string `json:"message_id"`        // 消息 ID
	Emoji     string `json:"emoji"`             // 表态名称
	UserId    string `json:"user_id,omitempty"` // 用户 ID
}

// HandleReactionDelete 处理删除表态请求
func HandleReactionDelete(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ReactionDeleteRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		// 不允许指定用户 ID
		if request.UserId != "" {
			return "", callapi.ErrBadRequest
		}

		dtoEmoji := dto.Emoji{
			ID:   request.Emoji,
			Type: 1, // 统一为 QQ 系统表情
		}
		err = apiv2.DeleteOwnMessageReaction(context.TODO(), request.ChannelId, request.MessageId, dtoEmoji)
		if err != nil {
			return "", err
		}
	}

	return "", callapi.ErrMethodNotAllowed
}
