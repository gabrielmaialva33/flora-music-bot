package modules

import (
	"strings"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func langHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	lang, err := database.Language(chatID)
	if err != nil {
		lang = config.DefaultLang
	}

	kb := telegram.NewKeyboard()
	var btns []telegram.KeyboardButton
	for _, l := range locales.GetAvailableLanguages() {
		name := locales.Get(l, "name", nil)
		if l == lang {
			name = "✔️ " + name
		}
		btns = append(btns, telegram.Button.Data(name, "lang:"+l))
	}
	kb.NewColumn(2, btns...)

	_, err = m.Reply(
		F(chatID, "lang_select"),
		&telegram.SendOptions{ReplyMarkup: kb.Build()},
	)
	if err != nil {
		return err
	}
	return telegram.ErrEndGroup
}

func langCallbackHandler(cb *telegram.CallbackQuery) error {
	data := cb.DataString()
	opt := &telegram.CallbackOptions{Alert: true}
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		cb.Answer("⚠️ Invalid data.", opt)
		return telegram.ErrEndGroup
	}
	lang := parts[1]

	chatID := cb.ChannelID()
	if isAdmin, err := utils.IsChatAdmin(cb.Client, chatID, cb.SenderID); err != nil ||
		!isAdmin {
		cb.Answer(F(chatID, "only_admin_or_auth_cb"), opt)
		return telegram.ErrEndGroup
	}

	currentLang, _ := database.Language(chatID)

	if lang == currentLang {
		cb.Answer(F(chatID, "lang_same"), opt)
		return telegram.ErrEndGroup
	}

	langName := locales.Get(lang, "name", nil)

	if err := database.SetLanguage(chatID, lang); err != nil {
		gologging.ErrorF("SetChatLanguage error: %v", err)
		cb.Answer(F(chatID, "lang_fail"), opt)
		return telegram.ErrEndGroup
	}

	msg := F(chatID, "lang_success", locales.Arg{"lang_name": langName})
	cb.Answer(msg, opt)
	cb.Edit(msg)
	return telegram.ErrEndGroup
}
