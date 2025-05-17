package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/guildrole"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.role.list", HandleGuildRoleList)
}

// RequestGuildRoleList 获取群组角色列表请求
type RequestGuildRoleList struct {
	GuildId string `json:"guild_id"`       // 群组 ID
	Next    string `json:"next,omitempty"` // 分页令牌
}

// ResponseGuildRoleList 获取群组角色列表响应
type ResponseGuildRoleList guildrole.GuildRoleList

// HandleGuildRoleList 处理获取群组角色列表请求
func HandleGuildRoleList(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildRoleList
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseGuildRoleList

		dtoGuildRoles, err := apiv2.Roles(context.TODO(), request.GuildId)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		for _, item := range dtoGuildRoles.Roles {
			guildRole, err := convertDtoRoleToGuildRole(item)
			if err != nil {
				return gin.H{}, &InternalServerError{err}
			}
			response.Data = append(response.Data, &guildRole)
		}

		return response, nil
	}

	return defaultResource(message)
}

// convertDtoRoleToGuildRole 将 dto.Role 转换为 guildrole.GuildRole
func convertDtoRoleToGuildRole(dtoRole *dto.Role) (guildrole.GuildRole, error) {
	var guildRole guildrole.GuildRole

	guildRole.Id = string(dtoRole.ID)
	guildRole.Name = dtoRole.Name

	return guildRole, nil
}
