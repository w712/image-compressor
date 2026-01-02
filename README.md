# Squash - 图片压缩工具

一款简单高效的桌面图片压缩工具，基于 Wails 框架开发，支持多种图片格式的压缩和 GIF 动图生成。

![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-blue)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![Wails](https://img.shields.io/badge/Wails-v2-red)

## 功能特性

### 图片压缩
- 支持 JPG、PNG、GIF、WebP、TIFF、BMP 等主流格式
- 可调节压缩质量（1-100%）
- 支持设置最大宽高限制，自动等比缩放
- 支持格式转换（原格式 / JPEG / PNG / WebP）
- 智能压缩：如果压缩后文件更大，自动保留原文件
- 批量处理：支持同时压缩多张图片
- 实时预览：压缩完成后可对比原图与压缩后效果

### GIF 动图生成
- 从序列帧图片生成 GIF 动图
- 支持自定义帧率（帧延迟 10-5000ms）
- 支持设置循环次数（无限循环 / 播放一次 / 自定义次数）
- 支持设置输出尺寸限制
- 自动按文件名排序（支持数字自然排序）

### 用户体验
- 拖放支持：直接拖放图片到窗口
- 深色主题界面
- 实时进度显示
- 压缩前后对比滑块

## 截图

（待添加）

## 安装

### 从源码构建

#### 前置要求
- Go 1.21+
- Node.js 16+
- Wails CLI

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆项目
git clone https://github.com/w712/image-compressor.git
cd image-compressor

# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式运行
wails dev

# 构建生产版本
wails build
```

构建完成后，可执行文件位于 `build/bin/` 目录。

### 平台支持

| 平台 | 状态 | 说明 |
|------|------|------|
| macOS | ✅ | 完整支持 |
| Windows | ✅ | 完整支持 |
| Linux | ✅ | 完整支持 |

## 使用说明

### 图片压缩

1. 点击「添加图片」按钮或直接拖放图片到窗口
2. 点击「输出目录」选择压缩后文件的保存位置
3. 调整压缩参数：
   - 输出格式：保持原格式或转换为 JPEG/PNG/WebP
   - 压缩质量：1-100%，推荐 80%
   - 尺寸限制：可选设置最大宽高
4. 点击「开始压缩」
5. 压缩完成后，点击图片可查看对比效果

### GIF 生成

1. 切换到「GIF」模式
2. 拖放序列帧图片（至少 2 张）
3. 设置参数：
   - 帧延迟：控制播放速度
   - 循环次数：无限循环或指定次数
   - 输出尺寸：可选限制最大宽高
   - 文件名：设置输出文件名
4. 点击「生成 GIF」

## 技术栈

- 后端：Go + [Wails v2](https://wails.io/)
- 前端：原生 JavaScript + CSS
- 图片处理：
  - [nfnt/resize](https://github.com/nfnt/resize) - 图片缩放
  - [golang.org/x/image](https://pkg.go.dev/golang.org/x/image) - TIFF/WebP 支持

## 项目结构

```
image-compressor/
├── main.go           # 应用入口
├── app.go            # 应用生命周期
├── compress.go       # 图片压缩核心逻辑
├── gif.go            # GIF 生成与压缩
├── quantize.go       # 颜色量化算法（PNG 压缩）
├── types.go          # 数据类型定义
├── utils.go          # 工具函数
├── webp_cgo.go       # WebP 编解码（macOS/Linux）
├── webp_windows.go   # WebP 编解码（Windows）
├── frontend/         # 前端代码
│   ├── src/
│   │   ├── main.js   # 前端主逻辑
│   │   └── style.css # 样式
│   └── wailsjs/      # Wails 生成的绑定
├── build/            # 构建配置与资源
└── wails.json        # Wails 项目配置
```

## 支持的格式

| 格式 | 输入 | 输出 | 说明 |
|------|------|------|------|
| JPEG | ✅ | ✅ | 有损压缩，质量可调 |
| PNG | ✅ | ✅ | 使用量化算法压缩 |
| WebP | ✅ | ✅ | 需要 libwebp 支持 |
| GIF | ✅ | ❌ | 仅支持读取，输出请使用 GIF 模式 |
| TIFF | ✅ | ❌ | 仅支持读取 |
| BMP | ✅ | ❌ | 仅支持读取 |

## 开发

```bash
# 开发模式（热重载）
wails dev

# 构建
wails build

# 构建 Windows 版本（需要交叉编译环境）
wails build -platform windows/amd64
```

## 许可证

MIT License

## 致谢

- [Wails](https://wails.io/) - 优秀的 Go 桌面应用框架
- [nfnt/resize](https://github.com/nfnt/resize) - 高质量图片缩放库
