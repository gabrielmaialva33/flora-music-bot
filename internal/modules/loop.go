package modules

import (
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/loop"] = `<i>Define quantas vezes a faixa atual vai repetir.</i>

<u>Uso:</u>
<b>/loop</b> — Mostra a contagem de loop atual
<b>/loop [contagem]</b> — Define o loop (0-10)

<b>⚙️ Comportamento:</b>
• 0 = Sem loop (toca uma vez só)
• 1-10 = Repete a faixa essa quantidade de vezes
• O contador decrementa depois de cada execução

<b>💡 Exemplos:</b>
<code>/loop 0</code> — Desativa o loop
<code>/loop 3</code> — Repete a faixa atual 3 vezes
<code>/loop 10</code> — Repete a faixa atual 10 vezes

<b>⚠️ Observações:</b>
• Limite máximo de loop: 10
• O loop afeta só a faixa atual
• Depois que os loops acabam, toca a próxima da fila
• Se a faixa for pulada com <code>/skip</code>, o loop para e reseta automaticamente`
}

func loopHandler(m *tg.NewMessage) error {
	return handleLoop(m, false)
}

func cloopHandler(m *tg.NewMessage) error {
	return handleLoop(m, true)
}

func handleLoop(m *tg.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	chatID := m.ChannelID()
	args := strings.Fields(m.Text())
	currentLoop := r.Loop()

	if !r.IsActiveChat() {
		m.Reply(F(chatID, "room_no_active"))
		return tg.ErrEndGroup
	}

	if len(args) < 2 {
		countLine := ""
		if currentLoop > 0 {
			countLine = "\n" + F(chatID, "loop_current", locales.Arg{
				"count": currentLoop,
			})
		}

		msg := F(m.ChannelID(), "loop_usage", locales.Arg{
			"cmd":        getCommand(m),
			"count_line": countLine,
		})

		m.Reply(msg)
		return tg.ErrEndGroup
	}

	newLoop, err := strconv.Atoi(args[1])
	if err != nil || newLoop < 0 || newLoop > 10 {
		m.Reply(F(chatID, "loop_invalid"))
		return tg.ErrEndGroup
	}

	if newLoop == currentLoop {
		m.Reply(F(chatID, "loop_already_set", locales.Arg{
			"count": currentLoop,
		}))
		return tg.ErrEndGroup
	}

	r.SetLoop(newLoop)

	mention := utils.MentionHTML(m.Sender)
	var msg string
	if newLoop == 0 {
		msg = F(chatID, "loop_disabled", locales.Arg{
			"user": mention,
		})
	} else {
		msg = F(chatID, "loop_set", locales.Arg{
			"count": newLoop,
			"user":  mention,
		})
	}

	m.Reply(msg)
	return tg.ErrEndGroup
}
