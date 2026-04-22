package modules

import (
	"errors"
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/reload"] = `<i>Recarrega o cache de admins e atualiza o estado do chat de voz.</i>

<u>Uso:</u>
<b>/reload</b> — Atualiza todos os dados em cache

<b>🔄 O que é recarregado:</b>
• Lista de admins do chat
• Status do chat de voz
• Status de presença do assistente
• Status de ban do assistente

<b>🔒 Proteção contra flood:</b>
• Usuários comuns: cooldown de 5 minutos
• Admins: cooldown de 2 minutos

<b>💡 Quando usar:</b>
• Depois de promover/rebaixar admins
• Problemas no chat de voz
• Problemas de permissão
• Bot tá se comportando errado

<b>⚠️ Observações:</b>
• Pode resetar o estado da room se precisar de permissões de admin
`
}

func reloadHandler(m *telegram.NewMessage) error {
	return handleReload(m, false)
}

func creloadHandler(m *telegram.NewMessage) error {
	return handleReload(m, true)
}

func handleReload(m *telegram.NewMessage, cplay bool) error {
	r, err := getEffectiveRoom(m, cplay)
	if err != nil {
		m.Reply(err.Error())
		return telegram.ErrEndGroup
	}

	chatID := m.ChannelID()
	actualChatID := r.ChatID()
	userID := m.SenderID()
	floodKey := fmt.Sprintf("reload:%d%d", actualChatID, userID)
	floodDuration := 5 * time.Minute

	if remaining := utils.GetFlood(floodKey); remaining > 0 {
		_, err := m.Reply(F(
			chatID,
			"flood_minutes",
			locales.Arg{
				"duration": formatDuration(int(remaining.Seconds())),
			},
		))
		return err
	}

	statusMsg, err := m.Reply(F(chatID, "reload_start"))
	if err != nil {
		return err
	}

	summary := ""

	admins, adminErr := utils.ReloadChatAdmin(m.Client, actualChatID)
	if adminErr != nil {
		summary += F(chatID, "reload_admin_cache_fail", locales.Arg{
			"error": adminErr.Error(),
		}) + "\n"
	} else {
		summary += F(chatID, "reload_admin_cache_ok") + "\n"
	}

	isAdmin := false
	if adminErr == nil {
		for _, id := range admins {
			if id == userID {
				isAdmin = true
				break
			}
		}
	}

	if isAdmin {
		floodDuration = 2 * time.Minute
	}
	utils.SetFlood(floodKey, floodDuration)

	cs, err := core.GetChatState(actualChatID)
	if err != nil {
		summary += F(chatID, "reload_assistant_fail", locales.Arg{
			"error": err.Error(),
		}) + "\n"
		utils.EOR(statusMsg, F(chatID, "reload_done", locales.Arg{
			"summary": summary,
		}))
		return nil
	}

	activeVC, vcErr := cs.IsActiveVC(true)
	if vcErr != nil {
		switch {
		case errors.Is(vcErr, core.ErrAdminPermissionRequired):
			summary += F(chatID, "reload_voice_admin_required") + "\n"
		default:
			summary += F(chatID, "reload_voice_fail", locales.Arg{
				"error": vcErr.Error(),
			}) + "\n"
		}
	} else if activeVC {
		summary += F(chatID, "reload_voice_active") + "\n"
	} else {
		summary += F(chatID, "reload_voice_inactive") + "\n"
	}

	banned, assErr := cs.IsAssistantBanned(true)
	if assErr != nil {
		switch {
		case errors.Is(assErr, core.ErrAdminPermissionRequired):
			summary += F(chatID, "reload_assistant_admin_required") + "\n"
		default:
			summary += F(chatID, "reload_assistant_fail", locales.Arg{
				"error": assErr.Error(),
			}) + "\n"
		}
	} else if banned {
		summary += F(chatID, "reload_assistant_banned") + "\n"
	} else {
		present, assErr2 := cs.IsAssistantPresent(false)
		if assErr2 != nil {
			switch {
			case errors.Is(assErr2, core.ErrAdminPermissionRequired):
				summary += F(chatID, "reload_assistant_admin_required") + "\n"
			default:
				summary += F(chatID, "reload_assistant_fail", locales.Arg{
					"error": assErr2.Error(),
				}) + "\n"
			}
		} else if present {
			summary += F(chatID, "reload_assistant_present") + "\n"
		} else {
			summary += F(chatID, "reload_assistant_not_present") + "\n"
		}
	}

	if isAdmin {
		if core.DeleteRoom(actualChatID) {
			summary += F(chatID, "reload_room_reset") + "\n"
		}
	}

	utils.EOR(statusMsg, F(chatID, "reload_done", locales.Arg{
		"summary": summary,
	}))

	return nil
}
