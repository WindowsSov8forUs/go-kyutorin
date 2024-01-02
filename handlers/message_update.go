package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("message", "update", HandleMessageUpdate)
}

// MessageUpdateRequest 编辑消息请求
type MessageUpdateRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
	Content   string `json:"content"`    // 消息内容
}

// HandleMessageUpdate 处理编辑消息请求
func HandleMessageUpdate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request MessageUpdateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var dtoMessageToCreate = &dto.MessageToCreate{}
		dtoMessageToCreate, err = convertToMessageToCreate(request.Content)
		if err != nil {
			return "", err
		}
		_, err := apiv2.PatchMessage(context.TODO(), request.ChannelId, request.MessageId, dtoMessageToCreate)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
