package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/WindowsSov8forUs/go-kyutorin/config"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/gin-gonic/gin"
	"github.com/satori-protocol-go/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/openapi"
)

// APIError API 错误
type APIError interface {
	Error() string
	Code() int
}

// BadRequestError 请求格式错误
type BadRequestError struct {
	err error
}

func (e *BadRequestError) Error() string {
	return e.err.Error()
}

func (e *BadRequestError) Code() int {
	return http.StatusBadRequest
}

// UnauthorizedError 缺失鉴权
type UnauthorizedError struct {
	invalidToken string
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("unauthorized with wrong token: %s", e.invalidToken)
}

func (e *UnauthorizedError) Code() int {
	return http.StatusUnauthorized
}

// ForbiddenError 权限不足
type ForbiddenError struct {
	message string
}

func (e *ForbiddenError) Error() string {
	return e.message
}

func (e *ForbiddenError) Code() int {
	return http.StatusForbidden
}

// NotFoundError 资源不存在
type NotFoundError struct {
	api      string
	platform string
}

func (e *NotFoundError) Error() string {
	if e.platform == "" {
		return fmt.Sprintf(`api "%s" not found`, e.api)
	} else {
		return fmt.Sprintf(`api "%s" is not supported on %s`, e.api, e.platform)
	}
}

func (e *NotFoundError) Code() int {
	return http.StatusNotFound
}

// MethodNotAllowedError 请求方法不支持
type MethodNotAllowedError struct {
	method string
}

func (e *MethodNotAllowedError) Error() string {
	return fmt.Sprintf(`method "%s" not allowed`, e.method)
}

func (e *MethodNotAllowedError) Code() int {
	return http.StatusMethodNotAllowed
}

// InternalServerError 服务器内部错误
type InternalServerError struct {
	err error
}

func (e *InternalServerError) Error() string {
	return e.err.Error()
}

func (e *InternalServerError) Code() int {
	return http.StatusInternalServerError
}

var ErrBadRequest = errors.New("bad request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not found")
var ErrMethodNotAllowed = errors.New("method not allowed")
var ErrServerError = errors.New("server error")

// ActionMessage Satori 应用发送的 HTTP API 调用信息
type ActionMessage struct {
	API      string     // 接口
	Bot      *user.User // 机器人信息
	Platform string     // 平台
	Data     []byte     // 应用发送的数据
}

// AdminActionMessage Satori 应用发送的管理接口调用信息
type AdminActionMessage struct {
	API  string // 接口
	Data []byte // 应用发送的数据
}

// NewActionMessage 创建一个新的 ActionMessage
func NewActionMessage(api string, bot *user.User, platform string, data []byte) *ActionMessage {
	return &ActionMessage{
		API:      api,
		Bot:      bot,
		Platform: platform,
		Data:     data,
	}
}

// NewAdminActionMessage 创建一个新的 AdminActionMessage
func NewAdminActionMessage(api string, data []byte) *AdminActionMessage {
	return &AdminActionMessage{
		API:  api,
		Data: data,
	}
}

// HTTP API 处理函数
type HandlerFunc func(api, apiV2 openapi.OpenAPI, action *ActionMessage) (any, APIError)

// 管理接口的处理函数
type AdminHandlerFunc func(action *AdminActionMessage) (any, APIError)

var handlers = make(map[string]HandlerFunc)
var adminHandlers = make(map[string]AdminHandlerFunc)

// defaultResource 资源默认处理函数
func defaultResource(action *ActionMessage) (any, APIError) {
	return gin.H{}, &NotFoundError{action.API, action.Platform}
}

// defaultAdmin 管理接口默认处理函数
func defaultAdmin(action *AdminActionMessage) (any, APIError) {
	return gin.H{}, &NotFoundError{api: action.API}
}

// RegisterHandler 注册特定资源与方法的处理函数
func RegisterHandler(api string, handler HandlerFunc) {
	handlers[api] = handler
}

// RegisterAdminHandler 注册管理接口的处理函数
func RegisterAdminHandler(api string, handler AdminHandlerFunc) {
	adminHandlers[api] = handler
}

// CallAPI 调用 Satori API
func CallAPI(api, apiV2 openapi.OpenAPI, action *ActionMessage) (any, APIError) {
	if _, ok := handlers[action.API]; !ok {
		return gin.H{}, &NotFoundError{api: action.API}
	}
	return handlers[action.API](api, apiV2, action)
}

// CallAdminAPI 调用 Satori 管理 API
func CallAdminAPI(action *AdminActionMessage) (any, APIError) {
	if _, ok := adminHandlers[action.API]; !ok {
		return gin.H{}, &NotFoundError{api: action.API}
	}
	return adminHandlers[action.API](action)
}

