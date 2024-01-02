package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/dezhishen/satori-model-go/pkg/guildrole"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.role", "update", HandleGuildRoleUpdate)
}

// GuildRoleUpdateRequest 修改群组角色请求
type GuildRoleUpdateRequest struct {
	GuildId string              `json:"guild_id"` // 群组 ID
	RoleId  string              `json:"role_id"`  // 角色 ID
	Role    guildrole.GuildRole `json:"role"`     // 角色数据
}

// HandleGuildRoleUpdate 处理修改群组角色请求
func HandleGuildRoleUpdate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildRoleUpdateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		dtoRole, err := convertGuildRoleToDtoRole(request.Role)
		if err != nil {
			return "", err
		}

		_, err = apiv2.PatchRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId), dtoRole)
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}
