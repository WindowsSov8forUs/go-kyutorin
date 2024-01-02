package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("channel", "update", HandleChannelUpdate)
}

// ChannelUpdateRequest 修改群组频道请求
type ChannelUpdateRequest struct {
	ChannelId string           `json:"channel_id"` // 频道 ID
	Data      *channel.Channel `json:"data"`       // 频道数据
}

// HandleChannelUpdate 处理修改群组频道请求
func HandleChannelUpdate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request ChannelUpdateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		_, err = apiv2.PatchChannel(context.TODO(), request.ChannelId, createChannelValue(request.Data))
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
