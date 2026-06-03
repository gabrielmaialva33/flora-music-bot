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
		return replyEnd(m, "room_no_active")
	}

	r.Parse()

	if arg == "" {
		state := F(chatID, "disabled")
		cmd := getCommand(m) + " on"
		if r.Shuffle() {
			state = F(chatID, "enabled")
			cmd = getCommand(m) + " off"
		}

		return replyEnd(m, "shuffle_current_state", locales.Arg{
			"state": state,
			"cmd":   cmd,
		})
	}

	var newState bool
	switch arg {
	case "on", "enable", "true", "1":
		newState = true
	case "off", "disable", "false", "0":
		newState = false
	default:
		// Input inválido: mostra o estado/uso atual em vez de desligar shuffle
		// silenciosamente (o zero-value de newState era false).
		state := F(chatID, "disabled")
		cmd := getCommand(m) + " on"
		if r.Shuffle() {
			state = F(chatID, "enabled")
			cmd = getCommand(m) + " off"
		}
		return replyEnd(m, "shuffle_current_state", locales.Arg{
			"state": state,
			"cmd":   cmd,
		})
	}

	r.SetShuffle(newState)

	state := F(chatID, "disabled")
	if newState {
		state = F(chatID, "enabled")
	}

	return replyEnd(m, "shuffle_updated", locales.Arg{
		"state": state,
		"user":  utils.MentionHTML(m.Sender),
	})
}
