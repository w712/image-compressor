package main

import "image/color"

// ImageInfo 存储图片的基本信息
type ImageInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Format  string `json:"format"`
	Preview string `json:"preview"`
}

// CompressOptions 压缩选项
type CompressOptions struct {
	Quality      int    `json:"quality"`
	MaxWidth     uint   `json:"maxWidth"`
	MaxHeight    uint   `json:"maxHeight"`
	OutputFormat string `json:"outputFormat"` // "original", "jpeg", "png", "webp"
	OutputDir    string `json:"outputDir"`
	KeepAspect   bool   `json:"keepAspect"`
}

// CompressResult 压缩结果
type CompressResult struct {
	Success          bool    `json:"success"`
	Message          string  `json:"message"`
	OriginalSize     int64   `json:"originalSize"`
	NewSize          int64   `json:"newSize"`
	OutputPath       string  `json:"outputPath"`
	OriginalBase64   string  `json:"originalBase64"`
	CompressedBase64 string  `json:"compressedBase64"`
	OriginalWidth    int     `json:"originalWidth"`
	OriginalHeight   int     `json:"originalHeight"`
	NewWidth         int     `json:"newWidth"`
	NewHeight        int     `json:"newHeight"`
	CompressionRatio float64 `json:"compressionRatio"`
}

// GifOptions GIF 生成选项
type GifOptions struct {
	FrameDelay int    `json:"frameDelay"` // 帧延迟，单位：毫秒
	LoopCount  int    `json:"loopCount"`  // 循环次数，0=无限循环
	MaxWidth   uint   `json:"maxWidth"`   // 最大宽度
	MaxHeight  uint   `json:"maxHeight"`  // 最大高度
	OutputDir  string `json:"outputDir"`  // 输出目录
	OutputName string `json:"outputName"` // 输出文件名（不含扩展名）
}

// GifResult GIF 生成结果
type GifResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OutputPath string `json:"outputPath"`
	FileSize   int64  `json:"fileSize"`
	FrameCount int    `json:"frameCount"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Preview    string `json:"preview"` // Base64 预览
}

// GifCompressOptions GIF 压缩选项
type GifCompressOptions struct {
	MaxWidth  uint   `json:"maxWidth"`  // 最大宽度，0表示不限制
	MaxHeight uint   `json:"maxHeight"` // 最大高度，0表示不限制
	Colors    int    `json:"colors"`    // 颜色数量 2-256，越少文件越小
	Lossy     int    `json:"lossy"`     // 有损压缩级别 0-200，0=无损
	OutputDir string `json:"outputDir"` // 输出目录
}

// colorBox 表示 Median Cut 算法中的颜色盒子
type colorBox struct {
	colors         []color.RGBA
	rMin, rMax     uint8
	gMin, gMax     uint8
	bMin, bMax     uint8
}
