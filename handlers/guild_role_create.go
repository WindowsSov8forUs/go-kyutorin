package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/satori-protocol-go/satori-model-go/pkg/guildrole"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.role", "create", HandleGuildRoleCreate)
}

// GuildRoleCreateRequest 创建群组角色请求
type GuildRoleCreateRequest struct {
	GuildId string              `json:"guild_id"` // 群组 ID
	Role    guildrole.GuildRole `json:"role"`     // 角色数据
}

// GuildRoleCreateResponse 创建群组角色响应
type GuildRoleCreateResponse guildrole.GuildRole

// HandleGuildRoleCreate 处理创建群组角色请求
func HandleGuildRoleCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildRoleCreateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response GuildRoleCreateResponse

		dtoRole, err := convertGuildRoleToDtoRole(request.Role)
		if err != nil {
			return "", err
		}

		dtoUpdateResult, err := apiv2.PostRole(context.TODO(), request.GuildId, dtoRole)
		if err != nil {
			return "", err
		}

		guildRole, err := convertDtoRoleToGuildRole(dtoUpdateResult.Role)
		if err != nil {
			return "", err
		}

		response = GuildRoleCreateResponse(guildRole)

		responseData, err := json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// convertGuildRoleToDtoRole 将 guildrole.GuildRole 转换为 dto.Role
func convertGuildRoleToDtoRole(guildRole guildrole.GuildRole) (*dto.Role, error) {
	dtoRole := &dto.Role{
		ID:   dto.RoleID(guildRole.Id),
		Name: guildRole.Name,
	}

	return dtoRole, nil
}
