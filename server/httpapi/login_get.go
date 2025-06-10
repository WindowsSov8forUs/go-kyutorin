package httpapi

import (
	"context"

	"github.com/WindowsSov8forUs/glyccat/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("login.get", HandleLoginGet)
}

// ResponseLoginGet 获取登录信息响应
type ResponseLoginGet login.Login

// HandleLoginGet 处理获取登录信息请求
func HandleLoginGet(api, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	var response ResponseLoginGet

	var me *dto.User
	me, err := api.Me(context.TODO())
	if err != nil {
		return gin.H{}, &InternalServerError{err}
	}

	// 更新 SelfID
	processor.SelfId = me.ID

	// 构建机器人对象
	bot := &user.User{
		Id:     processor.GetBot(message.Platform).Id,
		Name:   me.Username,
		Avatar: me.Avatar,
		IsBot:  me.Bot,
	}
	processor.SetBot(message.Platform, bot)

	// 获取机器人状态
	status := processor.GetStatus(message.Platform)

	response.Sn = processor.GenerateLoginSn()
	response.Platform = message.Platform
	response.User = bot
	response.Status = status
	response.Adapter = "kyutorin"
	response.Features = processor.Features()

	return response, nil
}
