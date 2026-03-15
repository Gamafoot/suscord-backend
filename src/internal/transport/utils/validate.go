package utils

import (
	"mime"
	"path/filepath"
	"strings"
)

func FilenameValidate(filename string, allowedMedia []string) bool {
	ok := false
	mimetype := GetMimeType(filename)

	for _, allowed := range allowedMedia {
		if strings.HasPrefix(mimetype, allowed) {
			ok = true
		}
	}

	return ok
}

func GetMimeType(filename string) string {
	ext := filepath.Ext(strings.ToLower(filename))
	return mime.TypeByExtension(ext)
}

func IsImage(filename string) bool {
	ext := filepath.Ext(strings.ToLower(filename))
	mimetype := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimetype, "image")
}
