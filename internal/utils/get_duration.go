package utils

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetDurationByFFProbe(filePath string) (int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	result := strings.TrimSpace(out.String())
	seconds, err := strconv.ParseFloat(result, 64)
	if err != nil {
		return 0, err
	}

	return int(seconds), nil
}

func GetDuration(f *telegram.MessageMediaDocument) int {
	if f.Document == nil {
		return 0
	}
	d, ok := f.Document.(*telegram.DocumentObj)

	if !ok {
		return 0
	}

	for _, attr := range d.Attributes {
		switch a := attr.(type) {
		case *telegram.DocumentAttributeAudio:
			return int(a.Duration)
		case *telegram.DocumentAttributeVideo:
			return int(a.Duration)
		}
	}

	return 0
}
