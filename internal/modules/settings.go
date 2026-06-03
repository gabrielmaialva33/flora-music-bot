package modules

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
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

	helpTexts["/cleanmode"] = `<i>Ativa a limpeza temporizada de mensagens de comando/serviço nesse chat.</i>

<u>Uso:</u>
<b>/cleanmode [enable|disable]</b> — Define o status do clean mode

<b>⚙️ Comportamento:</b>
• <b>enable</b> — Respostas do bot e mensagens de comando são apagadas após um tempinho
• <b>disable</b> — Mantém as mensagens no chat (padrão)`

	helpTexts["/adminmode"] = `<i>Controla quem pode usar comandos de música de nível admin nesse chat.</i>

<u>Uso:</u>
<b>/adminmode [admin|adminauth|everyone]</b> — Define o acesso aos comandos admin

<b>⚙️ Opções:</b>
• <b>admin</b> — Só admins do chat podem usar comandos admin
• <b>adminauth</b> — Admins + usuários autorizados podem usar (padrão)
• <b>everyone</b> — Qualquer um pode usar comandos admin`
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

func cleanModeHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	current, err := database.CleanMode(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		m.Reply(
			cleanModeStatusText(chatID, current)+"\n\n"+F(chatID, "cleanmode_hint"),
			&tg.SendOptions{ParseMode: "HTML"},
		)
		return tg.ErrEndGroup
	}

	enabled, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if err := database.SetCleanMode(chatID, enabled); err != nil {
		return err
	}
	if !enabled {
		cleanScheduler.cancel(chatID)
	}

	m.Reply(
		cleanModeStatusText(chatID, enabled)+"\n\n"+F(chatID, "cleanmode_hint"),
		&tg.SendOptions{ParseMode: "HTML"},
	)
	return tg.ErrEndGroup
}

func adminModeHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	current, err := database.GetAdminMode(chatID)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		m.Reply(F(chatID, "adminmode_help", locales.Arg{
			"status": F(chatID, adminModeStatusKey(current)),
		}), &tg.SendOptions{ParseMode: "HTML"})
		return tg.ErrEndGroup
	}

	mode, ok := parseAdminMode(args[1])
	if !ok {
		m.Reply(F(chatID, "adminmode_invalid"))
		return tg.ErrEndGroup
	}

	if err := database.SetAdminMode(chatID, mode); err != nil {
		return err
	}

	m.Reply(F(chatID, "adminmode_updated", locales.Arg{
		"status": F(chatID, adminModeStatusKey(mode)),
	}), &tg.SendOptions{ParseMode: "HTML"})
	return tg.ErrEndGroup
}

func adminModeStatusKey(mode database.AdminMode) string {
	switch mode {
	case database.AdminModeAdminsOnly:
		return "adminmode_status_admin"
	case database.AdminModeEveryone:
		return "adminmode_status_everyone"
	default:
		return "adminmode_status_adminauth"
	}
}

func parseAdminMode(input string) (database.AdminMode, bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "admin", "admins", "adminonly", "adminsonly", "admins_only":
		return database.AdminModeAdminsOnly, true
	case "adminauth", "auth", "admin+auth", "dj", "admin_auth":
		return database.AdminModeAdminAuth, true
	case "everyone", "all":
		return database.AdminModeEveryone, true
	default:
		return "", false
	}
}

func settingsHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()
	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return err
	}

	title := "Chat"
	if m.Channel != nil {
		title = m.Channel.Title
	}

	kb := buildSettingsMarkup(chatID, settings)
	_, err = m.Reply(F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return err
}