// ResourceMiddleware 资源中间件
func ResourceMiddleware(api, apiV2 openapi.OpenAPI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		resourceAPIHandler(ctx, api, apiV2)
	}
}

// resourceAPIHandler 处理资源 API
func resourceAPIHandler(c *gin.Context, api, apiV2 openapi.OpenAPI) {
	// 提取路径中参数
	action := c.Param("action")
	method := strings.TrimLeft(action, "/")

	// 提取请求头
	contentType := c.GetHeader("Content-Type")
	authorization := c.GetHeader("Authorization")
	xPlatform := c.GetHeader("X-Platform")
	xSelfID := c.GetHeader("X-Self-ID")

	// 判断请求头错误
	if contentType != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": "content type must be application/json",
		})
		return
	}

	// 鉴权
	if !authorize(authorization) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"error":   "unauthorized",
			"message": "authorize failed with token: " + authorization,
		})
		return
	}

	// 判断平台与 SelfID 是否正确
	bot := processor.GetBot(xPlatform)
	if bot == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": fmt.Sprintf(`invalid platform "%s"`, xPlatform),
		})
		return
	}
	if xSelfID != processor.SelfId {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": fmt.Sprintf(`invalid self id "%s"`, xSelfID),
		})
		return
	}

	// 构建 Action
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": err.Error(),
		})
		return
	}
	actionMessage := NewActionMessage(method, bot, xPlatform, bodyBytes)

	// 调用 API
	response, err := CallAPI(api, apiV2, actionMessage)
	if err != nil {
		switch err.(type) {
		case *BadRequestError:
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"error":   "bad request",
				"message": err.Error(),
			})
		case *UnauthorizedError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"error":   "unauthorized",
				"message": err.Error(),
			})
		case *ForbiddenError:
			c.JSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"error":   "forbidden",
				"message": err.Error(),
			})
		case *NotFoundError:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"error":   "not found",
				"message": err.Error(),
			})
		case *MethodNotAllowedError:
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"status":  http.StatusMethodNotAllowed,
				"error":   "method not allowed",
				"message": err.Error(),
			})
		case *InternalServerError:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"error":   "server error",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"error":   "server error",
				"message": err.Error(),
			})
		}
	} else {
		// 返回结果
		c.JSON(http.StatusOK, response)
	}
}

// AdminMiddleware 管理接口中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 在内部进行判断处理
		adminAPIHandler(ctx)
	}
}

// adminAPIHandler 处理管理 API
func adminAPIHandler(c *gin.Context) {
	// 提取路径中参数
	action := c.Param("action")

	// 提取 method
	method := strings.TrimLeft(action, "/admin/")

	// 提取请求头
	contentType := c.GetHeader("Content-Type")
	authorization := c.GetHeader("Authorization")

	// 判断请求头错误
	if contentType != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": "content type must be application/json",
		})
		return
	}

	// 鉴权
	if !authorize(authorization) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  http.StatusUnauthorized,
			"error":   "unauthorized",
			"message": "unauthorized",
		})
		return
	}

	// 构建 Action
	body := c.Request.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"error":   "bad request",
			"message": err.Error(),
		})
		return
	}
	adminActionMessage := NewAdminActionMessage(method, bodyBytes)

	// 调用 API
	response, err := CallAdminAPI(adminActionMessage)
	if err != nil {
		switch err.(type) {
		case *BadRequestError:
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  http.StatusBadRequest,
				"error":   "bad request",
				"message": err.Error(),
			})
		case *UnauthorizedError:
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"error":   "unauthorized",
				"message": err.Error(),
			})
		case *ForbiddenError:
			c.JSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"error":   "forbidden",
				"message": err.Error(),
			})
		case *NotFoundError:
			c.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"error":   "not found",
				"message": err.Error(),
			})
		case *MethodNotAllowedError:
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"status":  http.StatusMethodNotAllowed,
				"error":   "method not allowed",
				"message": err.Error(),
			})
		case *InternalServerError:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"error":   "server error",
				"message": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"error":   "server error",
				"message": err.Error(),
			})
		}
	} else {
		// 返回结果
		c.JSON(http.StatusOK, response)
	}
}

// authorize 鉴权
func authorize(authorization string) bool {
	// 获取令牌
	token := config.GetSatoriToken()

	// 如果令牌设置为空则不鉴权
	if token == "" {
		return true
	}

	// 构建并比对鉴权
	return authorization == "Bearer "+token
}
