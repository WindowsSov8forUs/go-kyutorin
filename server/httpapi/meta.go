package httpapi

import (
	"encoding/json"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"

	"github.com/satori-protocol-go/satori-model-go/pkg/login"
	"github.com/satori-protocol-go/satori-model-go/pkg/meta"
)

func init() {
	RegisterMetaHandler("", HandlerMeta)
	RegisterMetaHandler("webhook.create", HandlerWebHookCreate)
	RegisterMetaHandler("webhook.delete", HandlerWebHookDelete)
}

// MetaResponse 获取元信息响应
type MetaResponse meta.Meta

// WebHookCreateRequest 创建 WebHook 请求
type WebHookCreateRequest struct {
	URL   string `json:"url"`             // WebHook 地址
	Token string `json:"token,omitempty"` // 鉴权令牌
}

// WebHookDeleteRequest 移除 WebHook 请求
type WebHookDeleteRequest struct {
	URL string `json:"url"` // WebHook 地址
}

// HandlerMeta 处理获取元信息请求
func HandlerMeta(message *MetaActionMessage) (any, APIError) {
	var response MetaResponse

	bots := processor.GetBots()
	for platform, bot := range bots {
		login := login.Login{
			Sn:       processor.GenerateLoginSn(),
			Platform: platform,
			User:     bot,
			Status:   processor.GetStatus(platform),
			Adapter:  "kyutorin",
			Features: processor.Features(),
		}
		response.Logins = append(response.Logins, &login)
	}
	response.ProxyUrls = processor.ProxyUrls()

	return response, nil
}

// HandlerWebHookCreate 处理创建 WebHook 请求
func HandlerWebHookCreate(message *MetaActionMessage) (any, APIError) {
	var request WebHookCreateRequest
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	err = instance.webHookManager.CreateWebHook(request.URL, request.Token)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	return gin.H{}, nil
}

// HandlerWebHookDelete 处理移除 WebHook 请求
func HandlerWebHookDelete(message *MetaActionMessage) (any, APIError) {
	var request WebHookDeleteRequest
	err := json.Unmarshal(message.Data(), &request)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	err = instance.webHookManager.DeleteWebHook(request.URL)
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	return gin.H{}, nil
}
