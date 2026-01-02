package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nfnt/resize"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// CreateGifFromSequence 从序列帧创建 GIF
func (a *App) CreateGifFromSequence(imagePaths []string, options GifOptions) GifResult {
	if len(imagePaths) < 2 {
		return GifResult{Success: false, Message: "至少需要 2 张图片来创建 GIF"}
	}

	// 排序文件路径（按文件名自然排序）
	sortedPaths := sortImagePaths(imagePaths)

	// 读取所有图片
	var frames []image.Image
	var firstWidth, firstHeight int

	for i, path := range sortedPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return GifResult{Success: false, Message: fmt.Sprintf("无法读取文件 %s: %v", filepath.Base(path), err)}
		}

		img, _, err := decodeImage(data, path)
		if err != nil {
			return GifResult{Success: false, Message: fmt.Sprintf("无法解码图片 %s: %v", filepath.Base(path), err)}
		}

		// 记录第一张图的尺寸作为基准
		if i == 0 {
			bounds := img.Bounds()
			firstWidth = bounds.Dx()
			firstHeight = bounds.Dy()
		}

		frames = append(frames, img)
	}

	// 确定输出尺寸
	outWidth := uint(firstWidth)
	outHeight := uint(firstHeight)

	if options.MaxWidth > 0 && outWidth > options.MaxWidth {
		ratio := float64(options.MaxWidth) / float64(outWidth)
		outWidth = options.MaxWidth
		outHeight = uint(float64(outHeight) * ratio)
	}
	if options.MaxHeight > 0 && outHeight > options.MaxHeight {
		ratio := float64(options.MaxHeight) / float64(outHeight)
		outHeight = options.MaxHeight
		outWidth = uint(float64(outWidth) * ratio)
	}

	// 转换帧延迟：毫秒 -> 1/100秒
	delay := options.FrameDelay / 10
	if delay < 1 {
		delay = 10 // 默认 100ms
	}

	// 创建 GIF 结构
	gifImg := &gif.GIF{
		LoopCount: options.LoopCount,
	}

	// 生成全局调色板（使用第一帧）
	palette := generatePalette(frames[0])

	// 处理每一帧
	for _, frame := range frames {
		// 调整尺寸
		resizedFrame := resize.Resize(outWidth, outHeight, frame, resize.Lanczos3)

		// 转换为调色板图像
		bounds := resizedFrame.Bounds()
		palettedImg := image.NewPaletted(bounds, palette)

		// 使用 Floyd-Steinberg 抖动算法进行高质量颜色量化
		draw.FloydSteinberg.Draw(palettedImg, bounds, resizedFrame, image.Point{})

		gifImg.Image = append(gifImg.Image, palettedImg)
		gifImg.Delay = append(gifImg.Delay, delay)
		// 设置处置方法：每帧播放后清除为背景色，防止残影
		gifImg.Disposal = append(gifImg.Disposal, gif.DisposalBackground)
	}

	// 设置 GIF 配置
	if len(gifImg.Image) > 0 {
		gifImg.Config = image.Config{
			Width:      int(outWidth),
			Height:     int(outHeight),
			ColorModel: palette,
		}
	}

	// 生成输出路径
	outputName := options.OutputName
	if outputName == "" {
		outputName = "animation"
	}
	outputPath := filepath.Join(options.OutputDir, outputName+".gif")

	// 编码并保存
	var buf bytes.Buffer
	err := gif.EncodeAll(&buf, gifImg)
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("GIF 编码失败: %v", err)}
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("保存失败: %v", err)}
	}

	// 生成预览（小尺寸的 GIF base64）
	previewBase64 := ""
	if buf.Len() < 2*1024*1024 { // 小于 2MB 直接使用
		previewBase64 = "data:image/gif;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	} else {
		// 大文件生成缩略预览
		previewGif := createPreviewGif(gifImg, 200)
		var previewBuf bytes.Buffer
		gif.EncodeAll(&previewBuf, previewGif)
		previewBase64 = "data:image/gif;base64," + base64.StdEncoding.EncodeToString(previewBuf.Bytes())
	}

	return GifResult{
		Success:    true,
		Message:    "GIF 创建成功",
		OutputPath: outputPath,
		FileSize:   int64(buf.Len()),
		FrameCount: len(frames),
		Width:      int(outWidth),
		Height:     int(outHeight),
		Preview:    previewBase64,
	}
}

