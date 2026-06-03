package utils

import (
	"fmt"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func GetProgress(statusMsg *telegram.NewMessage) *telegram.ProgressManager {
	pm := telegram.NewProgressManager(2)

	if statusMsg == nil {
		return pm
	}

	var opts *telegram.SendOptions
	if replyMarkup := statusMsg.ReplyMarkup(); replyMarkup != nil {
		opts = &telegram.SendOptions{ReplyMarkup: *replyMarkup}
	}

	pm.WithCallback(func(pi *telegram.ProgressInfo) {
		text := fmt.Sprintf(
			"<b>📥 Downloading your track...</b>\n"+
				"<pre>"+
				"Progress : %6.2f%%\n"+
				"Speed    : %s\n"+
				"Eta      : %s\n"+
				"Elapsed  : %s"+
				"</pre>",
			pi.Percentage,
			pi.SpeedString(),
			pi.ETAString(),
			pi.ElapsedString(),
		)
		statusMsg.Edit(text, opts)
	})

	return pm
}

func GetProgressBar(playedSec, durationSec int) string {
	if durationSec <= 0 || playedSec <= 0 {
		return "◉—————————"
	}

	if playedSec >= durationSec {
		return "—————————◉"
	}

	index := (playedSec * 10) / durationSec
	if index > 9 {
		index = 9
	}

	return strings.Repeat("—", index) + "◉" + strings.Repeat("—", 9-index)
}
