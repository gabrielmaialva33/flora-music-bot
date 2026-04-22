package modules

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/nothumb"] = `<i>Liga/desliga a exibição de thumbnail/capa nas mensagens de playback.</i>

<u>Uso:</u>
<b>/nothumb</b> — Mostra o status atual
<b>/nothumb [enable|disable]</b> — Muda a configuração

<b>⚙️ Comportamento:</b>
• <b>Desativado (padrão):</b> Mostra a capa/thumbnail da faixa
• <b>Ativado:</b> Esconde a capa, mensagens só de texto

<b>💡 Exemplos:</b>
<code>/nothumb enable</code> — Desativa as thumbnails
<code>/nothumb disable</code> — Ativa as thumbnails

<b>⚠️ Observação:</b>
Essa config afeta todas as mensagens de playback futuras nesse chat.`
}

func nothumbHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	args := strings.Fields(m.Text())

	current, err := database.ThumbnailsDisabled(chatID)
	if err != nil {
		m.Reply(F(chatID, "nothumb_fetch_fail"))
		return tg.ErrEndGroup
	}

	if len(args) < 2 {
		action := utils.IfElse(!current, "enabled", "disabled")
		m.Reply(F(chatID, "nothumb_status", locales.Arg{
			"cmd":    getCommand(m),
			"action": action,
		}))
		return tg.ErrEndGroup
	}

	value, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if current == value {
		action := utils.IfElse(!value, "enabled", "disabled")
		m.Reply(F(chatID, "nothumb_already", locales.Arg{
			"action": action,
		}))
		return tg.ErrEndGroup
	}

	if err := database.SetThumbnailsDisabled(chatID, value); err != nil {
		m.Reply(F(chatID, "nothumb_update_fail"))
		return tg.ErrEndGroup
	}

	action := utils.IfElse(!value, "enabled", "disabled")

	m.Reply(F(chatID, "nothumb_updated", locales.Arg{
		"action": action,
	}))
	return tg.ErrEndGroup
}