// CompressGif 压缩 GIF 文件（带进度回调）
func (a *App) CompressGif(gifPath string, options GifCompressOptions) GifResult {
	// 读取 GIF 文件
	data, err := os.ReadFile(gifPath)
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("无法读取文件: %v", err)}
	}

	originalSize := int64(len(data))

	// 发送进度：解码中
	runtime.EventsEmit(a.ctx, "gif-compress-progress", map[string]interface{}{
		"stage":    "decoding",
		"progress": 0,
		"message":  "正在解码 GIF...",
	})

	// 解码 GIF
	gifImg, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("无法解码 GIF: %v", err)}
	}

	if len(gifImg.Image) == 0 {
		return GifResult{Success: false, Message: "GIF 文件没有帧"}
	}

	totalFrames := len(gifImg.Image)
	origWidth := gifImg.Config.Width
	origHeight := gifImg.Config.Height

	// 计算新尺寸
	newWidth := uint(origWidth)
	newHeight := uint(origHeight)
	needResize := false

	if options.MaxWidth > 0 && newWidth > options.MaxWidth {
		ratio := float64(options.MaxWidth) / float64(newWidth)
		newWidth = options.MaxWidth
		newHeight = uint(float64(newHeight) * ratio)
		needResize = true
	}
	if options.MaxHeight > 0 && newHeight > options.MaxHeight {
		ratio := float64(options.MaxHeight) / float64(newHeight)
		newHeight = options.MaxHeight
		newWidth = uint(float64(newWidth) * ratio)
		needResize = true
	}

	// 颜色数量限制 (2-256)
	colors := options.Colors
	if colors < 2 {
		colors = 256
	}
	if colors > 256 {
		colors = 256
	}

	// 发送进度：生成调色板
	runtime.EventsEmit(a.ctx, "gif-compress-progress", map[string]interface{}{
		"stage":    "palette",
		"progress": 5,
		"message":  "正在生成调色板...",
	})

	// 生成优化的调色板（使用快速版本）
	palette := generateFastPalette(gifImg.Image[0], colors)

	// 创建新的 GIF
	newGif := &gif.GIF{
		LoopCount: gifImg.LoopCount,
		Config: image.Config{
			Width:      int(newWidth),
			Height:     int(newHeight),
			ColorModel: palette,
		},
	}

	// 处理每一帧
	for i, frame := range gifImg.Image {
		// 发送进度
		progress := 10 + (i * 80 / totalFrames)
		runtime.EventsEmit(a.ctx, "gif-compress-progress", map[string]interface{}{
			"stage":    "processing",
			"progress": progress,
			"message":  fmt.Sprintf("正在处理帧 %d/%d...", i+1, totalFrames),
		})

		var processedFrame image.Image = frame

		// 如果需要缩放，使用更快的算法
		if needResize {
			processedFrame = resize.Resize(newWidth, newHeight, frame, resize.NearestNeighbor)
		}

		// 转换为调色板图像（使用快速绘制）
		bounds := processedFrame.Bounds()
		palettedImg := image.NewPaletted(bounds, palette)

		// 使用简单绘制而不是 Floyd-Steinberg（更快）
		draw.Draw(palettedImg, bounds, processedFrame, image.Point{}, draw.Src)

		newGif.Image = append(newGif.Image, palettedImg)
		newGif.Delay = append(newGif.Delay, gifImg.Delay[i])
		newGif.Disposal = append(newGif.Disposal, gif.DisposalBackground)
	}

	// 发送进度：编码中
	runtime.EventsEmit(a.ctx, "gif-compress-progress", map[string]interface{}{
		"stage":    "encoding",
		"progress": 90,
		"message":  "正在编码 GIF...",
	})

	// 生成输出路径
	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(gifPath)
	}

	baseName := strings.TrimSuffix(filepath.Base(gifPath), filepath.Ext(gifPath))
	outputPath := filepath.Join(outputDir, baseName+"_compressed.gif")

	// 编码并保存
	var buf bytes.Buffer
	err = gif.EncodeAll(&buf, newGif)
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("GIF 编码失败: %v", err)}
	}

	err = os.WriteFile(outputPath, buf.Bytes(), 0644)
	if err != nil {
		return GifResult{Success: false, Message: fmt.Sprintf("保存失败: %v", err)}
	}

	newSize := int64(buf.Len())

	// 发送进度：完成
	runtime.EventsEmit(a.ctx, "gif-compress-progress", map[string]interface{}{
		"stage":    "done",
		"progress": 100,
		"message":  "压缩完成!",
	})

	// 生成预览
	previewBase64 := ""
	if buf.Len() < 2*1024*1024 {
		previewBase64 = "data:image/gif;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	return GifResult{
		Success:    true,
		Message:    fmt.Sprintf("压缩完成！原始: %s → 压缩后: %s (节省 %.1f%%)", formatFileSize(originalSize), formatFileSize(newSize), float64(originalSize-newSize)/float64(originalSize)*100),
		OutputPath: outputPath,
		FileSize:   newSize,
		FrameCount: len(newGif.Image),
		Width:      int(newWidth),
		Height:     int(newHeight),
		Preview:    previewBase64,
	}
}

