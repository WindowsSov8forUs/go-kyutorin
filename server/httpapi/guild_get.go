package httpapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
)

func init() {
	RegisterHandler("guild.get", HandleGuildGet)
}

// RequestGuildGet 获取群组请求
type RequestGuildGet struct {
	GuildId string `json:"guild_id"` // 群组 ID
}

// ResponseGuildGet 获取群组响应
type ResponseGuildGet guild.Guild

// HandleGuildGet 处理获取群组请求
func HandleGuildGet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildGet
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}
	if message.Platform == "qqguild" {
		var response ResponseGuildGet
		var dtoGuild *dto.Guild

		dtoGuild, err = apiv2.Guild(context.TODO(), request.GuildId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		response.Id = dtoGuild.ID
		response.Name = dtoGuild.Name
		response.Avatar = dtoGuild.Icon

		return response, nil

	} else if message.Platform == "qq" {
		// 只是通过缓存模拟罢了

		var response ResponseGuildGet
		guildType := processor.GetOpenIdType(request.GuildId)
		if guildType != "group" {
			return gin.H{}, &InternalServerError{fmt.Errorf("no such guild stored in cache: %s", request.GuildId)}
		}
		response.Id = request.GuildId

		return response, nil
	}

	return defaultResource(message)
}
