//go:build windows
// +build windows

package main

import (
	"bytes"
	"fmt"
	"image"

	"golang.org/x/image/webp"
)

// decodeWebp 解码 WebP 图片 (纯 Go 版本)
func decodeWebp(data []byte) (image.Image, error) {
	return webp.Decode(bytes.NewReader(data))
}

// encodeWebp 编码为 WebP 格式 (Windows 不支持)
func encodeWebp(buf *bytes.Buffer, img image.Image, quality int) error {
	return fmt.Errorf("WebP 编码在 Windows 版本暂不支持，请选择其他格式")
}

// webpSupported 是否支持 WebP 输出
func webpSupported() bool {
	return false
}
