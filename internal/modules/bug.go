package modules

import (
	"fmt"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["bug"] = `<i>Reporta um bug, problema ou comportamento inesperado direto pros devs do bot.</i>

<u>Uso:</u>
<b>/bug &lt;descrição&gt;</b> — Envia um report com uma explicação curta.
<b>Responda + /bug</b> — Reporta uma mensagem ou mídia específica como bug.

<b>🧠 Detalhes:</b>
Quando usado, o bot encaminha automaticamente seu report (e a mensagem respondida, se tiver) pro <b>dono</b> e pros <b>canais de log</b>.
Tem proteção contra flood — você só pode mandar um report a cada <b>5 minutos</b> por chat.

<b>⚠️ Observação:</b>
Os reports são logados só pra debug. Uso indevido (tipo spam) pode restringir seu acesso a esse comando.`
}

func bugHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()
	reason := m.Args()

	if reason == "" && !m.IsReply() {
		m.Reply(F(chatID, "bug_usage", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	// Flood control
	key := fmt.Sprintf("room:%d:%d", m.SenderID(), m.ChannelID())
	if remaining := utils.GetFlood(key); remaining > 0 {
		m.Reply(F(chatID, "flood_minutes", locales.Arg{
			"minutes": formatDuration(int(remaining.Seconds())),
		}))
		return telegram.ErrEndGroup
	}
	utils.SetFlood(key, 5*time.Minute)

	// Forward the replied message if any
	if m.IsReply() {
		if config.LoggerID != 0 {
			m.Client.Forward(config.LoggerID, m.Peer, []int32{m.ReplyID()})
		}
		if config.OwnerID != 0 {
			m.Client.Forward(config.OwnerID, m.Peer, []int32{m.ReplyID()})
		}
	}

	userMention := utils.MentionHTML(m.Sender)
	chatTitle := "Private Chat"
	if m.Channel != nil {
		chatTitle = m.Channel.Title
	}
	chatMention := fmt.Sprintf("<a href=\"%s\">%s</a>", m.Link(), chatTitle)

	reportMsg := F(chatID, "bug_report_format", locales.Arg{
		"user":    userMention,
		"user_id": m.Sender.ID,
		"chat":    chatMention,
		"chat_id": m.ChannelID(),
		"report":  reason,
	})

	// Send report to dev channels
	if config.LoggerID != 0 && (reason != "" || m.IsReply()) {
		m.Client.SendMessage(config.LoggerID, reportMsg)
	}
	if config.OwnerID != 0 && (reason != "" || m.IsReply()) {
		m.Client.SendMessage(config.OwnerID, reportMsg)
	}
	m.Reply(F(chatID, "bug_thanks"))
	return telegram.ErrEndGroup
}
