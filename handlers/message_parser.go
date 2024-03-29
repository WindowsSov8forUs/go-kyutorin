package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/WindowsSov8forUs/go-kyutorin/fileserver"
	"github.com/WindowsSov8forUs/go-kyutorin/image"
	"github.com/WindowsSov8forUs/go-kyutorin/mp4"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/silk"
)

// getSrcKey 获取 src 的 key
func getSrcKey(src, messageType string) string {
	// 分析 src 获取 hash
	// 检查是否是 URL
	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" && u.Host != "" {
		// url 并不支持缓存
		return ""
	}

	// 检查是否是 data/mime 字符串
	re := regexp.MustCompile(`data:(.*);base64,(.*)`)
	matches := re.FindStringSubmatch(src)
	if len(matches) == 3 {
		// 解析 data/mime 字符串
		mimeType := matches[1]
		base64Data := matches[2]

		// 解析 mime 类型
		mimeParts := strings.Split(mimeType, "/")
		if len(mimeParts) != 2 {
			log.Errorf("错误的 mime 类型: %s", mimeType)
			return ""
		}

		// 解析 base64 编码
		var data []byte
		data, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			log.Errorf("解析 base64 编码失败: %s", err.Error())
			return ""
		}

		// 获取 hash 值
		hash := fileserver.GetHash(data)
		return messageType + ":" + hash
	}

	// 检查是否是本地文件
	if _, err := os.Stat(src); err == nil {
		// 读取文件数据
		data, err := os.ReadFile(src)
		if err != nil {
			return ""
		}

		// 获取 hash 值
		hash := fileserver.GetHash(data)
		return messageType + ":" + hash
	}

	return ""
}

// saveSrcToURL 将文件资源字符串保存并返回一个 URL
func saveSrcToURL(src string) string {
	// 检查是否是 URL
	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return src
	}

	// 检查是否是 data/mime 字符串
	re := regexp.MustCompile(`data:(.*);base64,(.*)`)
	matches := re.FindStringSubmatch(src)
	if len(matches) == 3 {
		// 解析 data/mime 字符串
		mimeType := matches[1]
		base64Data := matches[2]

		// 解析 mime 类型
		mimeParts := strings.Split(mimeType, "/")
		if len(mimeParts) != 2 {
			log.Errorf("错误的 mime 类型: %s", mimeType)
			return ""
		}
		fileType := mimeParts[0]

		// 判断是否为音频文件
		var isAudio bool
		if fileType == "audio" {
			isAudio = true
		}

		// 判断是否为视频文件
		var isVideo bool
		if fileType == "video" {
			isVideo = true
		}

		// 判断是否为图像文件
		var isImage bool
		if fileType == "image" {
			isImage = true
		}

		// 解析 base64 编码
		var data []byte
		data, err = base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			log.Errorf("解析 base64 编码失败: %s", err.Error())
			return ""
		}
		if isAudio {
			// 判断并转码
			data, err = convertAudioToSilk(data)
			if err != nil {
				log.Errorf("转码音频文件失败: %s", err.Error())
				return ""
			}
		} else if isVideo {
			// 判断并转码
			data, err = convertVideoToMP4(data)
			if err != nil {
				log.Errorf("转码视频文件失败: %s", err.Error())
				return ""
			}
		} else if isImage {
			// 判断并转码
			data, err = convertImage(data)
			if err != nil {
				log.Errorf("转码图像文件失败: %s", err.Error())
				return ""
			}
		}
		if err != nil {
			return ""
		}

		// 保存文件
		u := fileserver.SaveFile(data)
		if err != nil {
			log.Errorf("保存文件失败: %s", err.Error())
			return ""
		}

		// 返回 URL
		return u
	}

	// 检查是否是本地文件
	if _, err := os.Stat(src); err == nil {
		// 读取文件数据
		data, err := os.ReadFile(src)
		if err != nil {
			return ""
		}

		// 判断是否为音频文件
		fileType := http.DetectContentType(data)
		if strings.HasPrefix(fileType, "audio/") {
			// 判断并转码
			data, err = convertAudioToSilk(data)
			if err != nil {
				log.Errorf("转码音频文件失败: %s", err.Error())
				return ""
			}
		} else if strings.HasPrefix(fileType, "video/") {
			// 判断并转码
			data, err = convertVideoToMP4(data)
			if err != nil {
				log.Errorf("转码视频文件失败: %s", err.Error())
				return ""
			}
		} else if strings.HasPrefix(fileType, "image/") {
			// 判断并转码
			data, err = convertImage(data)
			if err != nil {
				log.Errorf("转码图像文件失败: %s", err.Error())
				return ""
			}
		} else {
			return ""
		}

		// 保存文件
		u := fileserver.SaveFile(data)

		// 返回 URL
		return u
	}

	log.Errorf("无法解析的资源字符串: %s", src)
	return ""
}

// convertAudioToSilk 将音频文件转换为 silk 格式
func convertAudioToSilk(data []byte) ([]byte, error) {
	// 判断并转码
	if !silk.IsAMRorSILK(data) {
		mimeType, ok := silk.CheckAudio(bytes.NewReader(data))
		if !ok {
			return nil, fmt.Errorf("错误的音频格式: %s", mimeType)
		}
		data = silk.EncoderSilk(data)
		return data, nil
	}
	return data, nil
}

// convertVideoToMP4 将视频文件转换为 MP4 格式
func convertVideoToMP4(data []byte) ([]byte, error) {
	// 判断并转码
	if !mp4.IsMP4(data) {
		mimeType, ok := mp4.CheckVideo(bytes.NewReader(data))
		if !ok {
			return nil, fmt.Errorf("错误的视频格式: %s", mimeType)
		}
		data = mp4.EncoderMP4(data)
		return data, nil
	}
	return data, nil
}

// convertImage 将图像文件转换为可用格式
func convertImage(data []byte) ([]byte, error) {
	// 判断并转码
	if !image.IsGIForPNGorJPG(data) {
		mimeType, ok := image.CheckImage(bytes.NewReader(data))
		if !ok {
			return nil, fmt.Errorf("错误的视频格式: %s", mimeType)
		}
		data = image.EncoderImage(data)
		return data, nil
	}
	return data, nil
}
