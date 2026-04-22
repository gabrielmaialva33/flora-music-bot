package modules

import (
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/shuffle"] = `<i>Liga/desliga o modo shuffle da fila.</i>

<u>Uso:</u>
<b>/shuffle</b> — Mostra o estado atual do shuffle
<b>/shuffle on</b> — Ativa o shuffle
<b>/shuffle off</b> — Desativa o shuffle

<b>⚙️ Comportamento:</b>
• Embaralha a fila aleatoriamente quando ativado
• Afeta a ordem de seleção das faixas
• Pode ser ligado/desligado a qualquer momento

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/shuffle on</code> — Ativa o modo shuffle
<code>/shuffle off</code> — Desativa o modo shuffle

<b>⚠️ Observação:</b>
O shuffle só afeta a ordem da fila, não a faixa que tá tocando.`
}

func shuffleHandler(m *telegram.NewMessage) error {
	return handleShuffle(m, false)
}

func cshuffleHandler(m *telegram.NewMessage) error {
	return handleShuffle(m, true)
}

func handleShuffle(m *telegram.NewMessage, cplay bool) error {
	arg := strings.ToLower(m.Args())

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}
	chatID := m.ChannelID()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return telegram.ErrEndGroup
	}

	r.Parse()

	if arg == "" {
		state := F(chatID, "disabled")
		cmd := getCommand(m) + " on"
		if r.Shuffle() {
			state = F(chatID, "enabled")
			cmd = getCommand(m) + " off"
		}

		m.Reply(F(chatID, "shuffle_current_state", locales.Arg{
			"state": state,
			"cmd":   cmd,
		}))
		return telegram.ErrEndGroup
	}

	var newState bool
	if arg == "on" || arg == "enable" || arg == "true" || arg == "1" {
		newState = true
	} else if arg == "off" || arg == "disable" || arg == "false" || arg == "0" {
		newState = false
	}

	r.SetShuffle(newState)

	state := F(chatID, "disabled")
	if newState {
		state = F(chatID, "enabled")
	}

	m.Reply(F(chatID, "shuffle_updated", locales.Arg{
		"state": state,
		"user":  utils.MentionHTML(m.Sender),
	}))

	return telegram.ErrEndGroup
}
