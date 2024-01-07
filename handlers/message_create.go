package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/WindowsSov8forUs/go-kyutorin/callapi"
	"github.com/WindowsSov8forUs/go-kyutorin/database"
	"github.com/WindowsSov8forUs/go-kyutorin/echo"
	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/WindowsSov8forUs/go-kyutorin/processor"

	"github.com/dezhishen/satori-model-go/pkg/channel"
	"github.com/dezhishen/satori-model-go/pkg/guild"
	"github.com/dezhishen/satori-model-go/pkg/guildmember"
	satoriMessage "github.com/dezhishen/satori-model-go/pkg/message"
	"github.com/dezhishen/satori-model-go/pkg/user"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/dto/keyboard"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("message", "create", HandleMessageCreate)
}

// MessageCreateRequest 发送消息请求
type MessageCreateRequest struct {
	ChannelId string `json:"channel_id"` // 频道 ID
	Content   string `json:"content"`    // 消息内容
}

// MessageCreateResponse 发送消息响应
type MessageCreateResponse []satoriMessage.Message

// HandleMessageCreate 处理发送消息请求
func HandleMessageCreate(api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	var request MessageCreateRequest
	err := json.Unmarshal(message.Data, &request)
	if err != nil {
		return "", callapi.ErrBadRequest
	}

	if message.Platform == "qqguild" {
		var response MessageCreateResponse

		// 尝试获取私聊频道，若没有获取则视为群组频道
		guildId := echo.GetDirectChannelGuild(request.ChannelId)
		if guildId == "" {
			// 输出日志
			log.Infof("发送消息到频道 %s : %s", request.ChannelId, request.Content)

			var dtoMessageToCreate = &dto.MessageToCreate{}
			dtoMessageToCreate, err = convertToMessageToCreate(request.Content)
			if err != nil {
				return "", err
			}
			var dtoMessage *dto.Message
			dtoMessage, err = api.PostMessage(context.TODO(), request.ChannelId, dtoMessageToCreate)
			if err != nil {
				return "", err
			}
			messageResponse, err := convertDtoMessageToMessage(dtoMessage)
			if err != nil {
				return "", err
			}
			response = append(response, *messageResponse)
		} else {
			// 输出日志
			log.Infof("发送消息到私聊频道 %s : %s", request.ChannelId, request.Content)

			var dtoMessageToCreate = &dto.MessageToCreate{}
			var dtoDirectMessage *dto.DirectMessage
			dtoMessageToCreate, err = convertToMessageToCreate(request.Content)
			if err != nil {
				return "", err
			}
			dtoDirectMessage.ChannelID = request.ChannelId
			dtoDirectMessage.GuildID = guildId
			var dtoMessage *dto.Message
			dtoMessage, err = api.PostDirectMessage(context.TODO(), dtoDirectMessage, dtoMessageToCreate)
			if err != nil {
				return "", err
			}
			messageResponse, err := convertDtoMessageToMessage(dtoMessage)
			if err != nil {
				return "", err
			}
			response = append(response, *messageResponse)
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	} else if message.Platform == "qq" {
		var response MessageCreateResponse

		// 尝试获取消息类型
		openIdType := echo.GetOpenIdType(request.ChannelId)
		if openIdType == "private" {
			// 输出日志
			log.Infof("发送消息到用户 %s : %s", request.ChannelId, request.Content)

			// 是私聊频道
			var dtoMessageToCreate = &dto.MessageToCreate{}
			dtoMessageToCreate, err = convertToMessageToCreateV2(request.Content, request.ChannelId, openIdType, apiv2)
			if err != nil {
				return "", err
			}
			var dtoC2CMessageResponse *dto.C2CMessageResponse
			dtoC2CMessageResponse, err = api.PostC2CMessage(context.TODO(), request.ChannelId, dtoMessageToCreate)
			if err != nil {
				return "", err
			}
			messageResponse, err := convertDtoMessageToMessage(dtoC2CMessageResponse.Message)
			if err != nil {
				return "", err
			}
			response = append(response, *messageResponse)
		} else {
			// 是群聊频道
			openIdType = "group"

			// 输出日志
			log.Infof("发送消息到群 %s : %s", request.ChannelId, request.Content)

			var dtoMessageToCreate = &dto.MessageToCreate{}
			dtoMessageToCreate, err = convertToMessageToCreateV2(request.Content, request.ChannelId, openIdType, apiv2)
			if err != nil {
				return "", err
			}
			var dtoGroupMessageResponse *dto.GroupMessageResponse
			dtoGroupMessageResponse, err = api.PostGroupMessage(context.TODO(), request.ChannelId, dtoMessageToCreate)
			if err != nil {
				return "", err
			}
			messageResponse, err := convertDtoMessageV2ToMessage(dtoGroupMessageResponse.Message)
			if err != nil {
				return "", err
			}
			response = append(response, *messageResponse)
		}

		var responseData []byte
		responseData, err = json.Marshal(response)
		if err != nil {
			return "", err
		}
		return string(responseData), nil
	}

	return "", callapi.ErrMethodNotAllowed
}

// convertToMessageToCreate 转换为消息体结构
func convertToMessageToCreate(content string) (*dto.MessageToCreate, error) {
	// 将文本消息内容转换为 satoriMessage.MessageElement
	fmt.Println(content)
	elements, err := satoriMessage.Parse(content)
	if err != nil {
		return nil, err
	}
	for _, element := range elements {
		fmt.Println(element)
	}

	// 处理 satoriMessage.MessageElement
	var dtoMessageToCreate = &dto.MessageToCreate{}
	err = parseElementsInMessageToCreate(elements, dtoMessageToCreate)
	if err != nil {
		return nil, err
	}
	temp, _ := json.Marshal(dtoMessageToCreate)
	fmt.Println(string(temp))
	return dtoMessageToCreate, nil
}

// parseElementsInMessageToCreate 将 Satori 消息元素转换为消息体结构
func parseElementsInMessageToCreate(elements []satoriMessage.MessageElement, dtoMessageToCreate *dto.MessageToCreate) error {
	// 处理 satoriMessage.MessageElement
	fmt.Println(elements)
	for _, element := range elements {
		// 根据元素类型进行处理
		switch e := element.(type) {
		case *satoriMessage.MessageElementText:
			dtoMessageToCreate.Content += escape(e.Content)
		case *satoriMessage.MessageElementAt:
			if e.Type == "all" {
				dtoMessageToCreate.Content += "@everyone"
			} else {
				if e.Id != "" {
					dtoMessageToCreate.Content += fmt.Sprintf("<@%s>", e.Id)
				} else {
					continue
				}
			}
		case *satoriMessage.MessageElementSharp:
			dtoMessageToCreate.Content += fmt.Sprintf("<#%s>", e.Id)
		case *satoriMessage.MessageElementA:
			dtoMessageToCreate.Content += escape(e.Href)
		case *satoriMessage.MessageElementImg:
			// TODO: 通过本地文件或 base64 上传图片
			dtoMessageToCreate.Image = e.Src
		case *satoriMessage.MessageElementAudio:
			// 频道不支持音频消息
			continue
		case *satoriMessage.MessageElementVideo:
			// 频道不支持视频消息
			continue
		case *satoriMessage.MessageElementFile:
			// 频道不支持文件消息
			continue
		// TODO: 修饰元素全部视为子元素集合，或许可以变成 dto.markdown ？
		case *satoriMessage.MessageElementStrong:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementEm:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementIns:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementDel:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementSpl:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementCode:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementSup:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementSub:
			// 递归调用
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElmentBr:
			dtoMessageToCreate.Content += "\n"
		case *satoriMessage.MessageElmentP:
			dtoMessageToCreate.Content += "\n"
			// 视为子元素集合
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
			dtoMessageToCreate.Content += "\n"
		case *satoriMessage.MessageElementMessage:
			// 视为子元素集合，目前不支持视为转发消息
			parseElementsInMessageToCreate(e.Children, dtoMessageToCreate)
		case *satoriMessage.MessageElementQuote:
			// 遍历子元素，只会处理第一个 satoriMessage.MessageElementMessage 元素
			for _, child := range e.Children {
				if m, ok := child.(*satoriMessage.MessageElementMessage); ok {
					if m.Id != "" {
						dtoMessageReference := &dto.MessageReference{
							MessageID: m.Id,
						}
						dtoMessageToCreate.MessageReference = dtoMessageReference
						break
					}
				}
			}
		case *satoriMessage.MessageElementPassive:
			// 被动元素处理，作为消息发送的基础
			dtoMessageToCreate.MsgID = e.Id
			dtoMessageToCreate.MsgSeq = e.Seq
		case *satoriMessage.MessageElementButton:
			// TODO: 频道的按钮怎么处理？
			continue
		default:
			continue
		}
	}
	return nil
}

// convertToMessageToCreateV2 转换为 V2 消息体结构
func convertToMessageToCreateV2(content string, OpenId string, messageType string, apiv2 openapi.OpenAPI) (*dto.MessageToCreate, error) {
	// 将文本消息内容转换为 satoriMessage.MessageElement
	elements, err := satoriMessage.Parse(content)
	if err != nil {
		return nil, err
	}

	// 处理 satoriMessage.MessageElement
	var dtoMessageToCreate = &dto.MessageToCreate{}
	err = parseElementsInMessageToCreateV2(elements, dtoMessageToCreate, OpenId, messageType, apiv2)
	if err != nil {
		return nil, err
	}

	// Content 字段不能为空
	if dtoMessageToCreate.Content == "" {
		dtoMessageToCreate.Content = " "
	}
	return dtoMessageToCreate, nil
}

// parseElementsInMessageToCreateV2 将 Satori 消息元素转换为 V2 消息体结构
func parseElementsInMessageToCreateV2(elements []satoriMessage.MessageElement, dtoMessageToCreate *dto.MessageToCreate, OpenId string, messageType string, apiv2 openapi.OpenAPI) error {
	// 处理 satoriMessage.MessageElement
	for _, element := range elements {
		// 根据元素类型进行处理
		switch e := element.(type) {
		case *satoriMessage.MessageElementText:
			dtoMessageToCreate.Content += escape(e.Content)
		case *satoriMessage.MessageElementAt:
			// 群聊/单聊目前似乎是不支持的
			continue
		case *satoriMessage.MessageElementSharp:
			// 群聊/单聊目前似乎是不支持的
			continue
		case *satoriMessage.MessageElementA:
			dtoMessageToCreate.Content += escape(e.Href)
		case *satoriMessage.MessageElementImg:
			if dtoMessageToCreate.Media.FileInfo != "" {
				// 富媒体信息只支持一个
				continue
			}
			if e.Cache {
				// 获取 key
				key := getSrcKey(e.Src, messageType)
				// 尝试获取缓存
				fileInfo, ok := database.GetImageCache(key)
				if ok {
					dtoMessageToCreate.Media.FileInfo = fileInfo
					dtoMessageToCreate.MsgType = 7
					continue
				}
			}
			dtoRichMediaMessage := generateDtoRichMediaMessage(dtoMessageToCreate.MsgID, e)
			if dtoRichMediaMessage == nil {
				continue
			}
			key := getSrcKey(e.Src, messageType)
			if messageType == "private" {
				fileInfo, err := uploadMediaPrivate(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			} else {
				fileInfo, err := uploadMedia(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			}
			dtoMessageToCreate.MsgType = 7
		case *satoriMessage.MessageElementAudio:
			if dtoMessageToCreate.Media.FileInfo != "" {
				// 富媒体信息只支持一个
				continue
			}
			if e.Cache {
				// 获取 key
				key := getSrcKey(e.Src, messageType)
				// 尝试获取缓存
				fileInfo, ok := database.GetAudioCache(key)
				if ok {
					dtoMessageToCreate.Media.FileInfo = fileInfo
					dtoMessageToCreate.MsgType = 7
					continue
				}
			}
			dtoRichMediaMessage := generateDtoRichMediaMessage(dtoMessageToCreate.MsgID, e)
			if dtoRichMediaMessage == nil {
				continue
			}
			key := getSrcKey(e.Src, messageType)
			if messageType == "private" {
				fileInfo, err := uploadMediaPrivate(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			} else {
				fileInfo, err := uploadMedia(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			}
			dtoMessageToCreate.MsgType = 7
		case *satoriMessage.MessageElementVideo:
			if dtoMessageToCreate.Media.FileInfo != "" {
				// 富媒体信息只支持一个
				continue
			}
			if e.Cache {
				// 获取 key
				key := getSrcKey(e.Src, messageType)
				// 尝试获取缓存
				fileInfo, ok := database.GetVideoCache(key)
				if ok {
					dtoMessageToCreate.Media.FileInfo = fileInfo
					dtoMessageToCreate.MsgType = 7
					continue
				}
			}
			dtoRichMediaMessage := generateDtoRichMediaMessage(dtoMessageToCreate.MsgID, e)
			if dtoRichMediaMessage == nil {
				continue
			}
			key := getSrcKey(e.Src, messageType)
			if messageType == "private" {
				fileInfo, err := uploadMediaPrivate(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			} else {
				fileInfo, err := uploadMedia(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			}
			dtoMessageToCreate.MsgType = 7
		case *satoriMessage.MessageElementFile:
			// TODO: 本地缓冲
			if dtoMessageToCreate.Media.FileInfo != "" {
				// 富媒体信息只支持一个
				continue
			}
			dtoRichMediaMessage := generateDtoRichMediaMessage(dtoMessageToCreate.MsgID, e)
			if dtoRichMediaMessage == nil {
				continue
			}
			key := getSrcKey(e.Src, messageType)
			if messageType == "private" {
				fileInfo, err := uploadMediaPrivate(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			} else {
				fileInfo, err := uploadMedia(context.TODO(), OpenId, dtoRichMediaMessage, apiv2, key)
				if err != nil {
					return err
				}
				dtoMessageToCreate.Media.FileInfo = fileInfo
			}
			dtoMessageToCreate.MsgType = 7
		// TODO: 修饰元素全部视为子元素集合，或许可以变成 dto.markdown ？
		case *satoriMessage.MessageElementStrong:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementEm:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementIns:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementDel:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementSpl:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementCode:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementSup:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementSub:
			// 递归调用
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElmentBr:
			dtoMessageToCreate.Content += "\n"
		case *satoriMessage.MessageElmentP:
			dtoMessageToCreate.Content += "\n"
			// 视为子元素集合
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
			dtoMessageToCreate.Content += "\n"
		case *satoriMessage.MessageElementMessage:
			// 视为子元素集合，目前不支持视为转发消息
			parseElementsInMessageToCreateV2(e.Children, dtoMessageToCreate, OpenId, messageType, apiv2)
		case *satoriMessage.MessageElementQuote:
			// 遍历子元素，只会处理第一个 satoriMessage.MessageElementMessage 元素
			for _, child := range e.Children {
				if m, ok := child.(*satoriMessage.MessageElementMessage); ok {
					if m.Id != "" {
						dtoMessageReference := &dto.MessageReference{
							MessageID: m.Id,
						}
						dtoMessageToCreate.MessageReference = dtoMessageReference
						break
					}
				}
			}
		case *satoriMessage.MessageElementPassive:
			// 被动元素处理，作为消息发送的基础
			dtoMessageToCreate.MsgID = e.Id
			dtoMessageToCreate.MsgSeq = e.Seq
		case *satoriMessage.MessageElementButton:
			dtoMessageToCreate.Keyboard = convertButtonToKeyboard(e)
		default:
			continue
		}
	}
	return nil
}

// convertDtoMessageToMessage 将收到的消息响应转化为符合 Satori 协议的消息
func convertDtoMessageToMessage(dtoMessage *dto.Message) (*satoriMessage.Message, error) {
	var message satoriMessage.Message

	message.Id = dtoMessage.ID
	message.Content = strings.TrimSpace(processor.ConvertToMessageContent(dtoMessage))

	// 判断消息类型
	if dtoMessage.ChannelID != "" {
		// 判断是否为私聊频道
		guildId := echo.GetDirectChannelGuild(dtoMessage.ChannelID)
		var channelType channel.ChannelType
		if guildId == "" {
			// 不是私聊频道
			channelType = channel.CHANNEL_TYPE_TEXT
		} else {
			// 是私聊频道
			channelType = channel.CHANNEL_TYPE_DIRECT
		}
		channel := &channel.Channel{
			Id:   dtoMessage.ChannelID,
			Type: channelType,
		}
		guild := &guild.Guild{
			Id: dtoMessage.GuildID,
		}
		var guildMember *guildmember.GuildMember
		if dtoMessage.Member != nil {
			guildMember = &guildmember.GuildMember{}
			if dtoMessage.Member != nil {
				guildMember.Nick = dtoMessage.Member.Nick
			}
			if dtoMessage.Author != nil {
				guildMember.Avatar = dtoMessage.Author.Avatar
			}
		}
		user := &user.User{
			Id:     dtoMessage.Author.ID,
			Name:   dtoMessage.Author.Username,
			Avatar: dtoMessage.Author.Avatar,
			IsBot:  dtoMessage.Author.Bot,
		}

		// 获取时间
		if dtoMessage.Member != nil {
			time, err := dtoMessage.Member.JoinedAt.Time()
			if err != nil {
				return nil, err
			}
			guildMember.JoinedAt = time.Unix()
		}

		message.Channel = channel
		message.Guild = guild
		message.Member = guildMember
		message.User = user
	} else {
		// 判断是否为单聊
		if dtoMessage.GroupID != "" {
			// 是群聊
			channel := &channel.Channel{
				Id:   dtoMessage.GroupID,
				Type: channel.CHANNEL_TYPE_TEXT,
			}
			guild := &guild.Guild{
				Id: dtoMessage.GroupID,
			}
			var guildMember *guildmember.GuildMember
			if dtoMessage.Member == nil {
				guildMember = &guildmember.GuildMember{}
				if dtoMessage.Member != nil {
					guildMember.Nick = dtoMessage.Member.Nick
				}
				if dtoMessage.Author != nil {
					guildMember.Avatar = dtoMessage.Author.Avatar
				}
			}
			user := &user.User{
				Id:     dtoMessage.Author.ID,
				Name:   dtoMessage.Author.Username,
				Avatar: dtoMessage.Author.Avatar,
				IsBot:  dtoMessage.Author.Bot,
			}

			// 获取时间
			if dtoMessage.Member != nil {
				time, err := dtoMessage.Member.JoinedAt.Time()
				if err != nil {
					return nil, err
				}
				guildMember.JoinedAt = time.Unix()
			}

			message.Channel = channel
			message.Guild = guild
			message.Member = guildMember
			message.User = user
		} else {
			// 是单聊
			channel := &channel.Channel{
				Id:   dtoMessage.Author.ID,
				Type: channel.CHANNEL_TYPE_DIRECT,
			}
			user := &user.User{
				Id:     dtoMessage.Author.ID,
				Name:   dtoMessage.Author.Username,
				Avatar: dtoMessage.Author.Avatar,
				IsBot:  dtoMessage.Author.Bot,
			}
			message.Channel = channel
			message.User = user
		}
	}

	time, err := dtoMessage.Timestamp.Time()
	if err != nil {
		return nil, err
	}
	message.CreateAt = time.Unix()

	return &message, nil
}

// convertDtoMessageV2ToMessage 将收到的 V2 消息响应转化为符合 Satori 协议的消息
func convertDtoMessageV2ToMessage(dtoMessage *dto.Message) (*satoriMessage.Message, error) {
	var message satoriMessage.Message

	message.Id = dtoMessage.ID
	message.Content = strings.TrimSpace(processor.ConvertToMessageContent(dtoMessage))

	// 判断是否为单聊
	if dtoMessage.GroupCode != "" {
		// 是群聊
		channel := &channel.Channel{
			Id:   dtoMessage.GroupCode,
			Type: channel.CHANNEL_TYPE_TEXT,
		}
		guild := &guild.Guild{
			Id: dtoMessage.GroupCode,
		}
		var guildMember *guildmember.GuildMember
		if dtoMessage.Member == nil {
			guildMember = &guildmember.GuildMember{}
			if dtoMessage.Member != nil {
				guildMember.Nick = dtoMessage.Member.Nick
			}
			if dtoMessage.Author != nil {
				guildMember.Avatar = dtoMessage.Author.Avatar
			}
		}
		var u *user.User
		if dtoMessage.Author != nil {
			u = &user.User{
				Id:     dtoMessage.Author.ID,
				Name:   dtoMessage.Author.Username,
				Avatar: dtoMessage.Author.Avatar,
				IsBot:  dtoMessage.Author.Bot,
			}
		}

		// 获取时间
		if dtoMessage.Member != nil {
			time, err := dtoMessage.Member.JoinedAt.Time()
			if err != nil {
				return nil, err
			}
			guildMember.JoinedAt = time.Unix()
		}

		message.Channel = channel
		message.Guild = guild
		message.Member = guildMember
		message.User = u
	} else {
		// 是单聊
		// TODO: 目前没有实际运用场景，很可能需要更改
		channel := &channel.Channel{
			Id:   dtoMessage.Author.ID,
			Type: channel.CHANNEL_TYPE_DIRECT,
		}
		user := &user.User{
			Id:     dtoMessage.Author.ID,
			Name:   dtoMessage.Author.Username,
			Avatar: dtoMessage.Author.Avatar,
			IsBot:  dtoMessage.Author.Bot,
		}
		message.Channel = channel
		message.User = user
	}

	time, err := dtoMessage.Timestamp.Time()
	if err != nil {
		return nil, err
	}
	message.CreateAt = time.Unix()

	return &message, nil
}

// convertButtonToKeyboard 将 Satori 协议的按钮转换为 QQ 的按钮
func convertButtonToKeyboard(button *satoriMessage.MessageElementButton) *keyboard.MessageKeyboard {
	// TODO: 或许需要支持更多的方式
	var messageKeyboard keyboard.MessageKeyboard

	messageKeyboard.ID = button.Id

	return &messageKeyboard
}

// uploadMedia 上传媒体并返回FileInfo
func uploadMedia(ctx context.Context, groupID string, richMediaMessage *dto.RichMediaMessage, apiv2 openapi.OpenAPI, key string) (string, error) {
	// 调用API来上传媒体
	messageReturn, err := apiv2.PostGroupMessage(ctx, groupID, richMediaMessage)
	if err != nil {
		return "", err
	}
	// 将获取到的信息保存到数据库
	switch richMediaMessage.FileType {
	case 1:
		// 图片
		err = database.SaveImageCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存图片缓存失败: %s", err.Error())
		}
	case 2:
		// 视频
		err = database.SaveVideoCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存视频缓存失败: %s", err.Error())
		}
	case 3:
		// 音频
		err = database.SaveAudioCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存音频缓存失败: %s", err.Error())
		}
	}
	// 返回上传后的FileInfo
	return messageReturn.MediaResponse.FileInfo, nil
}

// uploadMedia 上传媒体并返回FileInfo
func uploadMediaPrivate(ctx context.Context, userID string, richMediaMessage *dto.RichMediaMessage, apiv2 openapi.OpenAPI, key string) (string, error) {
	// 调用API来上传媒体
	messageReturn, err := apiv2.PostC2CMessage(ctx, userID, richMediaMessage)
	if err != nil {
		return "", err
	}
	// 将获取到的信息保存到数据库
	switch richMediaMessage.FileType {
	case 1:
		// 图片
		err = database.SaveImageCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存图片缓存失败: %s", err.Error())
		}
	case 2:
		// 视频
		err = database.SaveVideoCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存视频缓存失败: %s", err.Error())
		}
	case 3:
		// 音频
		err = database.SaveAudioCache(key, messageReturn.MediaResponse.FileInfo, int64(messageReturn.MediaResponse.TTL))
		if err != nil {
			log.Warnf("保存音频缓存失败: %s", err.Error())
		}
	}
	// 返回上传后的FileInfo
	return messageReturn.MediaResponse.FileInfo, nil
}

// generateDtoRichMediaMessage 创建 dto.RichMediaMessage
func generateDtoRichMediaMessage(id string, element satoriMessage.MessageElement) *dto.RichMediaMessage {
	// TODO: 根据本地文件或 base64 上传文件

	var dtoRichMediaMessage *dto.RichMediaMessage

	// 根据 element 的类型来创建 dto.RichMediaMessage
	switch e := element.(type) {
	case *satoriMessage.MessageElementImg:
		url := saveSrcToURL(e.Src)
		dtoRichMediaMessage = &dto.RichMediaMessage{
			EventID:    id,
			FileType:   1,
			URL:        url,
			SrvSendMsg: false,
		}
	case *satoriMessage.MessageElementVideo:
		url := saveSrcToURL(e.Src)
		dtoRichMediaMessage = &dto.RichMediaMessage{
			EventID:    id,
			FileType:   2,
			URL:        url,
			SrvSendMsg: false,
		}
	case *satoriMessage.MessageElementAudio:
		url := saveSrcToURL(e.Src)
		dtoRichMediaMessage = &dto.RichMediaMessage{
			EventID:    id,
			FileType:   3,
			URL:        url,
			SrvSendMsg: false,
		}
	case *satoriMessage.MessageElementFile:
		// TODO: 暂不开放
		return nil
	default:
		return nil
	}

	return dtoRichMediaMessage
}

// escape 转义
func escape(source string) string {
	source = strings.ReplaceAll(source, "&", "&amp;")
	source = strings.ReplaceAll(source, "<", "&lt;")
	source = strings.ReplaceAll(source, ">", "&gt;")
	return source
}
