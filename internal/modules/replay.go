package modules

import (
	"fmt"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/replay"] = `<i>Reinicia a faixa atual do começo.</i>

<u>Uso:</u>
<b>/replay</b> — Reinicia a faixa atual

<b>⚙️ Comportamento:</b>
• Volta a posição pra 0:00
• Mantém a config de velocidade
• Continua o playback imediatamente

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar
`
}

func replayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, false)
}

func creplayHandler(m *telegram.NewMessage) error {
	return handleReplay(m, true)
}

func handleReplay(m *telegram.NewMessage, cplay bool) error {
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
	t := r.Track()

	if err := r.Replay(); err != nil {
		m.Reply(F(chatID, "replay_failed", locales.Arg{
			"error": err,
		}))
	} else {
		trackTitle := utils.EscapeHTML(utils.ShortTitle(t.Title, 25))
		totalDuration := formatDuration(t.Duration)
		m.Reply(F(chatID, "replay_success", locales.Arg{
			"title":    trackTitle,
			"duration": totalDuration,
			"speed":    fmt.Sprintf("%.2f", r.Speed()),
		}))
	}

	return telegram.ErrEndGroup
}
