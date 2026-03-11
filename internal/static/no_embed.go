//go:build !web

package static

import (
	"io/fs"
	"os"
	"path/filepath"
)

// DefaultStaticDir 是 Docker 模式下默认的静态资源目录
const DefaultStaticDir = "/www/baihu"

func GetFS() fs.FS {
	dir := getStaticDir()
	if dir == "" {
		return nil
	}
	return os.DirFS(dir)
}

func ReadFile(name string) ([]byte, error) {
	dir := getStaticDir()
	if dir == "" {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(filepath.Join(dir, name))
}

func getStaticDir() string {
	// 优先从环境变量读取
	if dir := os.Getenv("BH_STATIC_DIR"); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}
	// 回退到默认目录
	if _, err := os.Stat(DefaultStaticDir); err == nil {
		return DefaultStaticDir
	}
	return ""
}
