package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.role.delete", HandleGuildRoleDelete)
}

// RequestGuildRoleDelete 删除群组角色请求
type RequestGuildRoleDelete struct {
	GuildId string `json:"guild_id"` // 群组 ID
	RoleId  string `json:"role_id"`  // 角色 ID
}

// HandleGuildRoleDelete 处理删除群组角色请求
func HandleGuildRoleDelete(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildRoleDelete
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		err = apiv2.DeleteRole(context.TODO(), request.GuildId, dto.RoleID(request.RoleId))
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}
