/* Package resolver */
package resolver

import (
	"os/exec"

	"media/internal/model"
)

func ResolveYouTube(url string) model.PlaybackSource {
	return model.PlaybackSource{
		Kind:     "process",
		Cmd:      exec.Command("yt-dlp", "-o", "-", url),
		Seekable: false,
	}
}

func ResolveDirect(url string) model.PlaybackSource {
	return model.PlaybackSource{
		Kind:     "url",
		URL:      url,
		Seekable: true,
	}
}
