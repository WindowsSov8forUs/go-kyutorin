package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("guild.member", "kick", HandleGuildMemberKick)
}

// GuildMemberKickRequest 踢出群组成员请求
type GuildMemberKickRequest struct {
	GuildId   string `json:"guild_id"`            // 群组 ID
	UserId    string `json:"user_id"`             // 用户 ID
	Permanent bool   `json:"permanent,omitempty"` // 是否永久踢出 (无法再次加入群组)
}

// HandleGuildMemberKick 处理踢出群组成员请求
func HandleGuildMemberKick(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request GuildMemberKickRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		// 根据 Permanent 字段值选择不同的处理函数
		if request.Permanent {
			err = api.DeleteGuildMember(context.TODO(), request.GuildId, request.UserId, setMemberDeleteOpts)
		} else {
			err = api.DeleteGuildMember(context.TODO(), request.GuildId, request.UserId)
		}
		if err != nil {
			return "", err
		}

		return "", nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// setMemberDeleteOpts 设置 dto.MemberDeleteOpts 的 AddBlackList 为 true
func setMemberDeleteOpts(opts *dto.MemberDeleteOpts) {
	opts.AddBlackList = true
}
