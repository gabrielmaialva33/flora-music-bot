package modules

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/unmute"] = `<i>Desmuta a saída de áudio do chat de voz.</i>

<u>Uso:</u>
<b>/unmute</b> — Volta o áudio

<b>⚙️ Comportamento:</b>
• Volta o áudio imediatamente
• Cancela o timer de auto-unmute se tiver ativo
• Mostra info do playback atual`
}

func unmuteHandler(m *tg.NewMessage) error {
	return handleUnmute(m, false)
}

func cunmuteHandler(m *tg.NewMessage) error {
	return handleUnmute(m, true)
}

func handleUnmute(m *tg.NewMessage, cplay bool) error {
	if m.Args() != "" {
		return tg.ErrEndGroup
	}
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	chatID := m.ChannelID()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if !r.IsMuted() {
		m.Reply(F(chatID, "unmute_already"))
		return tg.ErrEndGroup
	}

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))
	mention := utils.MentionHTML(m.Sender)

	if _, err := r.Unmute(); err != nil {
		m.Reply(F(chatID, "unmute_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	// optional speed line
	var speedOpt string
	if sp := r.Speed(); sp != 1.0 {
		speedOpt = F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	msg := F(chatID, "unmute_success", locales.Arg{
		"title":      title,
		"user":       mention,
		"speed_line": speedOpt,
	})

	m.Reply(msg)
	return tg.ErrEndGroup
}
