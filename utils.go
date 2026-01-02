package main

import (
	"fmt"
	"path/filepath"
)

// sortImagePaths 按文件名自然排序
func sortImagePaths(paths []string) []string {
	sorted := make([]string, len(paths))
	copy(sorted, paths)

	// 使用自然排序（处理数字序列）
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if naturalLess(filepath.Base(sorted[j]), filepath.Base(sorted[i])) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

// naturalLess 自然排序比较（数字按数值比较）
func naturalLess(a, b string) bool {
	ai, bi := 0, 0
	for ai < len(a) && bi < len(b) {
		if isDigit(a[ai]) && isDigit(b[bi]) {
			// 提取数字部分
			var numA, numB int
			for ai < len(a) && isDigit(a[ai]) {
				numA = numA*10 + int(a[ai]-'0')
				ai++
			}
			for bi < len(b) && isDigit(b[bi]) {
				numB = numB*10 + int(b[bi]-'0')
				bi++
			}
			if numA != numB {
				return numA < numB
			}
		} else {
			if a[ai] != b[bi] {
				return a[ai] < b[bi]
			}
			ai++
			bi++
		}
	}
	return len(a) < len(b)
}

// isDigit 判断是否为数字字符
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else {
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	}
}
