package httpapi

import (
	"github.com/WindowsSov8forUs/go-kyutorin/fileserver"
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/gin-gonic/gin"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	RegisterHandler("upload.create", HandleUploadCreate)
}

// HandleUploadCreate 处理文件上传请求
func HandleUploadCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message *ActionMessage) (any, APIError) {
	response := gin.H{}

	// 解析表单
	form, err := message.Ctx.MultipartForm()
	if err != nil {
		return gin.H{}, &BadRequestError{err}
	}

	// 遍历处理文件
	for name, files := range form.File {
		// 每个 name 只会对应一个文件
		file := files[0]
		reader, err := file.Open()
		if err != nil {
			log.Errorf("从请求表单中读取文件 %s 时发生错误: %v", name, err)
			continue
		}
		contentType := file.Header.Get("Content-Type")

		meta, err := fileserver.SaveFile(reader, message.Platform, message.Bot.Id, file.Filename, contentType)
		if err != nil {
			log.Errorf("保存文件 %s 时发生错误: %v", name, err)
			continue
		}

		response[name] = fileserver.InternalURL(meta)
	}

	return response, nil
}
