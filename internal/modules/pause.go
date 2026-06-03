package modules

import (
	"errors"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/pause"] = `<i>Pausa o playback atual.</i>

<u>Uso:</u>
<b>/pause</b> — Pausa o playback
<b>/pause [segundos]</b> — Pausa com auto-resume depois dos segundos especificados

<b>⚙️ Features:</b>
• Controle manual de pause/resume
• Timer de auto-resume (5-3600 segundos)

<b>💡 Exemplos:</b>
<code>/pause</code> — Pausa por tempo indeterminado
<code>/pause 30</code> — Pausa por 30 segundos e depois dá resume automático
`
}

func pauseHandler(m *tg.NewMessage) error {
	return handlePause(m, false)
}

func cpauseHandler(m *tg.NewMessage) error {
	return handlePause(m, true)
}

func handlePause(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() {
		return replyEnd(m, "room_no_active")
	}

	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		autoResumeLine := ""
		if remaining > 0 {
			autoResumeLine = F(chatID, "auto_resume_line", locales.Arg{
				"seconds": formatDuration(int(remaining.Seconds())),
			})
		}
		return replyEnd(m, "pause_already", locales.Arg{
			"auto_resume_line": autoResumeLine,
		})
	}

	args := strings.Fields(m.Text())
	var autoResumeDuration time.Duration
	if len(args) >= 2 {
		sec, err := parseDurationArg(args[1], 5, 3600)
		switch {
		case errors.Is(err, errOutOfRange):
			return replyEnd(m, "pause_invalid_duration")
		case errors.Is(err, errBadDuration):
			return replyEnd(m, "pause_invalid_format", locales.Arg{
				"cmd": getCommand(m),
			})
		}
		autoResumeDuration = time.Duration(sec) * time.Second
	}

	var pauseErr error
	if autoResumeDuration > 0 {
		_, pauseErr = r.Pause(autoResumeDuration)
	} else {
		_, pauseErr = r.Pause()
	}
	if pauseErr != nil {
		return replyEnd(m, "room_pause_failed", locales.Arg{
			"error": pauseErr.Error(),
		})
	}

	mention := utils.MentionHTML(m.Sender)
	title := utils.EscapeHTML(utils.ShortTitle(r.Track().Title, 25))

	autoResumeLine := ""
	if autoResumeDuration > 0 {
		autoResumeLine = F(chatID, "auto_resume_line", locales.Arg{
			"seconds": int(autoResumeDuration.Seconds()),
		})
	}

	msg := F(chatID, "pause_success", locales.Arg{
		"title":            title,
		"position":         formatDuration(r.Position()),
		"duration":         formatDuration(r.Track().Duration),
		"user":             mention,
		"auto_resume_line": autoResumeLine,
	})

	if sl := speedLine(chatID, r); sl != "" {
		msg += "\n" + sl
	}

	m.Reply(msg)
	return tg.ErrEndGroup
}
