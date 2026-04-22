package modules

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/position"] = `<i>Show current playback position and track info.</i>

<u>Usage:</u>
<b>/position</b> — Show position

<b>📊 Information Displayed:</b>
• Current track title
• Current position (MM:SS)
• Total duration (MM:SS)
• Playback speed (if not 1.0x)

<b>💡 Use Case:</b>
Quick position check without full queue display.`
}

func positionHandler(m *tg.NewMessage) error {
	return handlePosition(m, false)
}

func cpositionHandler(m *tg.NewMessage) error {
	return handlePosition(m, true)
}

func handlePosition(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() || r.Track().ID == "" {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	r.Parse()

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))

	m.Reply(F(chatID, "position_now", locales.Arg{
		"title":    title,
		"position": formatDuration(r.Position()),
		"duration": formatDuration(r.Track().Duration),
		"speed":    fmt.Sprintf("%.2f", r.Speed()),
	}))

	return tg.ErrEndGroup
}
