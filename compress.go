package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"golang.org/x/image/tiff"
)

// GetImageInfo 获取图片信息（用于拖放后显示）
func (a *App) GetImageInfo(filePath string) ImageInfo {
	info := ImageInfo{
		Path: filePath,
		Name: filepath.Base(filePath),
	}

	// 获取文件大小
	stat, err := os.Stat(filePath)
	if err != nil {
		return info
	}
	info.Size = stat.Size()

	// 读取并解码图片
	data, err := os.ReadFile(filePath)
	if err != nil {
		return info
	}

	img, format, err := decodeImage(data, filePath)
	if err != nil {
		return info
	}

	bounds := img.Bounds()
	info.Width = bounds.Dx()
	info.Height = bounds.Dy()
	info.Format = format

	// 生成预览缩略图
	previewImg := resize.Thumbnail(200, 200, img, resize.Lanczos3)
	var previewBuf bytes.Buffer
	jpeg.Encode(&previewBuf, previewImg, &jpeg.Options{Quality: 80})
	info.Preview = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(previewBuf.Bytes())

	return info
}

// decodeImage 解码各种格式的图片
func decodeImage(data []byte, filePath string) (image.Image, string, error) {
	reader := bytes.NewReader(data)
	ext := strings.ToLower(filepath.Ext(filePath))

	// 先尝试标准解码
	img, format, err := image.Decode(reader)
	if err == nil {
		return img, format, nil
	}

	// 根据扩展名尝试特定格式
	reader.Seek(0, 0)
	switch ext {
	case ".webp":
		img, err = decodeWebp(data)
		if err == nil {
			return img, "webp", nil
		}
	case ".tiff", ".tif":
		img, err = tiff.Decode(reader)
		if err == nil {
			return img, "tiff", nil
		}
	case ".gif":
		img, err = gif.Decode(reader)
		if err == nil {
			return img, "gif", nil
		}
	}

	return nil, "", fmt.Errorf("无法解码图片: %v", err)
}