// generatePalette 从图像生成 256 色调色板
func generatePalette(img image.Image) color.Palette {
	bounds := img.Bounds()
	colorMap := make(map[uint32]int)

	// 采样图像颜色
	step := 1
	if bounds.Dx() > 100 || bounds.Dy() > 100 {
		step = 2
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			// 量化到较少的颜色级别
			key := ((r >> 11) << 10) | ((g >> 11) << 5) | (b >> 11)
			colorMap[key]++
		}
	}

	// 选择最常见的颜色
	type colorCount struct {
		key   uint32
		count int
	}
	var colors []colorCount
	for k, v := range colorMap {
		colors = append(colors, colorCount{k, v})
	}

	// 按出现频率排序
	for i := 0; i < len(colors)-1; i++ {
		for j := i + 1; j < len(colors); j++ {
			if colors[j].count > colors[i].count {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}

	// 生成调色板
	palette := make(color.Palette, 0, 256)

	// 添加常用颜色
	for i := 0; i < len(colors) && len(palette) < 255; i++ {
		key := colors[i].key
		r := uint8((key >> 10) << 3)
		g := uint8(((key >> 5) & 0x1f) << 3)
		b := uint8((key & 0x1f) << 3)
		palette = append(palette, color.RGBA{r, g, b, 255})
	}

	// 如果颜色不够，添加灰度
	for i := 0; len(palette) < 256; i += 256 / (256 - len(palette) + 1) {
		palette = append(palette, color.Gray{uint8(i)})
	}

	// 确保有透明色
	if len(palette) == 256 {
		palette[255] = color.RGBA{0, 0, 0, 0}
	} else {
		palette = append(palette, color.RGBA{0, 0, 0, 0})
	}

	return palette
}

// generateFastPalette 快速生成调色板（牺牲一点质量换取速度）
func generateFastPalette(img image.Image, maxColors int) color.Palette {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 大幅采样：最多采样 10000 个像素
	step := 1
	totalPixels := width * height
	if totalPixels > 10000 {
		step = int(math.Sqrt(float64(totalPixels) / 10000))
		if step < 1 {
			step = 1
		}
	}

	// 使用量化的颜色桶来加速
	colorBuckets := make(map[uint32]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			// 量化到 5 位（32 级）以减少颜色数量
			key := ((r >> 11) << 10) | ((g >> 11) << 5) | (b >> 11)
			colorBuckets[key]++
		}
	}

	// 转换并排序
	type bucketCount struct {
		key   uint32
		count int
	}
	buckets := make([]bucketCount, 0, len(colorBuckets))
	for k, v := range colorBuckets {
		buckets = append(buckets, bucketCount{k, v})
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].count > buckets[j].count
	})

	// 生成调色板
	palette := make(color.Palette, 0, maxColors)
	for i := 0; i < len(buckets) && len(palette) < maxColors-1; i++ {
		key := buckets[i].key
		r := uint8((key >> 10) << 3)
		g := uint8(((key >> 5) & 0x1f) << 3)
		b := uint8((key & 0x1f) << 3)
		palette = append(palette, color.RGBA{r, g, b, 255})
	}

	// 补充灰度色
	for len(palette) < maxColors-1 {
		g := uint8(len(palette) * 255 / maxColors)
		palette = append(palette, color.Gray{g})
	}

	// 添加透明色
	palette = append(palette, color.RGBA{0, 0, 0, 0})

	return palette
}

// createPreviewGif 创建预览用的小尺寸 GIF
func createPreviewGif(original *gif.GIF, maxSize uint) *gif.GIF {
	if len(original.Image) == 0 {
		return original
	}

	// 计算缩放比例
	origWidth := original.Image[0].Bounds().Dx()
	origHeight := original.Image[0].Bounds().Dy()

	scale := float64(1)
	if uint(origWidth) > maxSize || uint(origHeight) > maxSize {
		scaleW := float64(maxSize) / float64(origWidth)
		scaleH := float64(maxSize) / float64(origHeight)
		if scaleW < scaleH {
			scale = scaleW
		} else {
			scale = scaleH
		}
	}

	newWidth := int(float64(origWidth) * scale)
	newHeight := int(float64(origHeight) * scale)

	preview := &gif.GIF{
		LoopCount: original.LoopCount,
		Delay:     original.Delay,
		Disposal:  original.Disposal,
		Config: image.Config{
			Width:  newWidth,
			Height: newHeight,
		},
	}

	for _, frame := range original.Image {
		resized := resize.Resize(uint(newWidth), uint(newHeight), frame, resize.NearestNeighbor)
		bounds := resized.Bounds()
		paletted := image.NewPaletted(bounds, frame.Palette)
		draw.FloydSteinberg.Draw(paletted, bounds, resized, image.Point{})
		preview.Image = append(preview.Image, paletted)
	}

	return preview
}