func settingsCallbackHandler(cb *tg.CallbackQuery) error {
	chatID := cb.ChannelID()
	data := cb.DataString()
	parts := strings.Split(data, ":")
	title := "Chat"
	if cb.Channel != nil {
		title = cb.Channel.Title
	}

	if len(parts) < 2 {
		return nil
	}

	// Check permissions
	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil ||
		!isAdmin {
		cb.Answer(F(chatID, "only_admin_cb"), &tg.CallbackOptions{Alert: true})
		return nil
	}

	settings, err := database.GetChatSettings(chatID)
	if err != nil {
		return err
	}

	action := parts[1]
	if strings.HasPrefix(data, "info:") {
		cb.Answer(F(chatID, "settings_info_"+action), &tg.CallbackOptions{Alert: true})
		return nil
	}
	if action == "main" {
		kb := buildSettingsMarkup(chatID, settings)
		cb.Edit(F(chatID, "settings_main", locales.Arg{
			"title": title,
			"id":    chatID,
		}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
		return nil
	}
	if action == "langmenu" {
		return showSettingsLangMenu(cb, chatID)
	}
	switch action {
	case "playmode":
		settings.PlayModeAdminsOnly = !settings.PlayModeAdminsOnly
	case "adminmode":
		switch settings.AdminMode {
		case database.AdminModeAdminsOnly:
			settings.AdminMode = database.AdminModeAdminAuth
		case database.AdminModeAdminAuth:
			settings.AdminMode = database.AdminModeEveryone
		default:
			settings.AdminMode = database.AdminModeAdminsOnly
		}
	case "cmddelete":
		settings.CommandDelete = !settings.CommandDelete
	case "cleanmode":
		settings.CleanMode = !settings.CleanMode
		if !settings.CleanMode {
			cleanScheduler.cancel(chatID)
		}
	case "cleanduration":
		next := cleanModeDurationOptions[0]
		for i, v := range cleanModeDurationOptions {
			if v == settings.CleanModeDurationMins {
				next = cleanModeDurationOptions[(i+1)%len(cleanModeDurationOptions)]
				break
			}
		}
		settings.CleanModeDurationMins = next
	case "nothumb":
		settings.ThumbnailsDisabled = !settings.ThumbnailsDisabled
	}

	if err := database.UpdateChatSettings(settings); err != nil {
		return err
	}

	cb.Answer(F(chatID, "settings_updated"))
	kb := buildSettingsMarkup(chatID, settings)

	cb.Edit(F(chatID, "settings_main", locales.Arg{
		"title": title,
		"id":    chatID,
	}), &tg.SendOptions{ParseMode: "HTML", ReplyMarkup: kb})
	return nil
}

// showSettingsLangMenu renders the language picker inline, reusing the Flora
// `lang:<code>` callback flow, with a back button to the settings panel.
func showSettingsLangMenu(cb *tg.CallbackQuery, chatID int64) error {
	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}

	kb := tg.NewKeyboard()
	var btns []tg.KeyboardButton
	for _, l := range locales.GetAvailableLanguages() {
		name := locales.Get(l, "name", nil)
		if l == lang {
			name = "✔️ " + name
		}
		btns = append(btns, tg.Button.Data(name, "lang:"+l))
	}
	kb.NewColumn(2, btns...)
	kb.AddRow(tg.Button.Data(F(chatID, "settings_btn_back"), "set:main"))

	cb.Edit(F(chatID, "lang_select"), &tg.SendOptions{ReplyMarkup: kb.Build()})
	return nil
}

func buildSettingsMarkup(chatID int64, s *database.ChatSettings) *tg.ReplyInlineMarkup {
	kb := tg.NewKeyboard()

	// Admin Mode
	adminModeStatus := F(chatID, adminModeStatusKey(s.AdminMode))
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_adminmode"), "info:adminmode"),
		tg.Button.Data(adminModeStatus, "set:adminmode"),
	)

	// Play Mode
	playModeStatus := F(
		chatID,
		utils.IfElse(
			s.PlayModeAdminsOnly,
			"playmode_status_admins",
			"playmode_status_everyone",
		),
	)
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_playmode"), "info:playmode"),
		tg.Button.Data(playModeStatus, "set:playmode"),
	)

	// Cmd Delete
	cmdDeleteStatus := utils.IfElse(s.CommandDelete, "enabled", "disabled")
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_cmddelete"), "info:cmddelete"),
		tg.Button.Data(F(chatID, cmdDeleteStatus), "set:cmddelete"),
	)

	// Clean Mode
	cleanModeStatus := utils.IfElse(s.CleanMode, "enabled", "disabled")
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_cleanmode"), "info:cleanmode"),
		tg.Button.Data(F(chatID, cleanModeStatus), "set:cleanmode"),
	)

	cleanDuration := s.CleanModeDurationMins
	if cleanDuration <= 0 {
		cleanDuration = 15
	}
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_cleanduration"), "info:cleanduration"),
		tg.Button.Data(utils.IntToStr(cleanDuration)+"m", "set:cleanduration"),
	)

	// Thumbnails
	thumbStatus := utils.IfElse(!s.ThumbnailsDisabled, "enabled", "disabled")
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_nothumb"), "info:nothumb"),
		tg.Button.Data(F(chatID, thumbStatus), "set:nothumb"),
	)

	// Language
	kb.AddRow(
		tg.Button.Data(F(chatID, "settings_btn_lang"), "info:lang"),
		tg.Button.Data(F(chatID, "name"), "set:langmenu"),
	)

	kb.AddRow(tg.Button.Data(F(chatID, "CLOSE_BTN"), "close"))

	return kb.Build()
}
