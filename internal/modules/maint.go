package modules

import (
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/maintenance"] = `<i>Liga/desliga o modo de manutenção.</i>

<u>Uso:</u>
<b>/maintenance</b> — Mostra o status atual
<b>/maintenance on [motivo]</b> — Ativa a manutenção
<b>/maintenance off</b> — Desativa a manutenção

<b>⚙️ Comportamento quando ativo:</b>
• Para todas as rooms ativas
• Bloqueia comandos de quem não for dono/sudoer
• Mostra mensagem de manutenção pros usuários

<b>🔒 Restrições:</b>
• Comando apenas pro <b>dono</b>

<b>💡 Exemplos:</b>
<code>/maintenance on Upgrade do servidor</code>
<code>/maintenance off</code>

<b>⚠️ Observações:</b>
• Dono e sudoers continuam podendo usar o bot
• Todas as rooms são destruídas quando ativado
• Usuários veem a mensagem de manutenção com o motivo`
	helpTexts["/maint"] = helpTexts["/maintenance"]
}

var maintCancel = struct {
	sync.Mutex
	cancel bool
}{}

func handleMaintenance(m *tg.NewMessage) error {
	args := strings.Fields(m.Text())
	chatID := m.ChannelID()
	current, err := database.IsMaintenanceEnabled()
	if err != nil {
		m.Reply(
			F(chatID, "maint_check_fail", locales.Arg{"error": err.Error()}),
		)
		return tg.ErrEndGroup
	}

	if len(args) < 2 {
		return showMaintenanceStatus(m, current)
	}

	enable, err := utils.ParseBool(args[1])
	if err != nil {
		m.Reply(F(chatID, "invalid_bool"))
		return tg.ErrEndGroup
	}

	reason := strings.Join(args[2:], " ")
	if current == enable {
		return handleSameMaintenanceState(m, enable, reason)
	}

	return applyMaintenanceState(m, enable, reason)
}

func showMaintenanceStatus(m *tg.NewMessage, current bool) error {
	chatID := m.ChannelID()
	reason, _ := database.MaintenanceReason()
	status := F(chatID, "disabled")
	if current {
		if reason != "" {
			status = F(
				chatID,
				"enabled_with_reason",
				locales.Arg{"reason": reason},
			)
		} else {
			status = F(chatID, "enabled")
		}
	}
	m.Reply(F(chatID, "maint_usage", locales.Arg{
		"cmd":    getCommand(m),
		"status": status,
	}))
	return tg.ErrEndGroup
}

func handleSameMaintenanceState(
	m *tg.NewMessage,
	enable bool,
	reason string,
) error {
	chatID := m.ChannelID()
	if !enable {
		m.Reply(F(chatID, "maint_already_disabled"))
		return tg.ErrEndGroup
	}

	oldReason, _ := database.MaintenanceReason()
	switch {
	case reason == oldReason:
		m.Reply(F(chatID, "maint_already_reason_same"))
	case reason == "" && oldReason != "":
		_ = database.SetMaintenance(true, "")
		m.Reply(F(chatID, "maint_reason_removed"))
	case reason != "" && reason != oldReason:
		_ = database.SetMaintenance(true, reason)
		m.Reply(
			F(chatID, "maint_reason_updated", locales.Arg{"reason": reason}),
		)
	default:
		m.Reply(F(chatID, "maint_already_enabled"))
	}
	return tg.ErrEndGroup
}

func applyMaintenanceState(m *tg.NewMessage, enable bool, reason string) error {
	chatID := m.ChannelID()
	database.SetMaintenance(enable, reason)
	gologging.InfoF(
		"User %d set maintenance: %v (reason: %s)",
		m.SenderID(),
		enable,
		reason,
	)

	maintCancel.Lock()
	maintCancel.cancel = !enable
	maintCancel.Unlock()

	if enable {
		go notifyMaintenanceStart(m.Client, reason)
		msgKey := "maint_enabled"
		args := locales.Arg{}
		if reason != "" {
			msgKey = "maint_enabled_reason"
			args["reason"] = reason
		}
		m.Reply(F(chatID, msgKey, args))
	} else {
		m.Reply(F(chatID, "maint_disabled"))
	}

	return tg.ErrEndGroup
}

func notifyMaintenanceStart(c *tg.Client, reason string) {
	for chatID := range core.GetAllRooms() {
		maintCancel.Lock()
		cancelled := maintCancel.cancel
		maintCancel.Unlock()

		if cancelled {
			break
		}

		core.DeleteRoom(chatID)
		msg := F(chatID, "maint_entering")
		if reason != "" {
			msg += "\n" + F(
				chatID,
				"maint_reason",
				locales.Arg{"reason": reason},
			)
		}
		c.SendMessage(chatID, msg)
		time.Sleep(time.Second)
	}
}
