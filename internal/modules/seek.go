package modules

import (
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/locales"
)

func init() {
	helpTexts["/seek"] = `<i>Pula pra frente na faixa que tá tocando.</i>

<u>Uso:</u>
<b>/seek [segundos]</b> — Pula pra frente a quantidade de segundos

<b>⚙️ Features:</b>
• Pula pra frente na faixa atual
• O tracking da posição é atualizado
• Não dá pra passar do fim da faixa (buffer de 10s)

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/seek 30</code> — Pula 30 segundos pra frente
<code>/seek 120</code> — Pula 2 minutos pra frente

<b>⚠️ Observações:</b>
• Mínimo: qualquer valor positivo
• Máximo: duração_da_faixa - posição_atual - 10 segundos`

	helpTexts["/seekback"] = `<i>Volta pra trás na faixa que tá tocando.</i>

<u>Uso:</u>
<b>/seekback [segundos]</b> — Volta a quantidade de segundos

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/seekback 15</code> — Volta 15 segundos
<code>/seekback 60</code> — Volta 1 minuto
`

	helpTexts["/jump"] = `<i>Pula pra uma posição específica na faixa.</i>

<u>Uso:</u>
<b>/jump [segundos]</b> — Pula pra posição exata

<b>⚙️ Features:</b>
• Seek de posição absoluta
• Controle preciso de tempo
• Buffer de 10 segundos do final

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/jump 90</code> — Pula pra 1:30
<code>/jump 0</code> — Pula pro início (igual ao /replay)

<b>⚠️ Observações:</b>
• A posição precisa estar dentro da duração da faixa - 10 segundos
• Mais preciso que <code>/seek</code> e <code>/seekback</code>`
}

func seekHandler(m *telegram.NewMessage) error {
	return handleSeek(m, false, false)
}

func cseekHandler(m *telegram.NewMessage) error {
	return handleSeek(m, true, false)
}

func seekbackHandler(m *telegram.NewMessage) error {
	return handleSeek(m, false, true)
}

func cseekbackHandler(m *telegram.NewMessage) error {
	return handleSeek(m, true, true)
}

func jumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, false)
}

func cjumpHandler(m *telegram.NewMessage) error {
	return handleJump(m, true)
}

func handleSeek(m *telegram.NewMessage, cplay, isBack bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}
	chatID := m.ChannelID()
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "seek_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "seek_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(F(chatID, "seek_invalid_seconds", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	var direction, emoji string
	var seekErr error

	if isBack {
		if (r.Position() - seconds) <= 10 {
			m.Reply(F(chatID, "seek_too_close_start", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.ErrEndGroup
		}
		seekErr = r.Seek(-seconds)
		direction = "backward"
		emoji = "⏪"
	} else {
		if (t.Duration - (r.Position() + seconds)) <= 10 {
			m.Reply(F(chatID, "seek_too_close_end", locales.Arg{
				"seconds": seconds,
			}))
			return telegram.ErrEndGroup
		}
		seekErr = r.Seek(seconds)
		direction = "forward"
		emoji = "⏩"
	}

	if seekErr != nil {
		m.Reply(F(chatID, "seek_failed", locales.Arg{
			"direction": direction,
			"seconds":   seconds,
			"error":     seekErr,
		}))
	} else {
		m.Reply(F(chatID, "seek_success", locales.Arg{
			"emoji":     emoji,
			"direction": direction,
			"position":  formatDuration(r.Position()),
			"duration":  formatDuration(t.Duration),
		}))
	}

	return telegram.ErrEndGroup
}

func handleJump(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	t := r.Track()

	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "jump_no_active"))
		return telegram.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "jump_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	seconds, err := strconv.Atoi(args[1])
	if err != nil || seconds < 0 {
		m.Reply(F(chatID, "jump_invalid_position", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	if t.Duration-seconds <= 10 {
		m.Reply(F(chatID, "jump_too_close_end", locales.Arg{
			"position": formatDuration(seconds),
		}))
		return telegram.ErrEndGroup
	}

	if err := r.Seek(seconds - r.Position()); err != nil {
		m.Reply(F(chatID, "jump_failed", locales.Arg{
			"position": formatDuration(seconds),
			"error":    err,
		}))
	} else {
		m.Reply(F(chatID, "jump_success", locales.Arg{
			"position": formatDuration(seconds),
			"duration": formatDuration(t.Duration),
		}))
	}

	return telegram.ErrEndGroup
}
