package modules

import (
	"fmt"

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
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	if !r.IsPaused() {
		m.Reply(F(chatID, "resume_already_playing"))
		return telegram.ErrEndGroup
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

		speedLine := ""
		if sp := r.Speed(); sp != 1.0 {
			speedLine = F(chatID, "speed_line", locales.Arg{
				"speed": fmt.Sprintf("%.2f", r.Speed()),
			})
		}

		m.Reply(F(chatID, "resume_success", locales.Arg{
			"title":      title,
			"position":   pos,
			"duration":   total,
			"user":       mention,
			"speed_line": speedLine,
		}))
	}

	return telegram.ErrEndGroup
}
