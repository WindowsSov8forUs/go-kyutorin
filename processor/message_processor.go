package processor

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
	"github.com/WindowsSov8forUs/go-kyutorin/log"
	"github.com/WindowsSov8forUs/go-kyutorin/pkg/image"
	"github.com/WindowsSov8forUs/go-kyutorin/pkg/mp4"
	"github.com/WindowsSov8forUs/go-kyutorin/pkg/silk"
	satoriMessage "github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/tencent-connect/botgo/dto"
)

// GetSrcKey 获取 src 的 key
func GetSrcKey(src, messageType string) string {
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

// SaveSrcToURL 将文件资源字符串保存并返回一个 URL
func SaveSrcToURL(src string) (string, string) {
	// 检查是否是 URL
	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return src, ""
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
			return "", ""
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
			return "", ""
		}
		if isAudio {
			// 判断并转码
			data, err = convertAudioToSilk(data)
			if err != nil {
				log.Errorf("转码音频文件失败: %s", err.Error())
				return "", ""
			}
		} else if isVideo {
			// 判断并转码
			data, err = convertVideoToMP4(data)
			if err != nil {
				log.Errorf("转码视频文件失败: %s", err.Error())
				return "", ""
			}
		} else if isImage {
			// 判断并转码
			data, err = convertImage(data)
			if err != nil {
				log.Errorf("转码图像文件失败: %s", err.Error())
				return "", ""
			}
		}
		if err != nil {
			return "", ""
		}

		// 保存文件
		u := fileserver.SaveFile(data)

		// 返回 URL
		return u, fileserver.GetHash(data)
	}

	// 检查是否是本地文件
	re = regexp.MustCompile(`file:///(.*)`)
	matches = re.FindStringSubmatch(src)
	path, err := url.PathUnescape(matches[1])
	if err == nil {
		if _, err := os.Stat(path); err == nil {
			// 读取文件数据
			data, err := os.ReadFile(path)
			if err != nil {
				log.Errorf("failed to read file: %s", err.Error())
				return "", ""
			}

			// 判断是否为音频文件
			fileType := http.DetectContentType(data)
			if strings.HasPrefix(fileType, "audio/") {
				// 判断并转码
				data, err = convertAudioToSilk(data)
				if err != nil {
					log.Errorf("转码音频文件失败: %s", err.Error())
					return "", ""
				}
			} else if strings.HasPrefix(fileType, "video/") {
				// 判断并转码
				data, err = convertVideoToMP4(data)
				if err != nil {
					log.Errorf("转码视频文件失败: %s", err.Error())
					return "", ""
				}
			} else if strings.HasPrefix(fileType, "image/") {
				// 判断并转码
				data, err = convertImage(data)
				if err != nil {
					log.Errorf("转码图像文件失败: %s", err.Error())
					return "", ""
				}
			} else {
				return "", ""
			}

			// 保存文件
			u := fileserver.SaveFile(data)

			// 返回 URL
			return u, fileserver.GetHash(data)
		}
	}

	log.Errorf("无法解析的资源字符串: %s", src)
	return "", ""
}

// convertAudioToSilk 将音频文件转换为 silk 格式
func convertAudioToSilk(data []byte) ([]byte, error) {
	// 判断并转码
	if !silk.IsAMRorSILK(data) {
		mimeType, ok := silk.CheckAudio(bytes.NewReader(data))
		if !ok {
			return nil, fmt.Errorf("错误的音频格式: %s", mimeType)
		}
		return silk.EncoderSilk(data)
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
		return mp4.EncoderMP4(data)
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
		return image.EncoderImage(data)
	}
	return data, nil
}

