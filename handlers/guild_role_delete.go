package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.role", "delete", HandleGuildRoleDelete)
}

// GuildRoleDeleteRequest 删除群组角色请求
type GuildRoleDeleteRequest struct {
	GuildId string `json:"guild_id"` // 群组 ID
	RoleId  string `json:"role_id"`  // 角色 ID
}

// HandleGuildRoleDelete 处理删除群组角色请求
func HandleGuildRoleDelete(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildRoleDeleteRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		err = apiv2.DeleteRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId))
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
