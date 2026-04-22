package modules

import (
	"html"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var helpTexts = map[string]string{}

func checkForHelpFlag(m *tg.NewMessage) bool {
	text := strings.Fields(strings.ToLower(strings.TrimSpace(m.Text())))
	for _, t := range text {
		switch t {
		case "-h", "--h", "-help", "--help", "help":
			return true
		}
	}
	return false
}

func showHelpFor(m *tg.NewMessage, cmd string) error {
	help, ok := helpTexts[cmd]
	if !ok {
		alt := strings.TrimPrefix(cmd, "/")
		if h, ok := helpTexts[alt]; ok {
			help = h
		}
	}
	if help == "" {
		_, err := m.Reply(
			"⚠️ <i>No help found for command <code>" + html.EscapeString(
				cmd,
			) + "</code></i>",
		)
		if err != nil {
			return err
		}
		return tg.ErrEndGroup
	}
	_, err := m.Reply("📘 <b>Help for</b> <code>" + cmd + "</code>:\n\n" + help)
	if err != nil {
		return err
	}
	return tg.ErrEndGroup
}