// getMessageLog 获取消息日志
func getMessageLog(data interface{}) string {
	// 强制类型转换获取 Message 结构
	var msg *dto.Message
	var isAt bool = false // 是否为 at 消息
	switch v := data.(type) {
	case *dto.GroupATMessageData:
		msg = (*dto.Message)(v)
		isAt = true
	case *dto.ATMessageData:
		msg = (*dto.Message)(v)
	case *dto.MessageData:
		msg = (*dto.Message)(v)
	case *dto.DirectMessageData:
		msg = (*dto.Message)(v)
	case *dto.C2CMessageData:
		msg = (*dto.Message)(v)
	case *dto.Message:
		msg = v
	default:
		return ""
	}
	var messageStrings = []string{}

	// 使用正则表达式查找特殊格式字符
	re := regexp.MustCompile(`(@everyone|<@!\d+>|<#\d+>|<emoji:\d+>)`)

	// 获取所有匹配项的位置
	indexes := re.FindAllStringIndex(msg.Content, -1)

	// 根据匹配项的位置分割字符串
	var result []string
	start := 0
	for _, index := range indexes {
		if start != index[0] {
			part := msg.Content[start:index[0]]
			if part != "" {
				result = append(result, part)
			}
		}
		result = append(result, msg.Content[index[0]:index[1]])
		start = index[1]
	}
	if start != len(msg.Content) {
		part := msg.Content[start:]
		if part != "" {
			result = append(result, part)
		}
	}

	// 匹配检查每个结果
	for _, r := range result {
		if r == "@everyone" {
			if msg.MentionEveryone {
				messageStrings = append(messageStrings, "@全体成员")
			}
		} else if strings.HasPrefix(r, "<@!") && strings.HasSuffix(r, ">") {
			// 提取 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<@!")
			for _, mention := range msg.Mentions {
				if mention.ID == id {
					messageStrings = append(messageStrings, "@"+mention.Username)
					break
				}
			}
		} else if strings.HasPrefix(r, "<#") && strings.HasSuffix(r, ">") {
			// 提取频道 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<#")
			messageStrings = append(messageStrings, "#"+id)
		} else if strings.HasPrefix(r, "<emoji:") && strings.HasSuffix(r, ">") {
			// 提取 emoji ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<emoji:")
			messageStrings = append(messageStrings, "[emoji"+id+"]")
		} else {
			// 普通文本
			messageStrings = append(messageStrings, r)
		}
	}

	// 处理 Attachments 字段
	for _, attachment := range msg.Attachments {
		if attachment == nil {
			continue
		}
		// 根据 ContentType 前缀判断文件类型
		switch {
		case strings.HasPrefix(attachment.ContentType, "image"):
			image := "[图片]"
			if strings.HasPrefix(attachment.URL, "http") {
				image += "(" + attachment.URL + ")"
			} else {
				image += "(https://" + attachment.URL + ")"
			}
			messageStrings = append(messageStrings, image)
		case strings.HasPrefix(attachment.ContentType, "audio"):
			audio := "[语音]"
			if strings.HasPrefix(attachment.URL, "http") {
				audio += "(" + attachment.URL + ")"
			} else {
				audio += "(https://" + attachment.URL + ")"
			}
			messageStrings = append(messageStrings, audio)
		case strings.HasPrefix(attachment.ContentType, "video"):
			video := "[视频]"
			if strings.HasPrefix(attachment.URL, "http") {
				video += "(" + attachment.URL + ")"
			} else {
				video += "(https://" + attachment.URL + ")"
			}
			messageStrings = append(messageStrings, video)
		default:
			file := "[文件]"
			if strings.HasPrefix(attachment.URL, "http") {
				file += "(" + attachment.URL + ")"
			} else {
				file += "(https://" + attachment.URL + ")"
			}
			messageStrings = append(messageStrings, file)
		}
	}

	// 添加 embed 消息
	for _, embed := range msg.Embeds {
		if embed == nil {
			continue
		}
		embedString := fmt.Sprintf("[embed](%s)", embed.Title)
		messageStrings = append(messageStrings, embedString)
	}

	// 添加 ark 消息
	if msg.Ark != nil {
		arkString := fmt.Sprintf("[ark](%d)", msg.Ark.TemplateID)
		messageStrings = append(messageStrings, arkString)
	}

	// 添加消息回复
	if msg.MessageReference != nil {
		quoteString := "[回复消息" + msg.MessageReference.MessageID + "] "
		messageStrings = append(messageStrings, quoteString)
	}

	// 添加消息前 at
	if isAt {
		bot := GetBot("qq") // 获取 qq 平台机器人实例
		if bot != nil {
			atString := "@" + bot.Name
			messageStrings = append(messageStrings, atString)
		}
	}

	// 拼接消息
	var content string
	for index, segment := range messageStrings {
		content += segment
		if index != len(messageStrings)-1 {
			content += " "
		}
	}

	return content
}

