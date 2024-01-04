package image

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
	"github.com/disintegration/imaging"
)

const (
	HeaderJPG  string = "\xFF\xD8"
	HeaderPNG  string = "\x89PNG\r\n\x1a\n"
	HeaderGIF  string = "GIF87a"
	HeaderGIF2 string = "GIF89a"
)

const cachePath = "data/cache"

const limit = 4 * 1024

// IsGIForPNGorJPG 判断是否为 GIF/PNG/JPG
func IsGIForPNGorJPG(file []byte) bool {
	if len(file) < 8 {
		return false
	}

	if bytes.HasPrefix(file, []byte(HeaderGIF)) || bytes.HasPrefix(file, []byte(HeaderGIF2)) {
		return true
	} else if bytes.HasPrefix(file, []byte(HeaderPNG)) {
		return true
	} else if bytes.HasPrefix(file, []byte(HeaderJPG)) {
		return true
	}

	return false
}

// CheckImage 判断给定图像流是否为合法图像
func CheckImage(readSeeker io.ReadSeeker) (string, bool) {
	t := scanType(readSeeker)
	if strings.Contains(t, "image") {
		return t, true
	}
	return t, false
}

// scanType 扫描格式
func scanType(readerSeeker io.ReadSeeker) string {
	_, _ = readerSeeker.Seek(0, io.SeekStart)
	defer readerSeeker.Seek(0, io.SeekStart)
	in := make([]byte, limit)
	_, _ = readerSeeker.Read(in)
	return http.DetectContentType(in)
}

// EncoderImage 重编码图像
func EncoderImage(data []byte) []byte {
	hash := md5.New()
	_, err := hash.Write(data)
	if err != nil {
		log.Warn("计算 md5 时出错。")
		return nil
	}
	name := hex.EncodeToString(hash.Sum(nil))
	mp4 := encode(data, name)
	return mp4
}

// encode 编码为 MP4
func encode(data []byte, name string) (imageData []byte) {
	// 0. 创建缓存目录
	err := createDirectoryIfNotExist(cachePath)
	if err != nil {
		log.Warnf("创建图像缓存目录失败: %v", err)
	}

	// 1. 创建临时文件
	rawPath := path.Join(cachePath, name)
	err = os.WriteFile(rawPath, data, os.ModePerm)
	if err != nil {
		log.Errorf("创建临时文件失败: %v", err)
		return nil
	}
	defer os.Remove(rawPath)

	// 2. 检查图像格式
	img, err := imaging.Open(rawPath)
	if err != nil {
		log.Errorf("打开临时图像文件失败: %v", err)
		return nil
	}
	reader := bytes.NewReader(data)
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		log.Errorf("解码临时图像文件失败: %v", err)
		return nil
	}

	// 3. 根据不同图像格式处理
	buffer := new(bytes.Buffer)
	switch format {
	case "bmp":
		// 转换为 JPG
		err = jpeg.Encode(buffer, img, nil)
		if err != nil {
			log.Errorf("转换图像格式失败: %v", err)
			return nil
		}
		imageData = buffer.Bytes()
	default:
		// 转换为 PNG
		err = png.Encode(buffer, img)
		if err != nil {
			log.Errorf("转换图像格式失败: %v", err)
			return nil
		}
		imageData = buffer.Bytes()
	}

	return imageData
}

// createDirectoryIfNotExist 检查目录是否存在，不存在则创建
func createDirectoryIfNotExist(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// 创建目录
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
