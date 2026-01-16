// Package env 提供读取环境变量的快捷方法，不再关心 .env 文件。
package utils

import "os"

// Get 读取环境变量，若为空则返回默认值。
func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return ""
}
