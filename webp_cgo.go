//go:build !windows
// +build !windows

package main

import (
	"bytes"
	"image"

	"github.com/chai2010/webp"
)

// decodeWebp 解码 WebP 图片 (CGO 版本)
func decodeWebp(data []byte) (image.Image, error) {
	return webp.Decode(bytes.NewReader(data))
}

// encodeWebp 编码为 WebP 格式 (CGO 版本)
func encodeWebp(buf *bytes.Buffer, img image.Image, quality int) error {
	return webp.Encode(buf, img, &webp.Options{Quality: float32(quality)})
}

// webpSupported 是否支持 WebP 输出
func webpSupported() bool {
	return true
}
