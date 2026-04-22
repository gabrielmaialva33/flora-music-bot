package modules

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var (
	autoLeaveMu      sync.Mutex
	autoLeaveRunning bool
	limit            = 50
)

func init() {
	helpTexts["autoleave"] = fmt.Sprintf(
		`<i>Faz o assistente sair automaticamente de chats inativos ou desnecessários a cada 10 minutos.</i>

<u>Uso:</u>
<b>/autoleave </b>— Mostra o status atual do auto-leave (ativado/desativado).
<b>/autoleave enable</b> — Ativa o modo auto-leave.
<b>/autoleave disable</b> — Desativa o modo auto-leave.

<b>🧠 Detalhes:</b>
Quando ativado, o bot checa todos os grupos/canais onde tá a cada <b>10 minutos</b> e sai de até <b>%d chats por ciclo</b> que não estão na lista de rooms ativas.

<b>⚠️ Restrições:</b>
Esse comando só pode ser usado pelo <b>dono</b> ou <b>sudoers</b>.`,
		limit,
	)
}

func autoLeaveHandler(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()

	currentState, err := database.AutoLeave()
	if err != nil {
		m.Reply(F(chatID, "autoleave_fetch_fail"))
		return tg.ErrEndGroup
	}

	status := F(chatID, utils.IfElse(currentState, "enabled", "disabled"))

	if len(args) < 2 {
		m.Reply(F(chatID, "autoleave_status", locales.Arg{
			"cmd":    getCommand(m),
			"action": status,
		}))
		return tg.ErrEndGroup
	}

	newState, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	if newState == currentState {
		m.Reply(F(chatID, "autoleave_already", locales.Arg{
			"action": status,
		}))
		return tg.ErrEndGroup
	}

	if err := database.SetAutoLeave(newState); err != nil {
		m.Reply(F(chatID, "autoleave_update_fail"))
		return tg.ErrEndGroup
	}

	newStatus := F(chatID, utils.IfElse(newState, "enabled", "disabled"))
	m.Reply(F(chatID, "autoleave_updated", locales.Arg{
		"action": newStatus,
	}))

	autoLeaveMu.Lock()
	if newState {
		if !autoLeaveRunning {
			go startAutoLeave()
		}
	} else {
		autoLeaveRunning = false
	}
	autoLeaveMu.Unlock()

	return tg.ErrEndGroup
}

func startAutoLeave() {
	autoLeaveMu.Lock()
	if autoLeaveRunning {
		autoLeaveMu.Unlock()
		return
	}
	autoLeaveRunning = true
	autoLeaveMu.Unlock()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		autoLeaveMu.Lock()
		if !autoLeaveRunning {
			autoLeaveMu.Unlock()
			return
		}
		autoLeaveMu.Unlock()

		activeRooms := core.GetAllRooms()
		core.Assistants.ForEach(func(a *core.Assistant) {
			if a == nil || a.Client == nil {
				return
			}
			go autoLeaveAssistant(a, activeRooms)
		})

		<-ticker.C
	}
}

func autoLeaveAssistant(
	ass *core.Assistant,
	activeRooms map[int64]*core.RoomState,
) {
	leaveCount := 0
	err := ass.Client.IterDialogs(func(d *tg.TLDialog) error {
		if d.IsUser() {
			return nil
		}
		chatID := d.GetChannelID()

		if chatID == 0 || chatID == config.LoggerID ||
			d.GetID() == config.LoggerID {
			return nil
		}

		if _, ok := activeRooms[chatID]; ok {
			return nil
		}

		if err := ass.Client.LeaveChannel(chatID); err != nil {
			if wait := tg.GetFloodWait(err); wait > 0 {
				gologging.ErrorF(
					"FloodWait detected (%ds). Sleeping...", wait,
				)
				time.Sleep(time.Duration(wait) * time.Second)
				return nil
			}

			if strings.Contains(err.Error(), "USER_NOT_PARTICIPANT") ||
				strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
				return nil
			}

			gologging.WarnF(
				"AutoLeave (Assistant %d) failed to leave %d: %v",
				ass.Index, chatID, err,
			)
			return nil
		}

		leaveCount++
		gologging.InfoF(
			"AutoLeave: Assistant %d left %d (%d/%d)",
			ass.Index, chatID, leaveCount, limit,
		)

		time.Sleep(3 * time.Second)

		if leaveCount >= limit {
			return tg.ErrStopIteration
		}

		return nil
	}, &tg.DialogOptions{
		Limit: 0,
	})

	if err != nil && err != tg.ErrStopIteration {
		gologging.WarnF(
			"AutoLeave: IterDialogs error (assistant %d): %v",
			ass.Index, err,
		)
	}
}
