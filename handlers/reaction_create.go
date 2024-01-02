package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("reaction", "create", HandleReactionCreate)
}

// ReactionCreateRequest 添加表态请求
type ReactionCreateRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
	Emoji     string `json:"emoji"`      // 表态名称
}

// HandleReactionCreate 处理添加表态请求
func HandleReactionCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ReactionCreateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		dtoEmoji := dto.Emoji{
			ID:   request.Emoji,
			Type: 1, // 统一为 QQ 系统表情
		}
		err = apiv2.CreateMessageReaction(context.TODO(), request.ChannelId, request.MessageId, dtoEmoji)
		if err != nil {
			return "", err
		}
	}

	return "", callapi.ErrMethodNotAllowed
}
