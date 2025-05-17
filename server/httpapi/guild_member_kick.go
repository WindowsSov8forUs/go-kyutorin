package httpapi

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("guild.member.kick", HandleGuildMemberKick)
}

// RequestGuildMemberKick 踢出群组成员请求
type RequestGuildMemberKick struct {
	GuildId   string `json:"guild_id"`            // 群组 ID
	UserId    string `json:"user_id"`             // 用户 ID
	Permanent bool   `json:"permanent,omitempty"` // 是否永久踢出 (无法再次加入群组)
}

// HandleGuildMemberKick 处理踢出群组成员请求
func HandleGuildMemberKick(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var request RequestGuildMemberKick
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	if message.Platform == "qqguild" {
		// 根据 Permanent 字段值选择不同的处理函数
		if request.Permanent {
			err = api.DeleteGuildMember(context.TODO(), request.GuildId, request.UserId, setMemberDeleteOpts)
		} else {
			err = api.DeleteGuildMember(context.TODO(), request.GuildId, request.UserId)
		}
		if err != nil {
			return gin.H{}, &InternalServerError{err}
		}

		return gin.H{}, nil
	}

	return defaultResource(message)
}

// setMemberDeleteOpts 设置 dto.MemberDeleteOpts 的 AddBlackList 为 true
func setMemberDeleteOpts(opts *dto.MemberDeleteOpts) {
	opts.AddBlackList = true
}
