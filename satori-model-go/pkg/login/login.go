package login

import "github.com/satori-protocol-go/satori-model-go/pkg/user"

type LoginStatus int32

const (
	OFFLINE    LoginStatus = 0
	ONLINE     LoginStatus = 1
	CONNECT    LoginStatus = 2
	DISCONNECT LoginStatus = 3
	RECONNECT  LoginStatus = 4
)

type Login struct {
	User     *user.User  `json:"user"`
	SelfId   string      `json:"self_id"`
	Platform string      `json:"platform"`
	Status   LoginStatus `json:"status"`
}
