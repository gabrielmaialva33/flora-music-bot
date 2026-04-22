package modules

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/playmode"] = `<i>Controla quem pode usar o comando /play nesse chat.</i>

<u>Uso:</u>
<b>/playmode [enable|disable]</b> — Define a restrição do modo play

<b>⚙️ Opções:</b>
• <b>enable</b> — Só admins e usuários autorizados podem tocar
• <b>disable</b> — Qualquer um pode tocar (padrão)`

	cmdDeleteHelp := `<i>Liga/desliga a exclusão automática dos comandos do bot nesse chat.</i>

<u>Uso:</u>
<b>/cmddelete [enable|disable]</b> — Define o status de exclusão dos comandos

<b>⚙️ Opções:</b>
• <b>enable</b> — Comandos são deletados depois de processados
• <b>disable</b> — Comandos ficam no chat (padrão)`

	helpTexts["/cmddelete"] = cmdDeleteHelp
	helpTexts["/commanddelete"] = cmdDeleteHelp
}

func playmodeHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	current, err := database.PlayModeAdminsOnly(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		statusKey := "playmode_status_everyone"
		if current {
			statusKey = "playmode_status_admins"
		}

		m.Reply(F(chatID, "playmode_help", locales.Arg{
			"status": F(chatID, statusKey),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	adminsOnly, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if err := database.SetPlayModeAdminsOnly(chatID, adminsOnly); err != nil {
		return err
	}

	successKey := "playmode_success_everyone"
	if adminsOnly {
		successKey = "playmode_success_admins"
	}

	m.Reply(F(chatID, successKey), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
}

func cmdDeleteHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()
	cmd := getCommand(m)

	current, err := database.CommandDelete(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		actionKey := "disabled"
		if current {
			actionKey = "enabled"
		}

		m.Reply(F(chatID, "cmddelete_status", locales.Arg{
			"cmd":    cmd,
			"action": F(chatID, actionKey),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if err := database.SetCommandDelete(chatID, enabled); err != nil {
		return err
	}

	actionKey := "disabled"
	if enabled {
		actionKey = "enabled"
	}

	m.Reply(F(chatID, "cmddelete_updated", locales.Arg{
		"action": F(chatID, actionKey),
	}), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
}
