package handlers

import (
	"context"
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("login", "get", HandleLoginGet)
}

// LoginGetResponse 获取登录信息响应
type LoginGetResponse login.Login

// HandleLoginGet 处理获取登录信息请求
func HandleLoginGet(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var response LoginGetResponse

	var me *dto.User
	me, err := api.Me(context.TODO())
	if err != nil {
		return "", err
	}

	// 构建机器人对象
	bot := &user.User{
		Id:     me.ID,
		Name:   me.Username,
		Avatar: me.Avatar,
		IsBot:  me.Bot,
	}
	processor.SetBot(message.Platform, bot)

	// 获取机器人状态
	status := processor.GetStatus(message.Platform)

	response.User = bot
	response.SelfId = processor.SelfId
	response.Platform = message.Platform
	response.Status = status

	data, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
