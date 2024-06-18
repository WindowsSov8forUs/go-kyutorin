package server

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/WindowsSov8forUs/go-kyutorin/operation"
)

// WebHook WebHook 客户端
type WebHook struct {
	url    string        // WebHook 地址
	token  string        // 鉴权令牌
	client *resty.Client // HTTP 客户端
	mu     sync.Mutex    // 互斥锁
}

// StartWebHook 启动 WebHook 客户端
func StartWebHook(url string, server *Server) *WebHook {
	// 创建 WebHook 客户端
	webhook := &WebHook{
		url:    url,
		token:  server.conf.Satori.Token,
		client: resty.New(),
	}

	// 设置请求头
	webhook.client.SetHeader("Content-Type", "application/json")
	if webhook.token != "" {
		webhook.client.SetHeader("Authorization", "Bearer "+webhook.token)
	}

	// 设置超时时间
	if server.conf.Satori.WebHook.Timeout > 0 {
		webhook.client.SetTimeout(time.Duration(server.conf.Satori.WebHook.Timeout) * time.Second)
	}

	// 返回 WebHook 客户端
	return webhook
}

// CreateWebHook 创建 WebHook 客户端
func (server *Server) CreateWebHook(url string) {
	// 创建 WebHook 客户端
	webhook := StartWebHook(url, server)

	// 添加 WebHook 客户端
	server.rwMutex.Lock()
	defer server.rwMutex.Unlock()
	server.webhooks = append(server.webhooks, webhook)
}

// DelWebHook 删除 WebHook 客户端
func (server *Server) DelWebHook(url string) {
	// 删除 WebHook 客户端
	server.rwMutex.Lock()
	defer server.rwMutex.Unlock()
	for i, webhook := range server.webhooks {
		if webhook.GetURL() == url {
			server.webhooks = append(server.webhooks[:i], server.webhooks[i+1:]...)
			return
		}
	}
}

// PostEvent 发送事件
func (w *WebHook) PostEvent(event *operation.Event) error {
	// 加锁
	w.mu.Lock()
	defer w.mu.Unlock()

	// 发送并接收响应
	response, err := w.client.R().
		SetBody(event).
		Post(w.url)
	if err != nil {
		return err
	}

	// 分类处理响应状态码
	if response.StatusCode() >= 200 && response.StatusCode() < 300 {
		// 能够顺利处理鉴权并处理请求
		return nil
	} else if response.StatusCode() >= 400 && response.StatusCode() < 500 {
		// 鉴权失败
		return ErrUnauthorized
	} else if response.StatusCode() >= 500 {
		return ErrServerError
	}

	return nil
}

// GetURL 获取 WebHook 地址
func (w *WebHook) GetURL() string {
	return w.url
}
