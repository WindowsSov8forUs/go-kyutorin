package processor

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/WindowsSov8forUs/glyccat/pkg/image"
	"github.com/WindowsSov8forUs/glyccat/pkg/mp4"
	"github.com/WindowsSov8forUs/glyccat/pkg/silk"
	satoriMessage "github.com/satori-protocol-go/satori-model-go/pkg/message"
	"github.com/tencent-connect/botgo/dto"
)

type fileSrc struct {
	MimeType string
	Data     []byte
}

// GetReader 获取文件资源的读取器
func (src *fileSrc) GetReader() (io.Reader, error) {
	if src == nil || src.Data == nil {
		return nil, fmt.Errorf("无效的文件资源")
	}

	// 返回一个 bytes.Reader
	return bytes.NewReader(src.Data), nil
}

// ParseSrc 解析 src 字符串
func ParseSrc(src string) (string, *fileSrc, error) {
	// 检查是否为 URL ，是则直接返回
	u, err := url.Parse(src)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return src, nil, nil
	}

	// 检查是否为 data/mime 字符串
	re := regexp.MustCompile(`data:(.*);base64,(.*)`)
	matches := re.FindStringSubmatch(src)
	if len(matches) == 3 {
		// 解析 data/mime 字符串
		mimeType := matches[1]
		base64Data := matches[2]

		// 解析 mime 类型
		mimeParts := strings.Split(mimeType, "/")
		if len(mimeParts) != 2 {
			return "", nil, fmt.Errorf("错误的 mime 类型: %s", mimeType)
		}

		// 解析 base64 编码
		data, err := base64.StdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", nil, fmt.Errorf("解析 base64 编码失败: %w", err)
		}

		return "", &fileSrc{MimeType: mimeType, Data: data}, nil
	}

	// 检查是否是本地文件
	re = regexp.MustCompile(`file:///(.*)`)
	matches = re.FindStringSubmatch(src)
	if len(matches) == 2 {
		path, err := url.PathUnescape(matches[1])
		if err != nil {
			return "", nil, fmt.Errorf("解析文件路径失败: %w", err)
		}

		if _, err := os.Stat(path); err == nil {
			// 读取文件数据
			data, err := os.ReadFile(path)
			if err != nil {
				return "", nil, fmt.Errorf("读取文件失败: %w", err)
			}
			return "", &fileSrc{MimeType: http.DetectContentType(data), Data: data}, nil
		}
	}

	return "", nil, fmt.Errorf("无法解析的资源字符串: %s", src)
}

// ParseSrcToString 解析 src 字符串，返回 src 标识用字符串
func ParseSrcToString(src string) (string, error) {
	url, fileSrc, err := ParseSrc(src)
	if err != nil {
		return "", err
	}

	if url != "" {
		// 如果是 URL，直接返回
		return url, nil
	}

	if fileSrc != nil {
		return base64.StdEncoding.EncodeToString(fileSrc.Data), nil
	}

	return "", fmt.Errorf("无法解析的资源字符串: %s", src)
}

// ParseSrcToAvailavle 解析 src 字符串，返回可用的 URL 或 base64 字符串
func ParseSrcToAvailavle(src string) (string, string, error) {
	url, fileSrc, err := ParseSrc(src)
	if err != nil {
		return "", "", err
	}

	if url != "" {
		// 如果是 URL，直接返回
		return url, "", nil
	}

	if fileSrc != nil {
		// 如果是 base64 字符串，返回没有 base64 头的 base64 编码字符串
		// 对于图片、音频与视频，需要转码为可接受的格式
		data, err := convertToAvailableFormat(fileSrc)
		if err != nil {
			return "", "", fmt.Errorf("转换文件格式失败: %w", err)
		}
		base64Data := base64.StdEncoding.EncodeToString(data)
		return "", base64Data, nil
	}

	return "", "", fmt.Errorf("无法解析的资源字符串: %s", src)
}

// convertToAvailableFormat 将文件资源转换为可用格式
func convertToAvailableFormat(src *fileSrc) ([]byte, error) {
	if src == nil || src.Data == nil {
		return nil, fmt.Errorf("无效的文件资源")
	}

	// 判断并转码
	if strings.HasPrefix(src.MimeType, "audio/") {
		return convertAudioToSilk(src.Data)
	} else if strings.HasPrefix(src.MimeType, "video/") {
		return convertVideoToMP4(src.Data)
	} else if strings.HasPrefix(src.MimeType, "image/") {
		return convertImage(src.Data)
	}

	return src.Data, nil
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
