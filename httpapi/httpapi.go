package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/handlers"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"
)

// ResourceMiddleware 资源中间件
func ResourceMiddleware(api openapi.OpenAPI, apiV2 openapi.OpenAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		satoriResourceAPIHandler(ctx, api, apiV2)
	}
}

// satoriResourceAPIHandler 处理 Satori 资源 API
func satoriResourceAPIHandler(c *gin.Context, api openapi.OpenAPI, apiV2 openapi.OpenAPI) {
	// 提取路径中参数
	action := c.Param("action")

	// 拆分 action 为 resource 和 method
	parts := strings.Split(action[1:], ".")
	if len(parts) < 2 {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	resource := strings.Join(parts[:len(parts)-1], ".")
	method := parts[len(parts)-1]

	// 提取请求头参数
	contentType := c.GetHeader("Content-Type")
	authorization := c.GetHeader("Authorization")
	xPlatform := c.GetHeader("X-Platform")
	xSelfID := c.GetHeader("X-Self-ID")

	// 判断请求头错误
	if contentType != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// 鉴权
	if !authorize(authorization) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	// 判断平台与 SelfID 是否正确
	bot := processor.GetBot(xPlatform)
	if bot == nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if xSelfID != processor.SelfId {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// 构建请求
	requestBody := c.Request.Body
	defer requestBody.Close()
	// 提取请求体 []byte
	requestBodyBytes, err := io.ReadAll(requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	actionMessage := callapi.NewActionMessage(resource, method, *bot, xPlatform, requestBodyBytes)

	// 调用 API
	response, err := callapi.CallAPI(api, apiV2, actionMessage)
	if err != nil {
		switch err {
		case callapi.ErrBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{})
		case callapi.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{})
		case callapi.ErrMethodNotAllowed:
			c.JSON(http.StatusMethodNotAllowed, gin.H{})
		default:
			log.Errorf("调用 API 时出错: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
		return
	}

	// 返回结果
	c.Data(http.StatusOK, "application/json", []byte(response))
}

// AdminMiddleware 管理接口中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		satoriAdminAPIHandler(ctx)
	}
}

// satoriAdminAPIHandler 处理 Satori 管理 API
func satoriAdminAPIHandler(c *gin.Context) {
	// 提取路径中参数
	action := strings.TrimPrefix(c.Param("action"), "/admin")

	// 拆分 action 为 resource 和 method
	parts := strings.Split(action[1:], ".")
	if len(parts) < 2 {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	resource := strings.Join(parts[:len(parts)-1], ".")
	method := parts[len(parts)-1]

	// 提取请求头参数
	contentType := c.GetHeader("Content-Type")
	authorization := c.GetHeader("Authorization")

	// 判断请求头错误
	if contentType != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	// 鉴权
	if !authorize(authorization) {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	// 构建请求
	requestBody := c.Request.Body
	defer requestBody.Close()
	// 提取请求体 []byte
	requestBodyBytes, err := io.ReadAll(requestBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	actionMessage := callapi.NewAdminMessage(resource, method, requestBodyBytes)

	// 调用 API
	response, err := callAdmin(actionMessage)
	if err != nil {
		switch err {
		case callapi.ErrBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{})
		case callapi.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{})
		case callapi.ErrMethodNotAllowed:
			c.JSON(http.StatusMethodNotAllowed, gin.H{})
		default:
			log.Errorf("调用 API 时出错: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{})
		}
		return
	}

	// 返回结果
	c.Data(http.StatusOK, "application/json", []byte(response))
}

// 调用 Admin 处理函数
func callAdmin(message callapi.AdminMessage) (string, error) {
	switch message.Resource {
	case "login":
		switch message.Method {
		case "list":
			log.Debugf("调用管理接口 API: %s %s", message.Resource, message.Method)
			return handlers.HandlerLoginList(message)
		default:
			return "", callapi.ErrMethodNotAllowed
		}
	case "webhook":
		switch message.Method {
		case "create":
			log.Debugf("调用管理接口 API: %s %s", message.Resource, message.Method)
			return handlers.HandlerWebHookCreate(message)
		case "delete":
			log.Debugf("调用管理接口 API: %s %s", message.Resource, message.Method)
			return handlers.HandlerWebHookDelete(message)
		default:
			return "", callapi.ErrMethodNotAllowed
		}
	default:
		return "", callapi.ErrNotFound
	}
}

// authorize 鉴权
func authorize(authorization string) bool {
	// 获取鉴权令牌
	token := config.GetSatoriToken()

	// 如果设置的令牌为空则默认不进行鉴权
	if token == "" {
		return true
	}

	// 构建并比对鉴权令牌
	return authorization == fmt.Sprintf("Bearer %s", token)
}
