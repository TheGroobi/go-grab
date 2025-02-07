package files

func GetFileExtension(s string) string {
	switch s {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/bmp":
		return "bmp"
	case "image/svg+xml":
		return "svg"
	case "image/tiff":
		return "tiff"
	case "image/vnd.microsoft.icon":
		return "ico"

	// Audio
	case "audio/mpeg":
		return "mp3"
	case "audio/wav":
		return "wav"
	case "audio/ogg":
		return "ogg"
	case "audio/webm":
		return "webm"
	case "audio/flac":
		return "flac"

	// Video
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/ogg":
		return "ogv"
	case "video/x-msvideo":
		return "avi"
	case "video/mpeg":
		return "mpeg"

	// Documents
	case "application/pdf":
		return "pdf"
	case "application/msword":
		return "doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case "application/vnd.ms-excel":
		return "xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case "application/vnd.ms-powerpoint":
		return "ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return "pptx"

	// Text & Code
	case "text/plain":
		return "txt"
	case "text/html":
		return "html"
	case "text/css":
		return "css"
	case "text/javascript":
		return "js"
	case "application/json":
		return "json"
	case "application/xml":
		return "xml"
	case "application/x-yaml":
		return "yaml"
	case "application/x-sh":
		return "sh"
	case "application/x-httpd-php":
		return "php"

	// Archives & Executables
	case "application/zip":
		return "zip"
	case "application/x-rar-compressed":
		return "rar"
	case "application/x-7z-compressed":
		return "7z"
	case "application/gzip":
		return "gz"
	case "application/x-tar":
		return "tar"
	case "application/java-archive":
		return "jar"
	case "application/x-msdownload":
		return "exe"
	case "application/x-iso9660-image":
		return "iso"

	default:
		return "bin"
	}
}
