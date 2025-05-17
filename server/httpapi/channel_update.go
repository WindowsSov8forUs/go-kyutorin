package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("channel.update", HandleChannelUpdate)
}

// RequestChannelUpdate 修改群组频道请求
type RequestChannelUpdate struct {
	ChannelId string           `json:"channel_id"` // 频道 ID
	Data      *channel.Channel `json:"data"`       // 频道数据
}

// HandleChannelUpdate 处理修改群组频道请求
func HandleChannelUpdate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestChannelUpdate
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		_, err = apiv2.PatchChannel(context.TODO(), request.ChannelId, createChannelValue(request.Data))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}
