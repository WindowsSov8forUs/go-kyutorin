package dto

// OPCode 通用 op 码
type OPCode int

// OPCode
const (
	DispatchEvent OPCode = iota
	WSHeartbeat
	WSIdentity
	_ // Presence Update
	_ // Voice State Update
	_
	WSResume
	WSReconnect
	_ // Request Guild Members
	WSInvalidSession
	WSHello
	WSHeartbeatAck
	HTTPCallbackAck
	HTTPCallbackValidation
)

// opMeans op 对应的含义字符串标识
var opMeans = map[OPCode]string{
	DispatchEvent:          "Event",
	WSHeartbeat:            "Heartbeat",
	WSIdentity:             "Identity",
	WSResume:               "Resume",
	WSReconnect:            "Reconnect",
	WSInvalidSession:       "InvalidSession",
	WSHello:                "Hello",
	WSHeartbeatAck:         "HeartbeatAck",
	HTTPCallbackAck:        "HTTPCallbackAck",
	HTTPCallbackValidation: "回调地址验证",
}

// OPMeans 返回 op 含义
func OPMeans(op OPCode) string {
	means, ok := opMeans[op]
	if !ok {
		means = "unknown"
	}
	return means
}
