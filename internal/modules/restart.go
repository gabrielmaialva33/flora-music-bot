package modules

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/restart"] = `<i>Restart the bot process.</i>

<u>Usage:</u>
<b>/restart</b> — Restart bot

<b>⚙️ Behavior:</b>
• Stops all active rooms
• Notifies all active chats
• Restarts bot process
• Clears download cache

<b>🔒 Restrictions:</b>
• <b>Owner only</b> command

<b>⚠️ Warning:</b>
All playback will be interrupted. Bot will be offline for a few seconds.`
}

func handleRestart(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	statusMsg, err := m.Reply(F(chatID, "restart"))
	if err != nil {
		gologging.Error("Failed to send restart message: " + err.Error())
	}

	exePath, err := os.Executable()
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_exepath_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_symlink_fail", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	for chatID := range core.GetAllRooms() {
		core.DeleteRoom(chatID)
		m.Client.SendMessage(chatID, F(chatID, "restart_service", locales.Arg{
			"bot": utils.MentionHTML(m.Client.Me()),
		}))
		time.Sleep(time.Second)

	}

	utils.EOR(statusMsg, F(chatID, "restart_initiated"))

	_ = os.RemoveAll("downloads")
	_ = os.RemoveAll("cache")

	if err := syscall.Exec(exePath, os.Args, os.Environ()); err != nil {
		utils.EOR(statusMsg, F(chatID, "restart_fail", locales.Arg{
			"error": err.Error(),
		}))
	}

	return tg.ErrEndGroup
}
