package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"sort"
)

// ============================================================
// TinyPNG 风格的 PNG 量化压缩实现
// 使用 Median Cut 算法 + Floyd-Steinberg 抖动
// ============================================================

// newColorBox 创建一个新的颜色盒子
func newColorBox(colors []color.RGBA) *colorBox {
	box := &colorBox{
		colors: colors,
		rMin:   255, rMax: 0,
		gMin: 255, gMax: 0,
		bMin: 255, bMax: 0,
	}
	for _, c := range colors {
		if c.R < box.rMin {
			box.rMin = c.R
		}
		if c.R > box.rMax {
			box.rMax = c.R
		}
		if c.G < box.gMin {
			box.gMin = c.G
		}
		if c.G > box.gMax {
			box.gMax = c.G
		}
		if c.B < box.bMin {
			box.bMin = c.B
		}
		if c.B > box.bMax {
			box.bMax = c.B
		}
	}
	return box
}

// longestAxis 返回颜色范围最大的通道 (0=R, 1=G, 2=B)
func (box *colorBox) longestAxis() int {
	rRange := int(box.rMax) - int(box.rMin)
	gRange := int(box.gMax) - int(box.gMin)
	bRange := int(box.bMax) - int(box.bMin)

	if rRange >= gRange && rRange >= bRange {
		return 0
	}
	if gRange >= rRange && gRange >= bRange {
		return 1
	}
	return 2
}

// averageColor 计算盒子中所有颜色的平均值
func (box *colorBox) averageColor() color.RGBA {
	if len(box.colors) == 0 {
		return color.RGBA{0, 0, 0, 255}
	}

	var rSum, gSum, bSum, aSum int64
	for _, c := range box.colors {
		rSum += int64(c.R)
		gSum += int64(c.G)
		bSum += int64(c.B)
		aSum += int64(c.A)
	}

	n := int64(len(box.colors))
	return color.RGBA{
		R: uint8(rSum / n),
		G: uint8(gSum / n),
		B: uint8(bSum / n),
		A: uint8(aSum / n),
	}
}

// split 沿最长轴分割盒子
func (box *colorBox) split() (*colorBox, *colorBox) {
	if len(box.colors) < 2 {
		return box, nil
	}

	axis := box.longestAxis()

	// 按指定轴排序
	sort.Slice(box.colors, func(i, j int) bool {
		switch axis {
		case 0:
			return box.colors[i].R < box.colors[j].R
		case 1:
			return box.colors[i].G < box.colors[j].G
		default:
			return box.colors[i].B < box.colors[j].B
		}
	})

	// 在中位数位置分割
	mid := len(box.colors) / 2
	return newColorBox(box.colors[:mid]), newColorBox(box.colors[mid:])
}

// medianCutQuantize 使用 Median Cut 算法量化颜色
// 返回指定数量的调色板颜色
func medianCutQuantize(img image.Image, numColors int) color.Palette {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 采样颜色（对大图像进行采样以提高性能）
	maxSamples := 100000
	totalPixels := width * height
	step := 1
	if totalPixels > maxSamples {
		step = int(math.Sqrt(float64(totalPixels) / float64(maxSamples)))
		if step < 1 {
			step = 1
		}
	}

	// 收集颜色并记录频率
	colorFreq := make(map[color.RGBA]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			}
			colorFreq[c]++
		}
	}

	// 转换为颜色列表（按频率展开，常见颜色权重更大）
	var colors []color.RGBA
	for c, freq := range colorFreq {
		// 根据频率添加颜色（最多添加 sqrt(freq) 次以避免过度偏重）
		count := int(math.Sqrt(float64(freq)))
		if count < 1 {
			count = 1
		}
		if count > 100 {
			count = 100
		}
		for i := 0; i < count; i++ {
			colors = append(colors, c)
		}
	}

	if len(colors) == 0 {
		// 返回默认调色板
		return generateDefaultPalette(numColors)
	}

	// 初始化盒子列表
	boxes := []*colorBox{newColorBox(colors)}

	// 分割直到达到目标颜色数量
	for len(boxes) < numColors {
		// 找到颜色数量最多的盒子
		maxIdx := 0
		maxLen := 0
		for i, box := range boxes {
			if len(box.colors) > maxLen {
				maxLen = len(box.colors)
				maxIdx = i
			}
		}

		if maxLen < 2 {
			break
		}

		// 分割该盒子
		box := boxes[maxIdx]
		box1, box2 := box.split()

		if box2 == nil {
			break
		}

		// 替换原盒子为两个新盒子
		boxes[maxIdx] = box1
		boxes = append(boxes, box2)
	}

	// 生成调色板
	palette := make(color.Palette, 0, numColors)
	for _, box := range boxes {
		palette = append(palette, box.averageColor())
	}

	// 如果颜色不够，补充灰度
	for len(palette) < numColors {
		g := uint8(len(palette) * 255 / numColors)
		palette = append(palette, color.RGBA{g, g, g, 255})
	}

	return palette
}

