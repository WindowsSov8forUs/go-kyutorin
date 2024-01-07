package webhook

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"
	"github.com/WindowsSov8forUs/go-kyutorin/signaling"
)

// WebHook WebHook 客户端
type WebHook struct {
	url    string        // WebHook 地址
	token  string        // 鉴权令牌
	client *resty.Client // HTTP 客户端
	mu     sync.Mutex    // 互斥锁
}

var Timeout uint32 = 0 // 全局设定 WebHook 超时时间

// StartWebHook 启动 WebHook 客户端
func StartWebHook(url string, token string, timeout uint32) *WebHook {
	// 创建 WebHook 客户端
	webhook := &WebHook{
		url:    url,
		token:  token,
		client: resty.New(),
	}

	// 设置请求头
	webhook.client.SetHeader("Content-Type", "application/json")
	if webhook.token != "" {
		webhook.client.SetHeader("Authorization", "Bearer "+webhook.token)
	}

	// 设置超时时间
	if timeout > 0 {
		webhook.client.SetTimeout(time.Duration(timeout) * time.Second)
	}

	// 返回 WebHook 客户端
	return webhook
}

// CreateWebHook 创建 WebHook 客户端
func CreateWebHook(url string, token string) {
	// 创建 WebHook 客户端
	webhook := StartWebHook(url, token, Timeout)

	// 添加 WebHook 客户端
	processor.SetWebHookClient(webhook)
}

// DelWebHook 删除 WebHook 客户端
func DelWebHook(url string) {
	processor.DelWebHookClient(url)
}

// PostEvent 发送事件
func (w *WebHook) PostEvent(event *signaling.Event) error {
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
		return callapi.ErrUnauthorized
	} else if response.StatusCode() >= 500 {
		return callapi.ErrServerError
	}

	return nil
}

// GetURL 获取 WebHook 地址
func (w *WebHook) GetURL() string {
	return w.url
}
