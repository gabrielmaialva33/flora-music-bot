package modules

import (
	"fmt"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	state "main/internal/core/models"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/queue"] = `<i>Mostra a fila de playback atual.</i>

<u>Uso:</u>
<b>/queue</b> — Mostra a fila

<b>📋 Formato de exibição:</b>
• Tocando agora - Faixa atual com a posição
• Próximas - As próximas 10 faixas da fila
• Info da faixa: Título, quem pediu, duração

<b>⚙️ Features:</b>
• Status da fila em tempo real
• Atribuição de quem pediu
• Exibição de duração
• Indicador de tamanho da fila

<b>💡 Comandos relacionados:</b>
• <code>/position</code> - Só a posição da faixa atual
• <code>/remove</code> - Remove uma faixa específica
• <code>/clear</code> - Limpa todas as faixas
• <code>/move</code> - Reordena as faixas`

	helpTexts["/restore"] = `<i>Restaura uma fila de música que foi limpa antes.</i>

<u>Uso:</u>
<b>/restore</b> — Recupera as faixas

<b>⚙️ Comportamento:</b>
• Recupera as faixas limpas pelo último <code>/clear</code>
• Só funciona se nenhuma música nova tiver sido adicionada depois do clear
• As faixas restauradas são colocadas no final da fila atual

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar`

	helpTexts["/remove"] = `<i>Remove uma faixa específica da fila.</i>

<u>Uso:</u>
<b>/remove [índice]</b> — Remove a faixa da posição

<b>⚙️ Comportamento:</b>
• Índice começa do 1 (primeira faixa da fila)
• Não dá pra remover a faixa que tá tocando agora
• As posições da fila atualizam automaticamente

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/remove 1</code> — Remove a primeira faixa da fila
<code>/remove 5</code> — Remove a 5ª faixa

<b>⚠️ Observações:</b>
• Usa <code>/queue</code> pra ver os índices das faixas
• Índice inválido mostra erro com o tamanho da fila
• Usa <code>/clear</code> pra remover todas as faixas`

	helpTexts["/clear"] = `<i>Limpa todas as faixas da fila.</i>

<u>Uso:</u>
<b>/clear</b> — Remove todas as faixas da fila

<b>⚙️ Comportamento:</b>
• Remove todas as faixas da fila
• A faixa que tá tocando continua
• A fila fica vazia depois que a faixa atual acabar

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Dicas:</b>
Se você limpou a fila sem querer, usa <code>/restore</code> ou <code>/crestore</code> na hora pra recuperar (antes de adicionar música nova).`

	helpTexts["/move"] = `<i>Reordena as faixas da fila.</i>

<u>Uso:</u>
<b>/move [de] [pra]</b> — Move a faixa de uma posição pra outra

<b>⚙️ Comportamento:</b>
• Move a faixa do índice 'de' pro índice 'pra'
• As outras faixas deslocam as posições de acordo
• Índices começam do 1

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>💡 Exemplos:</b>
<code>/move 3 1</code> — Move a 3ª faixa pra 1ª posição
<code>/move 1 5</code> — Move a 1ª faixa pra 5ª posição

<b>⚠️ Observações:</b>
• As duas posições precisam ser índices válidos da fila
• Usa <code>/queue</code> pra ver a ordem atual
• Não dá pra mover a faixa que tá tocando agora`
}

func queueHandler(m *tg.NewMessage) error {
	return handleQueue(m, false)
}

func cqueueHandler(m *tg.NewMessage) error {
	return handleQueue(m, true)
}

func removeHandler(m *tg.NewMessage) error {
	return handleRemove(m, false)
}

func cremoveHandler(m *tg.NewMessage) error {
	return handleRemove(m, true)
}

func moveHandler(m *tg.NewMessage) error {
	return handleMove(m, false)
}

func cmoveHandler(m *tg.NewMessage) error {
	return handleMove(m, true)
}

func clearHandler(m *tg.NewMessage) error {
	return handleClear(m, false)
}

func cclearHandler(m *tg.NewMessage) error {
	return handleClear(m, true)
}

func restoreHandler(m *tg.NewMessage) error {
	return handleRestoreQueue(m, false)
}

func crestoreHandler(m *tg.NewMessage) error {
	return handleRestoreQueue(m, true)
}

func handleQueue(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	var b strings.Builder

	b.WriteString(F(chatID, "queue_header"))
	b.WriteString("\n\n")

	b.WriteString(F(chatID, "queue_now_playing"))
	b.WriteString("\n")

	fmt.Fprintf(
		&b,
		"🎧 <a href=\"%s\">%s</a> — %s [%s]\n\n",
		t.URL,
		utils.EscapeHTML(utils.ShortTitle(t.Title, 35)),
		t.Requester,
		formatDuration(t.Duration),
	)

	queue := r.Queue()
	q := len(queue)
	useQuote := q >= 3

	if q > 0 {
		b.WriteString(F(chatID, "queue_up_next"))
		n := "\n"
		if !useQuote {
			n += "\n"
		}
		b.WriteString(n)

		if useQuote {
			b.WriteString("<blockquote>")
		}

		for i, track := range queue {
			if i >= 10 {
				break
			}

			fmt.Fprintf(
				&b,
				"%d. 🎵 <a href=\"%s\">%s</a> — %s [%s]\n",
				i+1,
				track.URL,
				utils.EscapeHTML(utils.ShortTitle(track.Title, 35)),
				track.Requester,
				formatDuration(track.Duration),
			)
		}

		if useQuote {
			b.WriteString("</blockquote>")
		}

		if q > 10 {
			var full strings.Builder

			full.WriteString(F(chatID, "queue_header"))
			full.WriteString("\n\n")

			full.WriteString(F(chatID, "queue_now_playing"))
			full.WriteString("\n")

			fmt.Fprintf(
				&full,
				"🎧 %s — %s [%s]\n\n",
				t.Title,
				t.Requester,
				formatDuration(t.Duration),
			)

			full.WriteString(F(chatID, "queue_up_next"))
			full.WriteString("\n\n")

			for i, track := range queue {
				fmt.Fprintf(
					&full,
					"%d. %s — %s [%s]\n",
					i+1,
					track.Title,
					track.Requester,
					formatDuration(track.Duration),
				)
			}

			link, err := utils.CreatePaste(full.String())
			remaining := q - 10

			if err == nil && link != "" {
				more := fmt.Sprintf("<a href=\"%s\">%d</a>", link, remaining)

				b.WriteString(F(chatID, "queue_more_line", locales.Arg{
					"remaining": more,
				}))
			} else {
				b.WriteString(F(chatID, "queue_more_line", locales.Arg{
					"remaining": remaining,
				}))
			}
		}
	} else {
		b.WriteString(F(chatID, "queue_empty_tail"))
	}

	m.Reply(b.String())
	return tg.ErrEndGroup
}

func handleRemove(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 2 {
		m.Reply(F(chatID, "remove_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	index, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply(F(chatID, "remove_invalid_index"))
		return tg.ErrEndGroup
	}

	if index <= 0 {
		m.Reply(F(chatID, "remove_index_too_small"))
		return tg.ErrEndGroup
	}

	total := len(r.Queue())
	if index > total {
		m.Reply(F(chatID, "remove_index_too_big", locales.Arg{
			"total": total,
		}))
		return tg.ErrEndGroup
	}

	r.RemoveFromQueue(index - 1)

	m.Reply(F(chatID, "remove_success", locales.Arg{
		"index": index,
		"user":  utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}

func handleRestoreQueue(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()
	if !filterAuthUsers(m) {
		return tg.ErrEndGroup
	}

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	ok, v := r.GetData("last_queue")
	if !ok || v == nil {
		m.Reply(F(chatID, "queue_restore_no_data"))
		return tg.ErrEndGroup
	}

	tracks, ok := v.([]*state.Track)
	if !ok {
		r.DeleteData("last_queue")
		m.Reply(F(chatID, "queue_restore_no_data"))
		return tg.ErrEndGroup
	}

	r.AddTracksToQueue(tracks)
	r.DeleteData("last_queue")

	m.Reply(F(chatID, "queue_restored", locales.Arg{
		"count": len(tracks),
	}))

	return tg.ErrEndGroup
}

func handleClear(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}
	t := r.Track()
	if !r.IsActiveChat() || t == nil {
		m.Reply(F(chatID, "clear_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	r.SetData("last_queue", r.Queue())
	r.RemoveFromQueue(-1)

	restoreCmd := "restore"
	if cplay {
		restoreCmd = "crestore"
	}

	m.Reply(F(chatID, "clear_success", locales.Arg{
		"user": utils.MentionHTML(m.Sender),
		"cmd":  restoreCmd,
	}))

	return tg.ErrEndGroup
}

func handleMove(m *tg.NewMessage, cplay bool) error {
	chatID := m.ChannelID()

	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return tg.ErrEndGroup
	}

	if !r.IsActiveChat() || r.Track() == nil {
		m.Reply(F(chatID, "queue_no_active"))
		return tg.ErrEndGroup
	}

	if len(r.Queue()) == 0 {
		m.Reply(F(chatID, "queue_empty"))
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())
	if len(args) < 3 {
		m.Reply(F(chatID, "move_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	from, err1 := strconv.Atoi(args[1])
	to, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		m.Reply(F(chatID, "move_invalid_numbers", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	if from <= 0 || to <= 0 {
		m.Reply(F(chatID, "move_invalid_indexes_min"))
		return tg.ErrEndGroup
	}

	queueLen := len(r.Queue())
	if from > queueLen || to > queueLen {
		m.Reply(F(chatID, "move_invalid_indexes_max", locales.Arg{
			"queue_len": queueLen,
		}))
		return tg.ErrEndGroup
	}

	r.MoveInQueue(from-1, to-1)

	m.Reply(F(chatID, "move_success", locales.Arg{
		"from": from,
		"to":   to,
		"user": utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}
