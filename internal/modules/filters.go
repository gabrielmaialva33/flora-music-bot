package modules

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/utils"
)

var (
	superGroupFilter    = tg.CustomFilter(filterSuperGroup)
	adminFilter         = tg.CustomFilter(filterChatAdmins)
	authFilter          = tg.CustomFilter(filterAuthUsers)
	ignoreChannelFilter = tg.CustomFilter(filterChannel)
	sudoOnlyFilter      = tg.CustomFilter(filterSudo)
	ownerFilter         = tg.CustomFilter(filterOwner)
)

func filterSuperGroup(m *tg.NewMessage) bool {
	if !filterChannel(m) {
		return false
	}

	switch m.ChatType() {
	case tg.EntityChat:
		// EntityChat can be basic group or supergroup — allow only supergroup
		if m.Channel != nil && !m.Channel.Broadcast {
			database.AddServedChat(m.ChannelID())
			return true // Supergroup
		}
		warnAndLeave(m.Client, m.ChannelID()) // Basic group → leave
		database.RemoveServedChat(m.ChannelID())
		return false

	case tg.EntityChannel:
		return false // Pure channel chat → ignore

	case tg.EntityUser:
		m.Reply(F(m.ChannelID(), "only_supergroup"))
		database.AddServedUser(m.ChannelID())
		return false // Private chat → warn
	}

	return false
}

func filterChatAdmins(m *tg.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err != nil || !isAdmin {
		m.Reply(F(m.ChannelID(), "only_admin"))
		return false
	}
	return true
}

func filterAuthUsers(m *tg.NewMessage) bool {
	isAdmin, err := utils.IsChatAdmin(m.Client, m.ChannelID(), m.SenderID())
	if err == nil && isAdmin {
		return true
	}

	isAuth, err := database.IsAuthorized(m.ChannelID(), m.SenderID())
	if err == nil && isAuth {
		return true
	}

	m.Reply(F(m.ChannelID(), "only_admin_or_auth"))
	return false
}

func filterSudo(m *tg.NewMessage) bool {
	is, _ := database.IsSudo(m.SenderID())

	if config.OwnerID == 0 || (m.SenderID() != config.OwnerID && !is) {
		if m.IsPrivate() ||
			strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
			m.Reply(F(m.ChannelID(), "only_sudo"))
		}
		return false
	}

	return true
}

func filterChannel(m *tg.NewMessage) bool {
	if _, ok := m.Message.FromID.(*tg.PeerChannel); ok {
		return false
	}
	return true
}

func filterOwner(m *tg.NewMessage) bool {
	if config.OwnerID == 0 || m.SenderID() != config.OwnerID {
		if m.IsPrivate() ||
			strings.HasSuffix(m.GetCommand(), m.Client.Me().Username) {
			m.Reply(F(m.ChannelID(), "only_owner"))
		}
		return false
	}
	return true
}
