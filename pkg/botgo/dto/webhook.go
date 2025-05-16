package dto

import (
	"fmt"
)

type Config struct {
	Host      string // WebHook 服务器监听地址
	Path      string // WebHook 服务器监听路径
	Port      int    // WebHook 服务器监听端口
	AppId     int    // QQ 机器人 Id
	BotSecret string // 机器人密钥
}

// String 输出配置字符串
func (c *Config) String() string {
	return fmt.Sprintf("[wh][%s:%d%s]",
		c.Host, c.Port, c.Path)
}

type WHValidationRequest struct {
	PlainToken string `json:"plain_token"` // 需要计算签名的字符串
	EventTs    string `json:"event_ts"`    // 计算签名使用时间戳
}

type WHValidationResponse struct {
	PlainToken string `json:"plain_token"` // 需要计算签名的字符串
	Signature  string `json:"signature"`   // 签名
}
