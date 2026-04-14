package router

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/engigu/baihu-panel/internal/static"

	"github.com/gin-gonic/gin"
)

// cacheControl 返回设置 Cache-Control header 的中间件
func cacheControl(value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", value)
		c.Next()
	}
}

func initStaticRoutes(root *gin.RouterGroup) {
	staticFS := static.GetFS()
	if staticFS == nil {
		return
	}

	// 专门处理 /assets 目录下的资源
	root.GET("/assets/*filepath", cacheControl("public, max-age=31536000, immutable"), func(ctx *gin.Context) {
		fullPath := "assets" + ctx.Param("filepath")
		fullPath = strings.TrimPrefix(fullPath, "/")

		isGzipSupported := strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip")
		gzPath := fullPath + ".gz"

		// 确定 MIME 类型
		ext := filepath.Ext(fullPath)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			switch ext {
			case ".js": contentType = "application/javascript"
			case ".css": contentType = "text/css"
			case ".svg": contentType = "image/svg+xml"
			default: contentType = "application/octet-stream"
			}
		}

		// 优先尝试读取 .gz 文件
		if gzFile, err := staticFS.Open(gzPath); err == nil {
			defer gzFile.Close()
			ctx.Header("Content-Type", contentType)
			
			if isGzipSupported {
				// 极致性能：流式透传压缩包 (RSS 占用极低)
				ctx.Header("Content-Encoding", "gzip")
				ctx.Status(http.StatusOK)
				io.Copy(ctx.Writer, gzFile)
			} else {
				// 兼容处理：流式解压发送
				gr, _ := gzip.NewReader(gzFile)
				defer gr.Close()
				ctx.Status(http.StatusOK)
				io.Copy(ctx.Writer, gr)
			}
			return
		}

		// 如果没有 .gz，流式读取原文件
		if file, err := staticFS.Open(fullPath); err == nil {
			defer file.Close()
			ctx.Header("Content-Type", contentType)
			ctx.Status(http.StatusOK)
			io.Copy(ctx.Writer, file)
			return
		}

		ctx.Status(404)
	})

	// logo.svg 等单文件处理
	root.GET("/logo.svg", func(ctx *gin.Context) {
		serveSingleFile(ctx, "logo.svg", "image/svg+xml", "public, max-age=86400")
	})

	// PWA 相关路由处理
	initPWARoutes(root)
}

func initPWARoutes(root *gin.RouterGroup) {
	// PWA 相关文件处理
	pwaRootFiles := map[string]string{
		"/sw.js":                "application/javascript",
		"/registerSW.js":        "application/javascript",
		"/manifest.webmanifest": "application/manifest+json",
		"/favicon.ico":          "image/x-icon",
		"/pwa-icon-192.png":     "image/png",
		"/pwa-icon-512.png":     "image/png",
	}

	for path, contentType := range pwaRootFiles {
		pPath := path
		pType := contentType
		root.GET(pPath, func(ctx *gin.Context) {
			file := strings.TrimPrefix(pPath, "/")
			serveSingleFile(ctx, file, pType, "public, no-cache")
		})
	}

	// 动态匹配 workbox-*.js (Vite PWA 生成的库文件)
	root.GET("/workbox-:hash.js", func(ctx *gin.Context) {
		file := "workbox-" + ctx.Param("hash") + ".js"
		serveSingleFile(ctx, file, "application/javascript", "public, max-age=31536000, immutable")
	})
}

func serveSingleFile(ctx *gin.Context, filename string, contentType string, cache string) {
	staticFS := static.GetFS()
	if staticFS == nil {
		ctx.Status(404)
		return
	}

	if cache != "" { ctx.Header("Cache-Control", cache) }
	ctx.Header("Content-Type", contentType)

	isGzipSupported := strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip")

	// 尝试流式发送压缩版
	if gzFile, err := staticFS.Open(filename + ".gz"); err == nil {
		defer gzFile.Close()
		if isGzipSupported {
			ctx.Header("Content-Encoding", "gzip")
			ctx.Status(200)
			io.Copy(ctx.Writer, gzFile)
		} else {
			gr, _ := gzip.NewReader(gzFile)
			defer gr.Close()
			ctx.Status(200)
			io.Copy(ctx.Writer, gr)
		}
		return
	}

	// 尝试流式发送原版
	if file, err := staticFS.Open(filename); err == nil {
		defer file.Close()
		ctx.Status(200)
		io.Copy(ctx.Writer, file)
		return
	}

	ctx.Status(404)
}

// serveSPA 注入配置并返回 index.html 给前端渲染
func serveSPA(ctx *gin.Context, urlPrefix string, status int) {
	staticFS := static.GetFS()
	if staticFS == nil {
		ctx.String(status, "Frontend assets not found.")
		return
	}

	var data []byte
	// index.html 较小且需要修改字符串，可以一次性读入内存
	if gzFile, err := staticFS.Open("index.html.gz"); err == nil {
		defer gzFile.Close()
		gr, _ := gzip.NewReader(gzFile)
		data, _ = io.ReadAll(gr)
		gr.Close()
	} else if file, err := staticFS.Open("index.html"); err == nil {
		defer file.Close()
		data, _ = io.ReadAll(file)
	}

	if data == nil {
		ctx.String(status, "index.html not found.")
		return
	}

	html := string(data)
	baseHref := urlPrefix + "/"
	if urlPrefix == "" { baseHref = "/" }
	html = strings.Replace(html, "<head>", "<head>\n    <base href=\""+baseHref+"\">", 1)
	configScript := `<script>window.__BASE_URL__ = "` + urlPrefix + `"; window.__API_VERSION__ = "/api/v1";</script>`
	html = strings.Replace(html, "</head>", configScript+"</head>", 1)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Data(status, "text/html; charset=utf-8", []byte(html))
}
