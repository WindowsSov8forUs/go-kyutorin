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
	callapi.RegisterHandler("guild.role", "list", HandleGuildRoleList)
}

// GuildRoleListRequest 获取群组角色列表请求
type GuildRoleListRequest struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// GuildRoleListResponse 获取群组角色列表响应
type GuildRoleListResponse guildrole.GuildRoleList

// HandleGuildRoleList 处理获取群组角色列表请求
func HandleGuildRoleList(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildRoleListRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response GuildRoleListResponse

		dtoGuildRoles, err := apiv2.Roles(context.TODO(), request.GuildId)
		if err != nil {
			return "", err
		}

		for _, item := range dtoGuildRoles.Roles {
			guildRole, err := convertDtoRoleToGuildRole(item)
			if err != nil {
				return "", err
			}
			response.Data = append(response.Data, guildRole)
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// convertDtoRoleToGuildRole 将 dto.Role 转换为 guildrole.GuildRole
func convertDtoRoleToGuildRole(dtoRole *dto.Role) (guildrole.GuildRole, error) {
	var guildRole guildrole.GuildRole

	guildRole.Id = string(dtoRole.ID)
	guildRole.Name = dtoRole.Name

	return guildRole, nil
}
