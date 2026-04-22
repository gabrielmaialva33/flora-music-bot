package modules

import (
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
)

func init() {
	helpTexts["/help"] = fmt.Sprintf(`ℹ️ <b>Comando de Ajuda</b>
<i>Mostra a ajuda geral do bot ou informações detalhadas sobre um comando específico.</i>

<u>Uso:</u>
<code>/help</code> — Mostra o menu principal de ajuda.
<code>/help &lt;comando&gt;</code> — Mostra a ajuda pra um comando específico.

<b>💡 Dica:</b> Você pode ver a ajuda de qualquer comando direto adicionando a flag <code>-h</code> ou <code>--help</code>, ex: <code>/play -h</code>

<b>⚠️ Observação:</b> Alguns comandos são <b>restritos</b> a contextos específicos (tipo <b>Grupos</b>, <b>Admins</b>, <b>Sudoers</b> ou o <b>Dono</b>).
Se você tentar usar <code>-h</code> ou <code>--help</code> em um chat restrito ou PM, o bot pode não responder.
Pra ainda ver a ajuda desses comandos, usa o formato global:
<code>/help &lt;comando&gt;</code>

Pra mais infos, visita nosso <a href="%s">Chat de Suporte</a>.`, config.SupportChat)
}

func helpHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	if len(args) > 1 {
		cmd := args[1]
		if cmd != "pm_help" {
			if !strings.HasPrefix(cmd, "/") {
				cmd = "/" + cmd
			}
			return showHelpFor(m, cmd)
		}
	}

	if m.ChatType() != tg.EntityUser {
		m.Reply(
			F(m.ChannelID(), "help_private_only"),
			&tg.SendOptions{
				ReplyMarkup: core.GetGroupHelpKeyboard(m.ChannelID()),
			},
		)
		return tg.ErrEndGroup
	}

	m.Reply(
		F(m.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(m.ChannelID())},
	)
	return tg.ErrEndGroup
}

func helpCB(c *tg.CallbackQuery) error {
	c.Edit(
		F(c.ChannelID(), "help_main"),
		&tg.SendOptions{ReplyMarkup: core.GetHelpKeyboard(c.ChannelID())},
	)
	c.Answer("")
	return tg.ErrEndGroup
}

func helpCallbackHandler(c *tg.CallbackQuery) error {
	data := c.DataString()
	c.Answer("")
	if data == "" {
		return tg.ErrEndGroup
	}
	chatID := c.ChannelID()
	parts := strings.SplitN(data, ":", 2)
	if len(parts) < 2 {
		return tg.ErrEndGroup
	}

	var text string
	btn := core.GetBackKeyboard(chatID)

	switch parts[1] {
	case "admins":
		text = F(chatID, "help_admin")
	case "sudoers":
		text = F(chatID, "help_sudo")
	case "owner":
		text = F(chatID, "help_owner")
	case "public":
		text = F(chatID, "help_public")
	case "main":
		text = F(chatID, "help_main")
		btn = core.GetHelpKeyboard(chatID)
	}

	c.Edit(text, &tg.SendOptions{ReplyMarkup: btn})
	return tg.ErrEndGroup
}
