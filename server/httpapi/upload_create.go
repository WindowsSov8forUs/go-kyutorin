package httpapi

import (
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("upload.create", HandleUploadCreate)
}

// HandleUploadCreate 处理文件上传请求
func HandleUploadCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {

}
