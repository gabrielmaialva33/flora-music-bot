package modules

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/speed"] = `<i>Controla a velocidade do playback (tempo).</i>

<u>Uso:</u>
<b>/speed</b> — Mostra a velocidade atual
<b>/speed [multiplicador]</b> — Define a velocidade (0.5-4.0x)
<b>/speed [multiplicador] [segundos]</b> — Define com timer de auto-reset
<b>/speed normal</b> ou <b>/speed reset</b> — Volta pra 1.0x

<b>⚙️ Features:</b>
• Range: 0.50x até 4.00x
• Timer de auto-reset (5-3600 segundos)
• Preservação de pitch
• Ajuste em tempo real

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/speed 1.5</code> — Toca 1.5x mais rápido
<code>/speed 0.75</code> — Toca mais devagar (0.75x)
<code>/speed 2.0 300</code> — Velocidade 2x por 5 minutos, depois reseta
<code>/speed normal</code> — Volta pra velocidade normal

<b>⚠️ Observações:</b>
• A velocidade afeta os cálculos de duração
• Auto-reset só funciona pra velocidades diferentes de 1.0x
• Sufixo 'x' é opcional: <code>1.5</code> = <code>1.5x</code>`
}

func speedHandler(m *telegram.NewMessage) error {
	return handleSpeed(m, false)
}

func cspeedHandler(m *telegram.NewMessage) error {
	return handleSpeed(m, true)
}

func handleSpeed(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	t := r.Track()

	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())

	// No args -> show current speed or usage hint
	if len(args) < 2 {
		if r.Speed() != 1.0 {
			remaining := r.RemainingSpeedDuration()
			if remaining > 0 {
				m.Reply(F(chatID, "speed_current_with_reset", locales.Arg{
					"speed": fmt.Sprintf("%.2f", r.Speed()),
					"title": utils.EscapeHTML(
						utils.ShortTitle(t.Title, 25),
					),
					"duration": formatDuration(int(remaining.Seconds())),
					"cmd":      getCommand(m),
				}))
			} else {
				m.Reply(F(chatID, "speed_current", locales.Arg{
					"speed": fmt.Sprintf("%.2f", r.Speed()),
					"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
					"cmd":   getCommand(m),
				}))
			}
		} else {
			m.Reply(F(chatID, "speed_usage", locales.Arg{
				"cmd": getCommand(m),
			}))
		}
		return telegram.ErrEndGroup
	}

	// Parse speed
	raw := strings.ToLower(strings.TrimSpace(args[1]))
	raw = strings.TrimSuffix(raw, "x")
	raw = strings.TrimSuffix(raw, "×")

	var newSpeed float64
	if raw == "normal" || raw == "reset" || raw == "default" {
		newSpeed = 1.0
	} else {
		s, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			m.Reply(F(chatID, "speed_invalid_value", locales.Arg{
				"cmd": getCommand(m),
			}))
			return telegram.ErrEndGroup
		}
		if s < 0.50 || s > 4.0 {
			m.Reply(F(chatID, "speed_invalid_range"))
			return telegram.ErrEndGroup
		}
		newSpeed = s
	}

	// Parse auto-reset duration
	var resetDuration time.Duration
	if len(args) >= 3 {
		d := strings.ToLower(strings.TrimSpace(args[2]))
		d = strings.TrimSuffix(d, "s")

		seconds, err := strconv.Atoi(d)
		if err != nil || seconds < 5 || seconds > 3600 {
			m.Reply(F(chatID, "speed_invalid_duration"))
			return telegram.ErrEndGroup
		}
		resetDuration = time.Duration(seconds) * time.Second
	}

	// Same speed → give info
	if newSpeed == r.Speed() {
		if resetDuration == 0 {
			m.Reply(F(chatID, "speed_already_set", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
			}))
		} else if newSpeed != 1.0 {
			m.Reply(F(chatID, "speed_already_set_reset_hint", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"title": utils.EscapeHTML(utils.ShortTitle(t.Title, 25)),
				"cmd":   getCommand(m),
			}))
		}
		return telegram.ErrEndGroup
	}

	// Apply speed
	var setErr error
	if resetDuration > 0 && newSpeed != 1.0 {
		setErr = r.SetSpeed(newSpeed, resetDuration)
	} else {
		setErr = r.SetSpeed(newSpeed)
	}

	if setErr != nil {
		m.Reply(F(chatID, "speed_failed", locales.Arg{
			"speed": fmt.Sprintf("%.2f", newSpeed),
			"error": setErr.Error(),
		}))
		return telegram.ErrEndGroup
	}

	mention := utils.MentionHTML(m.Sender)

	if newSpeed == 1.0 {
		m.Reply(F(chatID, "speed_reset_success", locales.Arg{
			"user": mention,
		}))
	} else {
		if resetDuration > 0 {
			m.Reply(F(chatID, "speed_set_with_reset", locales.Arg{
				"speed":    fmt.Sprintf("%.2f", newSpeed),
				"user":     mention,
				"duration": int(resetDuration.Seconds()),
			}))
		} else {
			m.Reply(F(chatID, "speed_set", locales.Arg{
				"speed": fmt.Sprintf("%.2f", newSpeed),
				"user":  mention,
			}))
		}
	}

	return telegram.ErrEndGroup
}
