package render

import "github.com/crumbyte/noxdir/structure"

// EntryIcon resolves an emoji icon for the provided Entry instance based on the
// file extension.
//
//nolint:cyclop,funlen // speed and simplicity over another map resolver
func EntryIcon(e *structure.Entry) string {
	icon := "📁"

	if e.IsDir {
		if e.HasChild() {
			icon = "📂"
		}

		return icon
	}

	switch e.Ext() {
	case "go", "py", "js", "ts", "java", "cpp", "c", "cs", "rb", "rs", "sh", "php":
		icon = "💻"
	case "jpg", "jpeg", "png", "gif", "bmp", "webp", "tiff":
		icon = "🖼"
	case "mp4", "mkv", "avi", "mov", "webm", "m4v", "wmv":
		icon = "🎞"
	case "json", "csv", "xml", "env", "yml", "yaml", "ini":
		icon = "🔧"
	case "jks", "pub", "key", "p12", "ppk":
		icon = "🔑"
	case "zip", "rar", "7z", "tar", "gz":
		icon = "🗃"
	case "mp3", "wav", "flac", "ogg":
		icon = "🎵"
	case "exe", "bin", "dll", "app":
		icon = "📦"
	case "doc", "docx":
		icon = "📝"
	case "xls", "xlsx":
		icon = "📊"
	case "ppt", "pptx":
		icon = "📈"
	case "html", "css":
		icon = "🌐"
	case "pdf":
		icon = "📕"
	case "md":
		icon = "📜"
	case "log":
		icon = "📗"
	case "iso":
		icon = "📀"
	default:
		icon = "📄"
	}

	return icon
}
