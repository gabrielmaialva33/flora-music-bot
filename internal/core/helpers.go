package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Laky-64/gologging"
)

func normalizeVideo(path string, speed float64) (int, int, int, string) {
	if speed <= 0 {
		speed = 1.0
	}
	w, h := getVideoDimensions(path)
	if w <= 0 || h <= 0 {
		w = 1280
		h = 720
	}
	maxW := 1280
	maxH := 720
	if w > maxW {
		h = h * maxW / w
		w = maxW
	}
	if h > maxH {
		w = w * maxH / h
		h = maxH
	}
	if w%2 != 0 {
		w--
	}
	if h%2 != 0 {
		h--
	}
	fps := 30
	videoSpeed := 1.0 / speed
	filter := fmt.Sprintf("setpts=%.4f*PTS,scale=%d:%d", videoSpeed, w, h)
	return w, h, fps, filter
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func getVideoDimensions(filePath string) (int, int) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	out, err := cmd.Output()
	if ctx.Err() == context.DeadlineExceeded {
		gologging.Error(
			"[getVideoDimensions] ffprobe timed out for " + filePath,
		)
		return 0, 0
	}
	if err != nil {
		gologging.Error(
			"[getVideoDimensions] ffprobe failed for " + filePath + " : " + err.Error(),
		)
		return 0, 0
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		gologging.Error(
			"[getVideoDimensions] failed to parse ffprobe JSON for " + filePath + " : " + err.Error(),
		)
		return 0, 0
	}

	for _, s := range probe.Streams {
		if s.CodecType == "video" && s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}

	for _, s := range probe.Streams {
		if s.Width > 0 && s.Height > 0 {
			return s.Width, s.Height
		}
	}

	gologging.Error(
		"[getVideoDimensions] no valid video stream found for " + filePath,
	)
	return 0, 0
}

func isStreamURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}
