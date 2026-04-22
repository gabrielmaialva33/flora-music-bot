package modules

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/nothumb"] = `<i>Toggle thumbnail/artwork display in playback messages.</i>

<u>Usage:</u>
<b>/nothumb</b> — Show current status
<b>/nothumb [enable|disable]</b> — Change setting

<b>⚙️ Behavior:</b>
• <b>Disabled (default):</b> Shows track artwork/thumbnail
• <b>Enabled:</b> Hides artwork, text-only messages

<b>💡 Examples:</b>
<code>/nothumb enable</code> — Disable thumbnails
<code>/nothumb disable</code> — Enable thumbnails

<b>⚠️ Note:</b>
This setting affects all future playback messages in this chat.`
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
