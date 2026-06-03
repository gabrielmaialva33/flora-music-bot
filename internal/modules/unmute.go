package modules

import (
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
		return replyEnd(m, "room_no_active")
	}

	if !r.IsMuted() {
		return replyEnd(m, "unmute_already")
	}

	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))
	mention := utils.MentionHTML(m.Sender)

	if _, err := r.Unmute(); err != nil {
		return replyEnd(m, "unmute_failed", locales.Arg{
			"error": err.Error(),
		})
	}

	msg := F(chatID, "unmute_success", locales.Arg{
		"title":      title,
		"user":       mention,
		"speed_line": speedLine(chatID, r),
	})

	m.Reply(msg)
	return tg.ErrEndGroup
}
