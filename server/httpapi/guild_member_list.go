package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildmember"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.member.list", HandleGuildMemberList)
}

// RequestGuildMemberList 获取群组成员列表请求
type RequestGuildMemberList struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// ResponseGuildMemberList 获取群组成员列表响应
type ResponseGuildMemberList guildmember.GuildMemberList

// HandleGuildMemberList 处理获取群组成员列表请求
func HandleGuildMemberList(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildMemberList
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseGuildMemberList

		var dtoMembers []*dto.Member
		dtoMembers, err = apiv2.GuildMembers(context.TODO(), request.GuildId, createGuildMembersPager(request.Next))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		response.Next = dtoMembers[len(dtoMembers)-1].User.ID

		for _, dtoMember := range dtoMembers {
			// 将 dto.Member 转换为 guildmember.GuildMember
			guildMember, err := convertDtoMemberToGuildMember(dtoMember)
			if err != nil {
				return gin.H{}, &InternalServerError{err}
			}

			response.Data = append(response.Data, &guildMember)
		}

		return response, nil
	}

	return defaultResource(message)
}

// createGuildMembersPager 构建频道成员列表查询参数
func createGuildMembersPager(next string) *dto.GuildMembersPager {
	return &dto.GuildMembersPager{
		After: next,
		Limit: "20",
	}
}
