package modules

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/position"] = `<i>Mostra a posição atual do playback e info da faixa.</i>

<u>Uso:</u>
<b>/position</b> — Mostra a posição

<b>📊 Informações exibidas:</b>
• Título da faixa atual
• Posição atual (MM:SS)
• Duração total (MM:SS)
• Velocidade do playback (se não for 1.0x)

<b>💡 Caso de uso:</b>
Checagem rápida de posição sem mostrar a fila inteira.`
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
