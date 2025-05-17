package botgo

import (
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/sessions/local"
)

// defaultWebhookManager 默认实现的 webhook manager 为本地版本
// 如果业务要自行实现分布式的 webhook 管理，则实现 WebhookManager 后替换掉 defaultWebhookManager
var defaultWebhookManager WebhookManager = local.NewWebhook()

// WebhookManager 接口，管理 webhook
type WebhookManager interface {
	// Start 启动 webhook
	Start(config *dto.Config, certFile string, keyFile string) error
}
