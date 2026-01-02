package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 应用程序结构
type App struct {
	ctx context.Context
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{}
}

// startup 应用启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectImages 选择图片文件
func (a *App) SelectImages() []string {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择图片",
		Filters: []runtime.FileFilter{
			{DisplayName: "图片文件", Pattern: "*.jpg;*.jpeg;*.png;*.gif;*.webp;*.tiff;*.tif;*.bmp"},
		},
	})
	if err != nil {
		return []string{}
	}
	return files
}

// SelectOutputDir 选择输出目录
func (a *App) SelectOutputDir() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择输出目录",
	})
	if err != nil {
		return ""
	}
	return dir
}
