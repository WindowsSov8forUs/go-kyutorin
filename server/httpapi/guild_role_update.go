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
	RegisterHandler("guild.role.update", HandleGuildRoleUpdate)
}

// RequestGuildRoleUpdate 修改群组角色请求
type RequestGuildRoleUpdate struct {
	GuildId string              `json:"guild_id"` // 群组 ID
	RoleId  string              `json:"role_id"`  // 角色 ID
	Role    guildrole.GuildRole `json:"role"`     // 角色数据
}

// HandleGuildRoleUpdate 处理修改群组角色请求
func HandleGuildRoleUpdate(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildRoleUpdate
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		dtoRole, err := convertGuildRoleToDtoRole(request.Role)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		_, err = apiv2.PatchRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId), dtoRole)
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}
