package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("channel", "delete", HandleChannelDelete)
}

// ChannelDeleteRequest 删除群组频道请求
type ChannelDeleteRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
}

// HandleChannelDelete 处理删除群组频道请求
func HandleChannelDelete(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ChannelDeleteRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		err = apiv2.DeleteChannel(context.TODO(), request.ChannelId)
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