// generateDefaultPalette 生成默认调色板
func generateDefaultPalette(numColors int) color.Palette {
	palette := make(color.Palette, numColors)
	for i := 0; i < numColors; i++ {
		g := uint8(i * 255 / (numColors - 1))
		palette[i] = color.RGBA{g, g, g, 255}
	}
	return palette
}

// quantizePNG 将图像量化为索引色 PNG（类似 TinyPNG）
// quality: 1-100，控制颜色数量 (1=最少颜色/最小文件, 100=256色/最高质量)
// dither: 是否使用 Floyd-Steinberg 抖动
func quantizePNG(img image.Image, quality int, dither bool) *image.Paletted {
	// 根据质量计算颜色数量
	// quality 1-100 映射到 16-256 色
	numColors := 16 + (quality * 240 / 100)
	if numColors > 256 {
		numColors = 256
	}
	if numColors < 8 {
		numColors = 8
	}

	bounds := img.Bounds()

	// 检测图像是否包含透明像素
	hasTransparency := false
	transparentPixels := make(map[int]bool) // 记录透明像素位置
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 65535/2 { // alpha < 50% 视为透明
				hasTransparency = true
				idx := (y-bounds.Min.Y)*bounds.Dx() + (x - bounds.Min.X)
				transparentPixels[idx] = true
			}
		}
	}

	// 使用 Median Cut 生成调色板
	palette := medianCutQuantize(img, numColors)

	// 如果有透明像素，确保调色板包含透明色
	if hasTransparency {
		// 将第一个颜色替换为透明色
		transparentColor := color.RGBA{0, 0, 0, 0}
		// 检查是否已有透明色
		hasTransparentInPalette := false
		for i, c := range palette {
			if rgba, ok := c.(color.RGBA); ok && rgba.A == 0 {
				hasTransparentInPalette = true
				// 将透明色移到第一位
				if i != 0 {
					palette[0], palette[i] = palette[i], palette[0]
				}
				break
			}
		}
		if !hasTransparentInPalette {
			palette[0] = transparentColor
		}
	}

	// 创建调色板图像
	palettedImg := image.NewPaletted(bounds, palette)

	// 应用颜色量化
	if dither {
		// 使用 Floyd-Steinberg 抖动（更高质量，减少色带）
		draw.FloydSteinberg.Draw(palettedImg, bounds, img, image.Point{})
	} else {
		// 直接映射（更快，但可能有色带）
		draw.Draw(palettedImg, bounds, img, image.Point{}, draw.Src)
	}

	// 恢复透明像素
	if hasTransparency {
		transparentIndex := uint8(0) // 透明色在调色板中的索���
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				idx := (y-bounds.Min.Y)*bounds.Dx() + (x - bounds.Min.X)
				if transparentPixels[idx] {
					palettedImg.SetColorIndex(x, y, transparentIndex)
				}
			}
		}
	}

	return palettedImg
}

// compressPNGLikeTinyPNG 使用类似 TinyPNG 的方式压缩 PNG
// 返回压缩后的字节和是否使用了量化
func compressPNGLikeTinyPNG(img image.Image, quality int) ([]byte, bool) {
	// 检查原图是否已经是低色图像
	uniqueColors := countUniqueColors(img, 1000) // 采样检测

	var buf bytes.Buffer

	// 如果颜色数量很少（<= 256），可能已经是索引色图像
	// 或者质量设置很高，优先保持质量
	if uniqueColors <= 256 && quality >= 95 {
		// 使用无损压缩
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		encoder.Encode(&buf, img)
		return buf.Bytes(), false
	}

	// 使用量化压缩
	palettedImg := quantizePNG(img, quality, true)

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	encoder.Encode(&buf, palettedImg)

	// 如果量化后反而更大（极少数情况），回退到原始压缩
	var origBuf bytes.Buffer
	origEncoder := png.Encoder{CompressionLevel: png.BestCompression}
	origEncoder.Encode(&origBuf, img)

	if buf.Len() > origBuf.Len() {
		return origBuf.Bytes(), false
	}

	return buf.Bytes(), true
}

// countUniqueColors 计算图像中的唯一颜色数量（采样）
func countUniqueColors(img image.Image, maxSamples int) int {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	totalPixels := width * height
	step := 1
	if totalPixels > maxSamples {
		step = totalPixels / maxSamples
	}

	colors := make(map[uint32]struct{})
	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if i%step == 0 {
				r, g, b, _ := img.At(x, y).RGBA()
				key := (r>>8)<<16 | (g>>8)<<8 | (b >> 8)
				colors[key] = struct{}{}
			}
			i++
		}
	}

	return len(colors)
}