// CompressImage 压缩单张图片
func (a *App) CompressImage(inputPath string, options CompressOptions) CompressResult {
	// 读取原始文件
	originalData, err := os.ReadFile(inputPath)
	if err != nil {
		return CompressResult{Success: false, Message: fmt.Sprintf("无法打开文件: %v", err)}
	}
	originalSize := int64(len(originalData))

	// 解码图片
	img, format, err := decodeImage(originalData, inputPath)
	if err != nil {
		return CompressResult{Success: false, Message: fmt.Sprintf("无法解码图片: %v", err)}
	}

	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// 调整尺寸
	resizedImg := img
	if options.MaxWidth > 0 || options.MaxHeight > 0 {
		if options.KeepAspect {
			resizedImg = resize.Thumbnail(options.MaxWidth, options.MaxHeight, img, resize.Lanczos3)
		} else {
			if options.MaxWidth > 0 && options.MaxHeight > 0 {
				resizedImg = resize.Resize(options.MaxWidth, options.MaxHeight, img, resize.Lanczos3)
			} else if options.MaxWidth > 0 {
				resizedImg = resize.Resize(options.MaxWidth, 0, img, resize.Lanczos3)
			} else {
				resizedImg = resize.Resize(0, options.MaxHeight, img, resize.Lanczos3)
			}
		}
	}

	newBounds := resizedImg.Bounds()
	newWidth := newBounds.Dx()
	newHeight := newBounds.Dy()

	// 确定输出格式
	outputFormat := options.OutputFormat
	if outputFormat == "" || outputFormat == "original" {
		// 按文件扩展名决定输出格式，而不是实际格式
		// ext := strings.ToLower(filepath.Ext(inputPath))
		// switch ext {
		// case ".jpg", ".jpeg":
		// 	outputFormat = "jpeg"
		// case ".png":
		// 	outputFormat = "png"
		// case ".webp":
		// 	outputFormat = "webp"
		// case ".gif":
		// 	outputFormat = "gif"
		// default:
		outputFormat = format // 无法识别时使用实际格式
		// }
	}

	// 调试日志
	fmt.Printf("[DEBUG] 输入格式: %s, 选项格式: %s, 输出格式: %s\n", format, options.OutputFormat, outputFormat)

	// 生成输出文件名
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	var outputExt string
	switch outputFormat {
	case "jpeg", "jpg":
		outputExt = ".jpg"
	case "png":
		outputExt = ".png"
	case "webp":
		outputExt = ".webp"
	case "gif":
		outputExt = ".gif"
	default:
		outputExt = ext
	}

	outputPath := filepath.Join(options.OutputDir, nameWithoutExt+outputExt)

	// 压缩图片
	var buf bytes.Buffer
	var mimeType string

	switch outputFormat {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: options.Quality})
		mimeType = "image/jpeg"
	case "png":
		// 使用类似 TinyPNG 的量化压缩
		pngData, _ := compressPNGLikeTinyPNG(resizedImg, options.Quality)
		buf.Write(pngData)
		mimeType = "image/png"
	case "webp":
		err = encodeWebp(&buf, resizedImg, options.Quality)
		mimeType = "image/webp"
	case "gif":
		err = gif.Encode(&buf, resizedImg, nil)
		mimeType = "image/gif"
	default:
		// 默认使用 JPEG
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: options.Quality})
		mimeType = "image/jpeg"
	}

	if err != nil {
		return CompressResult{Success: false, Message: fmt.Sprintf("压缩失败: %v", err)}
	}

	// 智能判断：如果压缩后更大且没有改变尺寸，使用原文件
	compressedData := buf.Bytes()
	newSize := int64(len(compressedData))

	// 检查是否尺寸未变（没有缩放）
	sizeUnchanged := (options.MaxWidth == 0 && options.MaxHeight == 0) ||
		(newWidth == originalWidth && newHeight == originalHeight)

	// 如果格式相同、尺寸未变、且压缩后更大，使用原文件
	sameFormat := (outputFormat == format) ||
		(outputFormat == "jpeg" && format == "jpg") ||
		(outputFormat == "jpg" && format == "jpeg")

	useOriginal := false
	if sameFormat && sizeUnchanged && newSize >= originalSize {
		// 压缩后反而更大，直接复制原文件
		compressedData = originalData
		newSize = originalSize
		useOriginal = true
	}

	// 保存压缩后的文件
	err = os.WriteFile(outputPath, compressedData, 0644)
	if err != nil {
		return CompressResult{Success: false, Message: fmt.Sprintf("保存失败: %v", err)}
	}

	compressionRatio := float64(originalSize-newSize) / float64(originalSize) * 100

	// 生成预览图
	originalBase64 := ""
	compressedBase64 := ""

	// 生成原图预览
	previewImg := resize.Thumbnail(800, 800, img, resize.Lanczos3)
	var previewBuf bytes.Buffer
	jpeg.Encode(&previewBuf, previewImg, &jpeg.Options{Quality: 85})
	originalBase64 = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(previewBuf.Bytes())

	// 生成压缩后预览
	if useOriginal {
		// 使用原文件，预览也用原图
		compressedBase64 = originalBase64
	} else if newSize < 500*1024 { // 小于500KB直接使用
		compressedBase64 = "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(compressedData)
	} else {
		// 大图片生成预览
		compressedPreview := resize.Thumbnail(800, 800, resizedImg, resize.Lanczos3)
		var compressedPreviewBuf bytes.Buffer
		jpeg.Encode(&compressedPreviewBuf, compressedPreview, &jpeg.Options{Quality: 85})
		compressedBase64 = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(compressedPreviewBuf.Bytes())
	}

	message := "压缩成功"
	if useOriginal {
		message = "已保持原文件（压缩后更大）"
	}

	return CompressResult{
		Success:          true,
		Message:          message,
		OriginalSize:     originalSize,
		NewSize:          newSize,
		OutputPath:       outputPath,
		OriginalBase64:   originalBase64,
		CompressedBase64: compressedBase64,
		OriginalWidth:    originalWidth,
		OriginalHeight:   originalHeight,
		NewWidth:         newWidth,
		NewHeight:        newHeight,
		CompressionRatio: compressionRatio,
	}
}

// GetSupportedFormats 获取支持的格式列表
func (a *App) GetSupportedFormats() map[string][]string {
	outputFormats := []string{"jpg", "png"}
	if webpSupported() {
		outputFormats = append(outputFormats, "webp")
	}
	return map[string][]string{
		"input":  {"jpg", "jpeg", "png", "gif", "webp", "tiff", "tif", "bmp"},
		"output": outputFormats,
	}
}
