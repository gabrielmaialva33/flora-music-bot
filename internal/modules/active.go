package modules

import (
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/locales"
)

func init() {
	helpTexts["/active"] = `<i>Mostra todas as sessões de chat de voz ativas.</i>

<u>Uso:</u>
<b>/active</b> ou <b>/ac</b> — Lista os chats ativos

<b>📊 Informações exibidas:</b>
• Total de chats ativos
• Conexões NTGCalls ativas
• Sessões quebradas/velhas

<b>🔒 Restrições:</b>
• Apenas <b>sudoers</b>

<b>💡 Caso de uso:</b>
Monitorar o uso do bot e identificar problemas.`

	keys := []string{"/ac", "/activevc", "/activevoice"}
	for _, k := range keys {
		helpTexts[k] = helpTexts["/active"]
	}
}

func activeHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	allRooms := core.GetAllRooms()
	activeCount := len(allRooms)

	ntgChats := make(map[int64]struct{})

	core.Assistants.ForEach(func(a *core.Assistant) {
		if a == nil || a.Ntg == nil {
			return
		}
		for id := range a.Ntg.Calls() {
			ntgChats[id] = struct{}{}
		}
	})

	brokenCount := 0
	for id := range allRooms {
		if _, ok := ntgChats[id]; !ok {
			brokenCount++
		}
	}

	msg := F(chatID, "active_chats_info", locales.Arg{
		"count": activeCount,
	})

	if brokenCount > 0 {
		msg = F(chatID, "active_chats_info_with_broken", locales.Arg{
			"count":  activeCount,
			"broken": brokenCount,
		})
	}

	m.Reply(msg)
	return telegram.ErrEndGroup
}
