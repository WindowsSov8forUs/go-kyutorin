package login

import "github.com/dezhishen/satori-model-go/pkg/user"

type LoginStatus uint8

const (
	OFFLINE    LoginStatus = 0 // 离线
	ONLINE     LoginStatus = 1 // 在线
	CONNECT    LoginStatus = 2 // 连接中
	DISCONNECT LoginStatus = 3 // 断开连接
	RECONNECT  LoginStatus = 4 // 重新连接
)

type Login struct {
	User     *user.User  `json:"user,omitempty"`     // 用户对象
	SelfId   string      `json:"self_id,omitempty"`  // 平台账号
	Platform string      `json:"platform,omitempty"` // 平台名称
	Status   LoginStatus `json:"status"`             // 登录状态
}
