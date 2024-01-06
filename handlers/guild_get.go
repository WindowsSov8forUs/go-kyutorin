package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"

	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild", "get", HandleGuildGet)
}

// GuildGetRequest 获取群组请求
type GuildGetRequest struct {
	GuildId string `json:"guild_id"` // 群组 ID
}

// GuildGetResponse 获取群组响应
type GuildGetResponse guild.Guild

// HandleGuildGet 处理获取群组请求
func HandleGuildGet(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildGetRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}
	if message.Platform == "qqguild" {
		var response GuildGetResponse
		var dtoGuild *dto.Guild
		dtoGuild, err = apiv2.Guild(context.TODO(), request.GuildId)
		if err != nil {
			return "", err
		}
		response.Id = dtoGuild.ID
		response.Name = dtoGuild.Name
		response.Avatar = dtoGuild.Icon

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qq" {
		// 只是通过缓存模拟罢了

		var response GuildGetResponse
		guildType := echo.GetOpenIdType(request.GuildId)
		if guildType != "group" {
			return "", fmt.Errorf("群组未被记录在缓存中: %s", request.GuildId)
		}
		response.Id = request.GuildId

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}
