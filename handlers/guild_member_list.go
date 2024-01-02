package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.member", "list", HandleGuildMemberList)
}

// GuildMemberListRequest 获取群组成员列表请求
type GuildMemberListRequest struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// GuildMemberListResponse 获取群组成员列表响应
type GuildMemberListResponse guildmember.GuildMemberList

// HandleGuildMemberList 处理获取群组成员列表请求
func HandleGuildMemberList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildMemberListRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response GuildMemberListResponse

		var dtoMembers []*dto.Member
		dtoMembers, err = apiv2.GuildMembers(context.TODO(), request.GuildId, createGuildMembersPager(request.Next))
		if err != nil {
			return "", err
		}

		response.Next = dtoMembers[len(dtoMembers)-1].User.ID

		for _, dtoMember := range dtoMembers {
			// 将 dto.Member 转换为 guildmember.GuildMember
			guildMember, err := convertDtoMemberToGuildMember(dtoMember)
			if err != nil {
				return "", err
			}

			response.Data = append(response.Data, guildMember)
		}

		var responseBytes []byte
		responseBytes, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseBytes), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// createGuildMembersPager 构建频道成员列表查询参数
func createGuildMembersPager(next string) *dto.GuildMembersPager {
	return &dto.GuildMembersPager{
		After: next,
		Limit: "20",
	}
}
