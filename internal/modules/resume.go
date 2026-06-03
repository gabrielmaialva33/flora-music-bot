package modules

import (
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/resume"] = `<i>Dá resume no playback pausado.</i>

<u>Uso:</u>
<b>/resume</b> — Continua o playback da pausa

<b>⚙️ Comportamento:</b>
• Continua da última posição pausada
• Cancela o timer de auto-resume se tiver ativo

<b>⚠️ Observações:</b>
• Só dá pra resumir se tiver pausado
• A posição é preservada durante a pausa
• As configs de velocidade continuam ativas depois do resume`
}

func resumeHandler(m *telegram.NewMessage) error {
	return handleResume(m, false)
}

func cresumeHandler(m *telegram.NewMessage) error {
	return handleResume(m, true)
}

func handleResume(m *telegram.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	if !r.IsActiveChat() {
		return replyEnd(m, "room_no_active")
	}

	if !r.IsPaused() {
		return replyEnd(m, "resume_already_playing")
	}

	t := r.Track()
	if _, err := r.Resume(); err != nil {
		m.Reply(F(chatID, "resume_failed", locales.Arg{
			"error": err,
		}))
	} else {
		title := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
		pos := formatDuration(r.Position())
		total := formatDuration(t.Duration)
		mention := utils.MentionHTML(m.Sender)

		m.Reply(F(chatID, "resume_success", locales.Arg{
			"title":      title,
			"position":   pos,
			"duration":   total,
			"user":       mention,
			"speed_line": speedLine(chatID, r),
		}))
	}

	return telegram.ErrEndGroup
}
