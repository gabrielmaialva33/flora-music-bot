package modules

import (
	"context"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/platforms"
	"main/internal/utils"
)

func init() {
	helpTexts["/skip"] = `<i>Pula a faixa que tá tocando e toca a próxima da fila.</i>

<u>Uso:</u>
<b>/skip</b> — Pula a faixa atual

<b>⚙️ Comportamento:</b>
• Baixa a próxima faixa da fila
• Começa o playback automaticamente
• Se a fila tiver vazia e o loop for 0, para o playback

<b>🔒 Restrições:</b>
• Só <b>admins do chat</b> ou <b>usuários autorizados</b> podem usar

<b>⚠️ Observações:</b>
• Não dá pra desfazer
• Se não tiver faixa na fila, o playback para
• A contagem de loop afeta o comportamento do skip`
}

func skipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, false)
}

func cskipHandler(m *telegram.NewMessage) error {
	return handleSkip(m, true)
}

func handleSkip(m *telegram.NewMessage, cplay bool) error {
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

	mention := utils.MentionHTML(m.Sender)

	if len(r.Queue()) == 0 {
		core.DeleteRoom(r.ChatID())
		m.Reply(F(chatID, "skip_stopped", locales.Arg{
			"user": mention,
		}))
		return telegram.ErrEndGroup
	}

	r.SetLoop(0)
	t := r.NextTrack()

	statusMsg, err := core.Bot.SendMessage(
		chatID,
		F(chatID, "stream_downloading_next"),
	)
	if err != nil {
		gologging.ErrorF("[skip.go] err: %v", err)
	}

	path, err := platforms.Download(context.Background(), t, statusMsg)
	if err != nil {
		txt := F(chatID, "stream_download_fail", locales.Arg{
			"error": err.Error(),
		})

		if statusMsg != nil {
			utils.EOR(statusMsg, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}

		core.DeleteRoom(r.ChatID())
		return telegram.ErrEndGroup
	}

	if err := r.Play(t, path, true); err != nil {
		txt := F(chatID, "stream_play_fail")
		if statusMsg != nil {
			utils.EOR(statusMsg, txt)
		} else {
			core.Bot.SendMessage(chatID, txt)
		}
		core.DeleteRoom(r.ChatID())
		return telegram.ErrEndGroup
	}

	title := utils.ShortTitle(t.Title, 25)
	safeTitle := utils.EscapeHTML(title)

	msg := F(chatID, "stream_now_playing", locales.Arg{
		"url":      t.URL,
		"title":    safeTitle,
		"duration": formatDuration(t.Duration),
		"by":       t.Requester,
	})

	opt := &telegram.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: core.GetPlayMarkup(chatID, r, false),
	}

	if t.Artwork != "" && shouldShowThumb(chatID) {
		opt.Media = utils.CleanURL(t.Artwork)
	}

	var newStatusMsg *telegram.NewMessage
	if statusMsg != nil {
		newStatusMsg, _ = utils.EOR(statusMsg, msg, opt)
	} else {
		newStatusMsg, _ = core.Bot.SendMessage(chatID, msg, opt)
	}

	if newStatusMsg != nil {
		r.SetStatusMsg(newStatusMsg)
	}

	return telegram.ErrEndGroup
}
