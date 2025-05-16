package webhook

import "github.com/tencent-connect/botgo/dto"

type WebHook interface {
	// New 创建一个新的webhook实例
	New(config dto.Config) WebHook
	// Listen 监听webhook事件
	Listen() error
	// Write 发送数据
	Write(message *dto.Payload) error
	// Close 关闭连接
	Close() error
}
