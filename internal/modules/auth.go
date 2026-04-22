package modules

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/addauth"] = fmt.Sprintf(
		`<i>Dá permissão pra um usuário comum controlar o playback e outras features de admin sem precisar torná-lo admin do Telegram.</i>

<u>Uso:</u>
<b>/addauth [responda ao usuário]</b> — Adiciona um usuário respondendo a mensagem dele.
<b>/addauth &lt;user_id / username&gt;</b> — Adiciona direto pelo ID ou @username.

<b>⚙️ Observações:</b>
• Só <b>admins do chat</b> podem usar esse comando.
• Usuários autorizados podem controlar o playback com comandos tipo <code>/pause</code>, <code>/resume</code>, <code>/skip</code>, <code>/seek</code>, <code>/mute</code>, etc.
• 🤖 Bots não podem ser adicionados como usuários autorizados.
• 🔢 Você pode ter até <b>%d</b> usuários autorizados por chat.
• 👑 O <b>Dono do Bot</b>, o <b>Assistente</b> e todos os <b>Sudoers</b> já são <b>autorizados por padrão</b> — eles não aparecem na lista e não podem ser removidos.

Pra comandos relacionados, veja <code>/delauth</code> e <code>/authlist</code>.`,
		config.MaxAuthUsers,
	)

	helpTexts["/delauth"] = `<i>Revoga a permissão de um usuário que foi autorizado anteriormente a controlar o playback.</i>

<u>Uso:</u>
<b>/delauth [responda ao usuário]</b> — Remove respondendo a mensagem dele.
<b>/delauth &lt;user_id / username&gt; </b>— Remove pelo ID ou @username.

<b>⚙️ Observações:</b>
• Só <b>admins do chat</b> podem usar esse comando.
• Use pra tirar acesso de usuários que tão zoando.
• Pra ver quem tá autorizado agora, usa <code>/authlist</code>.`

	helpTexts["/authlist"] = `<u>Uso:</u>
<b>/authlist</b> - <i>Mostra todos os usuários autorizados a controlar o playback nesse chat.</i>

<b>⚙️ Observações:</b>
• Qualquer um do chat pode usar esse comando.
• Mostra só os usuários autorizados manualmente — o Dono, o Assistente e os Sudoers não aparecem na lista mas sempre têm autorização.`
}

func addAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	if au, _ := database.AuthorizedUsers(chatID); len(
		au,
	) >= config.MaxAuthUsers {
		m.Reply(F(chatID, "auth_limit_reached", locales.Arg{
			"limit": config.MaxAuthUsers,
		}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	// owner, bot, self, already auth, or admin — all treated the same
	if userID == config.OwnerID || userID == m.Client.Me().ID ||
		userID == m.SenderID() {
		m.Reply(F(chatID, "cannot_authorize_user"))
		return telegram.ErrEndGroup
	}

	if ok, _ := database.IsAuthorized(chatID, userID); ok {
		m.Reply(F(chatID, "already_authed"))
		return telegram.ErrEndGroup
	}

	if ok, _ := utils.IsChatAdmin(m.Client, chatID, userID); ok {
		m.Reply(F(chatID, "addauth_user_is_admin"))
		return telegram.ErrEndGroup
	}

	user, err := m.Client.GetUser(userID)
	if err != nil || user == nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), ""),
		}))
		return telegram.ErrEndGroup
	}

	if user.Bot {
		m.Reply(F(chatID, "addauth_bot_user"))
		return telegram.ErrEndGroup
	}

	if err := database.Authorize(chatID, userID); err != nil {
		m.Reply(F(chatID, "addauth_add_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	uname := utils.MentionHTML(user)
	if user.Username != "" {
		uname += " (@" + user.Username + ")"
	}

	if au, _ := database.AuthorizedUsers(chatID); len(au) > 0 {
		m.Reply(F(chatID, "addauth_success_with_count", locales.Arg{
			"user":  uname,
			"count": len(au),
			"limit": config.MaxAuthUsers,
		}))
	} else {
		m.Reply(F(chatID, "addauth_success", locales.Arg{
			"user": uname,
		}))
	}

	return telegram.ErrEndGroup
}

func delAuthHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	if m.Args() == "" && !m.IsReply() {
		m.Reply(F(chatID, "auth_no_user", locales.Arg{
			"cmd": getCommand(m),
		}))
		return telegram.ErrEndGroup
	}

	userID, err := utils.ExtractUser(m)
	if err != nil {
		m.Reply(F(chatID, "user_extract_fail", locales.Arg{
			"error": utils.IfElse(err != nil, err.Error(), "unknown error"),
		}))
		return telegram.ErrEndGroup
	}

	if ok, err := database.IsAuthorized(chatID, userID); !ok && err == nil {
		m.Reply(F(chatID, "del_auth_not_authorized", nil))
		return telegram.ErrEndGroup
	}

	user, _ := m.Client.GetUser(userID)

	if err := database.Unauthorize(chatID, userID); err != nil {
		m.Reply(F(chatID, "del_auth_remove_fail", locales.Arg{
			"error": err.Error(),
		}))
		return telegram.ErrEndGroup
	}

	var uname string
	if user != nil {
		uname = utils.MentionHTML(user)
		if user.Username != "" {
			uname += " (@" + user.Username + ")"
		}
	} else {
		uname = "User (<code>" + strconv.FormatInt(userID, 10) + "</code>)"
	}

	m.Reply(F(chatID, "del_auth_success", locales.Arg{
		"user": uname,
	}))
	return telegram.ErrEndGroup
}

func authListHandler(m *telegram.NewMessage) error {
	chatID := m.ChannelID()

	authUsers, err := database.AuthorizedUsers(chatID)
	if err != nil {
		m.Reply(F(chatID, "authlist_fetch_fail", locales.Arg{
			"error": err.Error(),
		}))
		return nil
	}

	if len(authUsers) == 0 {
		m.Reply(F(chatID, "authlist_empty", nil))
		return nil
	}

	statusMsg, err := m.Reply(F(chatID, "authlist_fetching", nil))
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(F(chatID, "authlist_header", nil) + "\n")

	for i, userID := range authUsers {
		user, err := m.Client.GetUser(userID)
		if err != nil || user == nil {
			sb.WriteString(F(chatID, "authlist_user_fail", locales.Arg{
				"index":   i + 1,
				"user_id": userID,
			}) + "\n")
			continue
		}

		uname := utils.MentionHTML(user)
		if user.Username != "" {
			uname += " (@" + user.Username + ")"
		}

		sb.WriteString(F(chatID, "authlist_user_entry", locales.Arg{
			"index":   i + 1,
			"user":    uname,
			"user_id": user.ID,
		}) + "\n")
	}

	sb.WriteString("\n" + F(chatID, "authlist_total", locales.Arg{
		"count": len(authUsers),
	}))

	utils.EOR(statusMsg, sb.String())
	return telegram.ErrEndGroup
}
