package urlpath

import (
	"strings"
)

func GetMediaURL(mediaURL, filepath string) string {
	if filepath == "" {
		return ""
	}
	if mediaURL == "" {
		return filepath
	}

	mediaURL = strings.TrimSuffix(mediaURL, "/")
	filepath = strings.TrimPrefix(filepath, "/")

	return mediaURL + "/" + filepath
}
