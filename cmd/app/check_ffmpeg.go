package main

import (
	"os/exec"

	"github.com/Laky-64/gologging"
)

func checkFFmpegAndFFprobe() {
	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(bin); err != nil {
			gologging.FatalF(
				"❌ %s not found in PATH. Please install %s.",
				bin,
				bin,
			)
		}
	}
}
