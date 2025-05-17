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
	RegisterHandler("guild.role.create", HandleGuildRoleCreate)
}

// RequestGuildRoleCreate 创建群组角色请求
type RequestGuildRoleCreate struct {
	GuildId string              `json:"guild_id"` // 群组 ID
	Role    guildrole.GuildRole `json:"role"`     // 角色数据
}

// ResponseGuildRoleCreate 创建群组角色响应
type ResponseGuildRoleCreate guildrole.GuildRole

// HandleGuildRoleCreate 处理创建群组角色请求
func HandleGuildRoleCreate(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildRoleCreate
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		var response ResponseGuildRoleCreate

		dtoRole, err := convertGuildRoleToDtoRole(request.Role)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		dtoUpdateResult, err := apiv2.PostRole(context.TODO(), request.GuildId, dtoRole)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		guildRole, err := convertDtoRoleToGuildRole(dtoUpdateResult.Role)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		response = ResponseGuildRoleCreate(guildRole)

		return response, nil
	}

	return defaultResource(message)
}

// convertGuildRoleToDtoRole 将 guildrole.GuildRole 转换为 dto.Role
func convertGuildRoleToDtoRole(guildRole guildrole.GuildRole) (*dto.Role, error) {
	dtoRole := &dto.Role{
		ID:   dto.RoleID(guildRole.Id),
		Name: guildRole.Name,
	}

	return dtoRole, nil
}
