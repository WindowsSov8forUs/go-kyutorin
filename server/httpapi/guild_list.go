package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/guild"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
)

func init() {
	RegisterHandler("guild.list", HandleGuildList)
}

// RequestGuildList 获取群组列表请求
type RequestGuildList struct {
	Next string `json:"next,omitempty"` // 分页令牌
}

// ResponseGuildList 获取群组列表响应
type ResponseGuildList guild.GuildList

// HandleGuildList 处理获取群组列表请求
func HandleGuildList(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildList
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseGuildList
		var dtoGuilds []*dto.Guild
		dtoGuilds, err = apiv2.MeGuilds(context.TODO(), createGuildPager(request.Next))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		response.Next = dtoGuilds[len(dtoGuilds)-1].ID
		response.Data = make([]*guild.Guild, len(dtoGuilds))
		for i, dtoGuild := range dtoGuilds {
			response.Data[i].Id = dtoGuild.ID
			response.Data[i].Name = dtoGuild.Name
			response.Data[i].Avatar = dtoGuild.Icon
		}

		return response, nil

	} else if message.Platform == "qq" {
		// 只是通过缓存模拟而已

		var response ResponseGuildList
		guildData := processor.GetOpenIdData()

		// 遍历栈中所有已存储的 openid
		for guildId, guildType := range guildData {
			var g guild.Guild
			g.Id = guildId
			if guildType != "group" {
				continue
			}
			response.Data = append(response.Data, &g)
		}

		return response, nil
	}

	return defaultResource(message)
}

// createGuildPager 构建频道列表范围
func createGuildPager(next string) *dto.GuildPager {
	return &dto.GuildPager{
		After: next,
		Limit: "20",
	}
}
