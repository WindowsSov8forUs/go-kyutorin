package mp4

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/WindowsSov8forUs/go-kyutorin/mylog"
)

const cachePath = "data/cache"

const limit = 4 * 1024

var HeadersMP4 []string = []string{
	"ftypisom",
	"ftypmp42",
}

// IsMP4 判断是否为 MP4 文件
func IsMP4(file []byte) bool {
	for _, header := range HeadersMP4 {
		if bytes.HasPrefix(file, []byte(header)) {
			return true
		}
	}
	return false
}

// CheckVideo 判断给定视频流是否为合法视频
func CheckVideo(readSeeker io.ReadSeeker) (string, bool) {
	t := scanType(readSeeker)
	if strings.Contains(t, "video") {
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

// EncoderMP4 编码为 MP4
func EncoderMP4(data []byte) []byte {
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
func encode(data []byte, name string) (mp4Video []byte) {
	// 0. 创建缓存目录
	err := createDirectoryIfNotExist(cachePath)
	if err != nil {
		log.Warnf("创建视频缓存目录失败: %v", err)
	}

	// 1. 创建临时文件
	rawPath := path.Join(cachePath, name)
	err = os.WriteFile(rawPath, data, os.ModePerm)
	if err != nil {
		log.Errorf("创建临时文件失败: %v", err)
		return nil
	}
	defer os.Remove(rawPath)

	// 2. 转换 MP4
	mp4Path := path.Join(cachePath, name+".mp4")
	cmd := exec.Command("ffmpeg", "-i", rawPath, "-vcodec", "libx264", "-acodec", "aac", mp4Path)
	if err := cmd.Run(); err != nil {
		log.Errorf("转换 MP4 失败: %v", err)
		return nil
	}
	mp4Video, err = os.ReadFile(mp4Path)
	if err != nil {
		log.Errorf("读取 MP4 文件失败: %v", err)
		return nil
	}
	defer os.Remove(mp4Path)

	return mp4Video
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