// ConvertToMessageContent 将收到的消息转化为符合 Satori 协议的消息
func ConvertToMessageContent(data interface{}) string {
	// 强制类型转换获取 Message 结构
	var msg *dto.Message
	var isAt bool = false // 是否为 at 消息
	switch v := data.(type) {
	case *dto.GroupATMessageData:
		msg = (*dto.Message)(v)
		isAt = true
	case *dto.ATMessageData:
		msg = (*dto.Message)(v)
	case *dto.MessageData:
		msg = (*dto.Message)(v)
	case *dto.DirectMessageData:
		msg = (*dto.Message)(v)
	case *dto.C2CMessageData:
		msg = (*dto.Message)(v)
	case *dto.Message:
		msg = v
	default:
		return ""
	}
	var messageSegments []satoriMessage.MessageElement

	// 使用正则表达式查找特殊格式字符
	re := regexp.MustCompile(`(@everyone|<@!\d+>|<#\d+>|<emoji:\d+>)`)

	// 获取所有匹配项的位置
	indexes := re.FindAllStringIndex(msg.Content, -1)

	// 根据匹配项的位置分割字符串
	var result []string
	start := 0
	for _, index := range indexes {
		if start != index[0] {
			part := msg.Content[start:index[0]]
			if part != "" {
				result = append(result, part)
			}
		}
		result = append(result, msg.Content[index[0]:index[1]])
		start = index[1]
	}
	if start != len(msg.Content) {
		part := msg.Content[start:]
		if part != "" {
			result = append(result, part)
		}
	}

	// 匹配检查每个结果
	for _, r := range result {
		if r == "@everyone" {
			if msg.MentionEveryone {
				at := satoriMessage.MessageElementAt{Type: "all"}
				messageSegments = append(messageSegments, &at)
			}
		} else if strings.HasPrefix(r, "<@!") && strings.HasSuffix(r, ">") {
			// 提取 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<@!")
			for _, mention := range msg.Mentions {
				if mention.ID == id {
					// 如果是机器人自己则进行替换
					//
					// 这种时候一般来说都是频道
					if id == SelfId {
						id = GetBot("qqguild").Id
					}

					at := satoriMessage.MessageElementAt{
						Id:   id,
						Name: mention.Username,
					}
					messageSegments = append(messageSegments, &at)
					break
				}
			}
		} else if strings.HasPrefix(r, "<#") && strings.HasSuffix(r, ">") {
			// 提取频道 ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<#")
			sharp := satoriMessage.MessageElementSharp{Id: id}
			messageSegments = append(messageSegments, &sharp)
		} else if strings.HasPrefix(r, "<emoji:") && strings.HasSuffix(r, ">") {
			// 提取 emoji ID
			id := strings.TrimPrefix(strings.TrimSuffix(r, ">"), "<emoji:")
			emoji := satoriMessage.NewMessageElementExtend(
				"qqguild:emoji",
				map[string]string{"id": id},
			)
			messageSegments = append(messageSegments, emoji)
		} else {
			// 普通文本
			text := satoriMessage.MessageElementText{Content: r}
			messageSegments = append(messageSegments, &text)
		}
	}

	// 处理 Attachments 字段
	for _, attachment := range msg.Attachments {
		// 根据 ContentType 前缀判断文件类型
		switch {
		case strings.HasPrefix(attachment.ContentType, "image"):
			image := satoriMessage.MessageElementImg{}
			if strings.HasPrefix(attachment.URL, "http") {
				image.Src = attachment.URL
			} else {
				image.Src = "https://" + attachment.URL
			}

			// 添加可能存在的长宽属性
			if attachment.Width != 0 {
				image.Width = uint32(attachment.Width)
			}
			if attachment.Height != 0 {
				image.Height = uint32(attachment.Height)
			}
			messageSegments = append(messageSegments, &image)
		case strings.HasPrefix(attachment.ContentType, "audio"):
			audio := satoriMessage.MessageElementAudio{}
			if strings.HasPrefix(attachment.URL, "http") {
				audio.Src = attachment.URL
			} else {
				audio.Src = "https://" + attachment.URL
			}
			messageSegments = append(messageSegments, &audio)
		case strings.HasPrefix(attachment.ContentType, "video"):
			video := satoriMessage.MessageElementVideo{}
			if strings.HasPrefix(attachment.URL, "http") {
				video.Src = attachment.URL
			} else {
				video.Src = "https://" + attachment.URL
			}
			messageSegments = append(messageSegments, &video)
		default:
			file := satoriMessage.MessageElementFile{}
			if strings.HasPrefix(attachment.URL, "http") {
				file.Src = attachment.URL
			} else {
				file.Src = "https://" + attachment.URL
			}
			messageSegments = append(messageSegments, &file)
		}
	}

	// 添加消息回复
	if msg.MessageReference != nil {
		quote := satoriMessage.MessageElementQuote{
			Id: msg.MessageReference.MessageID,
		}

		// 添加为第一个元素
		messageSegments = append([]satoriMessage.MessageElement{&quote}, messageSegments...)
	}

	// 添加消息前 at
	if isAt {
		bot := GetBot("qq") // 获取 qq 平台机器人实例
		at := satoriMessage.MessageElementAt{
			Id:   bot.Id,
			Name: bot.Name,
		}
		// 添加为第一个元素
		messageSegments = append([]satoriMessage.MessageElement{&at}, messageSegments...)
	}

	// 拼接消息
	var content string
	for _, segment := range messageSegments {
		content += segment.Stringify()
	}
	return content
}
