package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/WindowsSov8forUs/go-kyutorin/fileserver"
	"github.com/WindowsSov8forUs/go-kyutorin/image"
	"github.com/WindowsSov8forUs/go-kyutorin/mp4"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/silk"
	"github.com/dezhishen/satori-model-go/pkg/login"
	"github.com/dezhishen/satori-model-go/pkg/user"
)

// BotMapping 机器人映射
type BotMapping struct {
	mapping map[string]*user.User
	mu      sync.Mutex
}

// StatusMapping 机器人状态映射
type StatusMapping struct {
	mapping map[string]login.LoginStatus
	mu      sync.Mutex
}

var SelfId string // 机器人 ID

var globalBotMapping = &BotMapping{
	mapping: make(map[string]*user.User),
}

var globalStatusMapping = &StatusMapping{
	mapping: make(map[string]login.LoginStatus),
}

// SetBot 设置机器人
func SetBot(platform string, bot *user.User) {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	globalBotMapping.mapping[platform] = bot
}

// GetBot 获取机器人
func GetBot(platform string) *user.User {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	return globalBotMapping.mapping[platform]
}

// GetBots 获取所有机器人
func GetBots() map[string]*user.User {
	globalBotMapping.mu.Lock()
	defer globalBotMapping.mu.Unlock()
	return globalBotMapping.mapping
}

// SetStatus 设置机器人状态
func SetStatus(platform string, status login.LoginStatus) {
	globalStatusMapping.mu.Lock()
	defer globalStatusMapping.mu.Unlock()
	globalStatusMapping.mapping[platform] = status
}

// GetStatus 获取机器人状态
func GetStatus(platform string) login.LoginStatus {
	globalStatusMapping.mu.Lock()
	defer globalStatusMapping.mu.Unlock()
	return globalStatusMapping.mapping[platform]
}

// saveSrcToURL 将文件资源字符串保存并返回一个 URL
func saveSrcToURL(src string) string {
	// 检查是否是 URL
	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return src
	}

	// 检查是否是 data/mime 字符串
	if strings.HasPrefix(src, "data:") {
		// 解析 data/mime 字符串
		parts := strings.Split(src, ",")
		if len(parts) != 2 {
			return ""
		}

		// 解析 mime 类型
		mimeParts := strings.Split(parts[0], ";")
		if len(mimeParts) != 2 {
			return ""
		}
		fileType := mimeParts[0]

		// 判断是否为音频文件
		var isAudio bool
		if strings.HasPrefix(fileType, "audio/") {
			isAudio = true
		}

		// 判断是否为视频文件
		var isVideo bool
		if strings.HasPrefix(fileType, "video/") {
			isVideo = true
		}

		// 判断是否为图像文件
		var isImage bool
		if strings.HasPrefix(fileType, "image/") {
			isImage = true
		}

		// 解析 base64 编码
		encoding := mimeParts[1]
		var data []byte
		switch encoding {
		case "base64":
			data, err = base64.StdEncoding.DecodeString(parts[1])
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
			} else {
				return ""
			}
		default:
			return ""
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
