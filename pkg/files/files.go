package files

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	DownloadDirNames []string = []string{"Downloads", "downloads", "download", "Downloads", "Pobrane"}
	MimeToExt                 = map[string]string{
		// Images
		"image/jpeg":               "jpg",
		"image/png":                "png",
		"image/gif":                "gif",
		"image/webp":               "webp",
		"image/bmp":                "bmp",
		"image/svg+xml":            "svg",
		"image/tiff":               "tiff",
		"image/vnd.microsoft.icon": "ico",

		// Audio
		"audio/mpeg": "mp3",
		"audio/wav":  "wav",
		"audio/ogg":  "ogg",
		"audio/webm": "webm",
		"audio/flac": "flac",

		// Video
		"video/mp4":       "mp4",
		"video/x-m4v":     "m4v",
		"video/webm":      "webm",
		"video/ogg":       "ogv",
		"video/x-msvideo": "avi",
		"video/mpeg":      "mpeg",

		// Documents
		"application/pdf":    "pdf",
		"application/msword": "doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
		"application/vnd.ms-excel": "xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "xlsx",
		"application/vnd.ms-powerpoint":                                             "ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",

		// Text & Code
		"text/plain":              "txt",
		"text/html":               "html",
		"text/css":                "css",
		"text/javascript":         "js",
		"application/json":        "json",
		"application/xml":         "xml",
		"application/x-yaml":      "yaml",
		"application/x-sh":        "sh",
		"application/x-httpd-php": "php",

		// Archives & Executables
		"application/zip":              "zip",
		"application/x-rar-compressed": "rar",
		"application/x-7z-compressed":  "7z",
		"application/gzip":             "gz",
		"application/x-tar":            "tar",
		"application/java-archive":     "jar",
		"application/x-msdownload":     "exe",
		"application/x-iso9660-image":  "iso",
	}
)

func GetFileExtension(mimeType string) string {
	if ext, exists := MimeToExt[mimeType]; exists {
		return ext
	}
	return "bin" // Default if MIME type not found
}

func GetDownloadsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}

	var baseDir string
	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("USERPROFILE")
	} else {
		baseDir = homeDir
	}

	for _, ddn := range DownloadDirNames {
		dir := filepath.Join(baseDir, ddn)
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			return dir
		}
	}

	return homeDir
}
