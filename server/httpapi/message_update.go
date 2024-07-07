package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("message.update", HandleMessageUpdate)
}

// RequestMessageUpdate 编辑消息请求
type RequestMessageUpdate struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	MessageId string `json:"message_id"` // 消息 ID
	Content   string `json:"content"`    // 消息内容
}

// HandleMessageUpdate 处理编辑消息请求
func HandleMessageUpdate(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestMessageUpdate
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var dtoMessageToCreate = &dto.MessageToCreate{}
		dtoMessageToCreate, err = convertToMessageToCreate(request.Content)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		_, err := apiv2.PatchMessage(context.TODO(), request.ChannelId, request.MessageId, dtoMessageToCreate)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		return gin.H{}, nil
	}

	return defaultResource(message)
}
