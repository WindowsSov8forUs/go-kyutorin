package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("channel.delete", HandleChannelDelete)
}

// RequestChannelDelete 删除群组频道请求
type RequestChannelDelete struct {
	ChannelId string `json:"channel_id"` // 频道 ID
}

// HandleChannelDelete 处理删除群组频道请求
func HandleChannelDelete(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestChannelDelete
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		err = apiv2.DeleteChannel(context.TODO(), request.ChannelId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}
