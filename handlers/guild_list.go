package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild", "list", HandleGuildList)
}

// GuildListRequest 获取群组列表请求
type GuildListRequest struct {
	Next string `json:"next,omitempty"` // 分页令牌
}

// GuildListResponse 获取群组列表响应
type GuildListResponse guild.GuildList

// HandleGuildList 处理获取群组列表请求
func HandleGuildList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildListRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}
	if message.Platform == "qqguild" {
		var response GuildListResponse
		var dtoGuilds []*dto.Guild
		dtoGuilds, err = apiv2.MeGuilds(context.TODO(), createGuildPager(request.Next))
		if err != nil {
			return "", err
		}

		response.Next = dtoGuilds[len(dtoGuilds)-1].ID
		response.Data = make([]guild.Guild, len(dtoGuilds))
		for i, dtoGuild := range dtoGuilds {
			response.Data[i].Id = dtoGuild.ID
			response.Data[i].Name = dtoGuild.Name
			response.Data[i].Avatar = dtoGuild.Icon
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// createGuildPager 构建频道列表范围
func createGuildPager(next string) *dto.GuildPager {
	return &dto.GuildPager{
		After: next,
		Limit: "20",
	}
}
