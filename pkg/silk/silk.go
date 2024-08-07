package silk

import (
	"bytes"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
)

//go:embed exec/*
var silkCodecs embed.FS

const (
	HeaderAmr  string = "#!AMR"         // AMR 文件头
	HeaderSilk string = "\x02#!SILK_V3" // Silkv3 文件头
)

const cachePath = "data/cache"

const limit = 4 * 1024

// IsAMRorSILK 判断是否是 AMR 或 SILK 文件
func IsAMRorSILK(file []byte) bool {
	return bytes.HasPrefix(file, []byte(HeaderAmr)) || bytes.HasPrefix(file, []byte(HeaderSilk))
}

// CheckAudio 判断给定音频流是否为合法音频
func CheckAudio(readSeeker io.ReadSeeker) (string, bool) {
	t := scanType(readSeeker)
	if strings.Contains(t, "audio") {
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

// EncoderSilk 编码为 SILK
func EncoderSilk(data []byte) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to compute md5: %v", err)
	}
	name := hex.EncodeToString(hash.Sum(nil))
	return encode(data, name)
}

// encode 编码为 SILK
func encode(data []byte, name string) (silkWav []byte, err error) {
	// 0. 创建缓存目录
	err = createDirectoryIfNotExist(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio cache directory: %v", err)
	}

	// 1. 创建临时文件
	rawPath := path.Join(cachePath, name+".wav")
	err = os.WriteFile(rawPath, data, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(rawPath)

	// 2. 转换 PCM
	sampleRate := 24000 // 固定采样率，之后可能采取配置或动态决定
	pcmPath := path.Join(cachePath, name+".pcm")
	cmd := exec.Command("ffmpeg", "-i", rawPath, "-f", "s16le", "-ar", strconv.Itoa(sampleRate), "-ac", "1", pcmPath)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert to pcm: %v", err)
	}
	defer os.Remove(pcmPath)

	silkPath := path.Join(cachePath, name+".silk")

	// 3. 转换 SILK
	codecFileName, err := getSilkCodecPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get silk codec path: %v", err)
	}
	codecData, err := silkCodecs.ReadFile(codecFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read silk codec: %v", err)
	}
	filePattern := "silk_codec*"
	if runtime.GOOS == "windows" {
		filePattern += ".exe"
	}
	file, err := os.CreateTemp("", filePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create silk codec temporary file: %v", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(codecData); err != nil {
		return nil, fmt.Errorf("failed to write silk codec temporary file: %v", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("failed to close silk codec temporary file: %v", err)
	}
	if err := os.Chmod(file.Name(), 0700); err != nil {
		return nil, fmt.Errorf("failed to change silk codec temporary file permission: %v", err)
	}
	if runtime.GOOS == "windows" {
		cmd = exec.Command(file.Name(), "-i", pcmPath, "-o", silkPath, "-s", strconv.Itoa(sampleRate))
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to encode silk: %v", err)
		}
	} else {
		cmd = exec.Command(file.Name(), "pts", "-i", pcmPath, "-o", silkPath, "-s", strconv.Itoa(sampleRate))
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to encode silk: %v", err)
		}
	}
	silkWav, err = os.ReadFile(silkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read silk file: %v", err)
	}
	defer os.Remove(silkPath)

	return silkWav, nil
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

// getSilkCodecPath 获取 SILK 编解码器路径
func getSilkCodecPath() (string, error) {
	var codecFileName string
	// 根据 OS 不同获取不同路径
	switch runtime.GOOS {
	case "windows":
		// Windows 下统一使用叶大神编码器
		codecFileName = "silk_codec-windows.exe"
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			codecFileName = "silk_codec-linux-x64"
		case "arm64":
			codecFileName = "silk_codec-linux-arm64"
		default:
			return "", fmt.Errorf("unsupported architecture for Linux: %s", runtime.GOARCH)
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			codecFileName = "silk_codec-macos"
		case "arm64":
			codecFileName = "silk_codec-macos"
		default:
			return "", fmt.Errorf("unsupported architecture for macOS: %s", runtime.GOARCH)
		}
	case "android":
		switch runtime.GOARCH {
		case "arm64":
			codecFileName = "silk_codec-android-arm64"
		case "x86":
			codecFileName = "silk_codec-android-x86"
		case "x86_64":
			codecFileName = "silk_codec-android-x86_64"
		default:
			return "", fmt.Errorf("unsupported architecture for macOS: %s", runtime.GOARCH)
		}
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return "exec/" + codecFileName, nil
}
