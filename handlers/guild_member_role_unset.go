package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.member.role", "unset", HandleGuildMemberRoleUnset)
}

// GuildMemberRoleUnsetRequest 取消群组成员角色请求
type GuildMemberRoleUnsetRequest struct {
	GuildId string `json:"guild_id"` // 群组 ID
	UserId  string `json:"user_id"`  // 用户 ID
	RoleId  string `json:"role_id"`  // 角色 ID
}

// HandleGuildMemberRoleUnset 处理取消群组成员角色请求
func HandleGuildMemberRoleUnset(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildMemberRoleUnsetRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		dtoMemberAddRoleBody := &dto.MemberAddRoleBody{
			Channel: &dto.Channel{},
		}
		err = apiv2.MemberDeleteRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId), request.UserId, dtoMemberAddRoleBody)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
