package modules

import (
	"fmt"
	"strconv"
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
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if r.IsPaused() {
		remaining := r.RemainingResumeDuration()
		autoResumeLine := ""
		if remaining > 0 {
			autoResumeLine = F(chatID, "auto_resume_line", locales.Arg{
				"seconds": formatDuration(int(remaining.Seconds())),
			})
		}
		m.Reply(F(chatID, "pause_already", locales.Arg{
			"auto_resume_line": autoResumeLine,
		}))
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	var autoResumeDuration time.Duration
	if len(args) >= 2 {
		raw := strings.ToLower(strings.TrimSpace(args[1]))
		raw = strings.TrimSuffix(raw, "s")
		if sec, convErr := strconv.Atoi(raw); convErr == nil {
			if sec < 5 || sec > 3600 {
				m.Reply(F(chatID, "pause_invalid_duration"))
				return tg.ErrEndGroup
			}
			autoResumeDuration = time.Duration(sec) * time.Second
		} else {
			m.Reply(F(chatID, "pause_invalid_format", locales.Arg{
				"cmd": getCommand(m),
			}))
			return tg.ErrEndGroup
		}
	}

	var pauseErr error
	if autoResumeDuration > 0 {
		_, pauseErr = r.Pause(autoResumeDuration)
	} else {
		_, pauseErr = r.Pause()
	}
	if pauseErr != nil {
		m.Reply(F(chatID, "room_pause_failed", locales.Arg{
			"error": pauseErr.Error(),
		}))
		return tg.ErrEndGroup
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

	if sp := r.Speed(); sp != 1.0 {
		msg += "\n" + F(chatID, "speed_line", locales.Arg{
			"speed": fmt.Sprintf("%.2f", sp),
		})
	}

	m.Reply(msg)
	return tg.ErrEndGroup
}
