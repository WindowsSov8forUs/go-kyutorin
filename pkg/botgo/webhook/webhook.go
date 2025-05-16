// package webhook SDK 需要实现的 webhook 定义。
package webhook

import (
	"runtime"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
	"github.com/tencent-connect/botgo/log"
)

var (
	ServerImpl WebHook
)

// Register 注册webhook实现
func Register(wh WebHook) {
	ServerImpl = wh
}

// PanicBufLen Panic 堆栈大小
var PanicBufLen = 1024

// PanicHandler 处理websocket场景的 panic ，打印堆栈
func PanicHandler(e interface{}, session *dto.Session) {
	buf := make([]byte, PanicBufLen)
	buf = buf[:runtime.Stack(buf, false)]
	log.Errorf("[PANIC]%s\n%v\n%s\n", session, e, buf)
}

// RegisterHandlers 兼容老版本的注册方式
func RegisterHandlers(handlers ...interface{}) dto.Intent {
	return event.RegisterHandlers(handlers...)
}
