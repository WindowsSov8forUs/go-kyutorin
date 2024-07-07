package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.member.role.set", HandleGuildMemberRoleSet)
}

// RequestGuildMemberRoleSet 设置群组成员角色请求
type RequestGuildMemberRoleSet struct {
	GuildId string `json:"guild_id"` // 群组 ID
	UserId  string `json:"user_id"`  // 用户 ID
	RoleId  string `json:"role_id"`  // 角色 ID
}

// HandleGuildMemberRoleSet 处理设置群组成员角色请求
func HandleGuildMemberRoleSet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildMemberRoleSet
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		dtoMemberAddRoleBody := &dto.MemberAddRoleBody{
			Channel: &dto.Channel{},
		}
		err = apiv2.MemberAddRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId), request.UserId, dtoMemberAddRoleBody)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}
		return gin.H{}, nil
	}

	return defaultResource(message)
}
