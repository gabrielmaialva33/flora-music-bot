package modules

import (
	"fmt"
	"log"

	"github.com/Laky-64/gologging"
	"github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
)

type MsgHandlerDef struct {
	Pattern string
	Handler telegram.MessageHandler
	Filters []telegram.Filter
}

type CbHandlerDef struct {
	Pattern string
	Handler telegram.CallbackHandler
	Filters []telegram.Filter
}

// Construtores de MsgHandlerDef que fixam os conjuntos de filtros mais comuns,
// evitando repetir o slice literal em cada entrada.
func sgAuth(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{superGroupFilter, authFilter}}
}

func sgAdmin(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{superGroupFilter, adminFilter}}
}

func sg(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{superGroupFilter}}
}

func sudo(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{sudoOnlyFilter, ignoreChannelFilter}}
}

func ownerCh(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{ownerFilter, ignoreChannelFilter}}
}

func ignoreCh(pat string, h telegram.MessageHandler) MsgHandlerDef {
	return MsgHandlerDef{pat, h, []telegram.Filter{ignoreChannelFilter}}
}

var handlers = []MsgHandlerDef{
	{Pattern: "json", Handler: jsonHandle},
	{
		Pattern: "eval",
		Handler: evalHandle,
		Filters: []telegram.Filter{ownerFilter},
	},
	{
		Pattern: "ev",
		Handler: evalCommandHandler,
		Filters: []telegram.Filter{ownerFilter},
	},
	{
		Pattern: "(bash|sh)",
		Handler: shellHandle,
		Filters: []telegram.Filter{ownerFilter},
	},
	ownerCh("restart", handleRestart),

	ownerCh("(addsudo|addsudoer|sudoadd)", handleAddSudo),
	ownerCh(
		"(delsudo|delsudoer|sudodel|remsudo|rmsudo|sudorem|dropsudo|unsudo)",
		handleDelSudo,
	),
	ignoreCh("(sudoers|listsudo|sudolist)", handleGetSudoers),

	ownerCh("(blockuser|blacklistuser|blackuser|bluser)", handleBlockUser),
	ownerCh(
		"(unblockuser|unblacklistuser|unbluser|whitelistuser)",
		handleUnblockUser,
	),
	ownerCh("(blockchat|blacklistchat|blackchat|blchat)", handleBlockChat),
	ownerCh(
		"(unblockchat|unblacklistchat|unblackchat|whitechat|unblchat)",
		handleUnblockChat,
	),
	ownerCh("(blocked|blacklisted)", handleBlacklisted),

	sudo("(speedtest|spt)", sptHandle),

	ownerCh("(broadcast|gcast|bcast)", broadcastHandler),

	sudo("(ac|active|activevc|activevoice)", activeHandler),
	ownerCh("(maintenance|maint)", handleMaintenance),
	sudo("logger", handleLogger),
	sudo("autoleave", autoLeaveHandler),
	sudo("(log|logs)", logsHandler),

	ignoreCh("help", helpHandler),
	ignoreCh("ping", pingHandler),
	ignoreCh("start", startHandler),
	{
		Pattern: "stats",
		Handler: statsHandler,
		Filters: []telegram.Filter{ignoreChannelFilter, sudoOnlyFilter},
	},
	ignoreCh("bug", bugHandler),
	ignoreCh("privacy", privacyHandler),
	ignoreCh("(anime|a|tomato)", animeHandler),
	sgAuth("(lang|language)", langHandler),

	// SuperGroup & Admin Filters

	sg("stream", streamHandler),
	sgAuth("streamstop", streamStopHandler),
	sg("streamstatus", streamStatusHandler),
	{Pattern: "(rtmp|setrtmp)", Handler: setRTMPHandler},
	// play/cplay/vplay/fplay commands
	sg("play", playHandler),
	sgAuth("(fplay|playforce)", fplayHandler),
	sg("cplay", cplayHandler),
	sgAuth("(cfplay|fcplay|cplayforce)", cfplayHandler),
	sg("vplay", vplayHandler),
	sgAuth("(fvplay|vfplay|vplayforce)", fvplayHandler),
	sg("(vcplay|cvplay)", vcplayHandler),
	sgAuth("(fvcplay|fvcpay|vcplayforce)", fvcplayHandler),

	sgAuth("(speed|setspeed|speedup)", speedHandler),
	sgAuth("skip", skipHandler),
	sgAuth("pause", pauseHandler),
	sgAuth("resume", resumeHandler),
	sgAuth("replay", replayHandler),
	sgAuth("mute", muteHandler),
	sgAuth("unmute", unmuteHandler),
	sgAuth("seek", seekHandler),
	sgAuth("seekback", seekbackHandler),
	sgAuth("jump", jumpHandler),
	sg("(pos|position)", positionHandler),
	sg("queue", queueHandler),
	sgAuth("clear", clearHandler),
	sgAuth("remove", removeHandler),
	sgAuth("move", moveHandler),
	sgAuth("shuffle", shuffleHandler),
	sgAuth("(loop|setloop)", loopHandler),
	sgAuth("(end|stop)", stopHandler),
	sg("reload", reloadHandler),
	sgAuth("restore", restoreHandler),
	sgAdmin("addauth", addAuthHandler),
	sgAdmin("delauth", delAuthHandler),
	sg("authlist", authListHandler),

	// CPlay commands
	sgAuth("(cplay|cvplay)", cplayHandler),
	sgAuth("(cfplay|fcplay|cforceplay)", cfplayHandler),
	sgAuth("cpause", cpauseHandler),
	sgAuth("cresume", cresumeHandler),
	sgAuth("cmute", cmuteHandler),
	sgAuth("cunmute", cunmuteHandler),
	sgAuth("(cstop|cend)", cstopHandler),
	sgAuth("cqueue", cqueueHandler),
	sgAuth("cskip", cskipHandler),
	sgAuth("(cloop|csetloop)", cloopHandler),
	sgAuth("cseek", cseekHandler),
	sgAuth("cseekback", cseekbackHandler),
	sgAuth("cjump", cjumpHandler),
	sgAuth("cremove", cremoveHandler),
	sgAuth("cclear", cclearHandler),
	sgAuth("cmove", cmoveHandler),
	sgAuth("channelplay", channelPlayHandler),
	sgAuth("(cspeed|csetspeed|cspeedup)", cspeedHandler),
	sgAuth("creplay", creplayHandler),
	sgAuth("(cpos|cposition)", cpositionHandler),
	sgAuth("cshuffle", cshuffleHandler),
	sgAuth("creload", creloadHandler),
	sgAuth("crestore", crestoreHandler),

	sgAuth("(nothumb|nothumbs)", nothumbHandler),
	sgAdmin("playmode", playmodeHandler),
	sgAdmin("(cmddelete|commanddelete)", cmdDeleteHandler),
	sgAdmin("cleanmode", cleanModeHandler),
	sgAdmin("adminmode", adminModeHandler),
	sg("settings", settingsHandler),
}

var cbHandlers = []CbHandlerDef{
	{Pattern: "start", Handler: startCB},
	{Pattern: "help_cb", Handler: helpCB},
	{Pattern: "^lang:[a-z]", Handler: langCallbackHandler},
	{Pattern: `^help:(.+)`, Handler: helpCallbackHandler},

	{Pattern: "^close$", Handler: closeHandler},
	{Pattern: "^cancel$", Handler: cancelHandler},
	{Pattern: "^bcast_cancel$", Handler: broadcastCancelCB},

	{Pattern: `^room:(\w+)$`, Handler: roomHandle},
	{Pattern: `^anime:`, Handler: animeCB},
	{Pattern: "progress", Handler: emptyCBHandler},
	{Pattern: "^(set|info):", Handler: settingsCallbackHandler},
}

func Init(bot *telegram.Client, assistants *core.AssistantManager) {
	bot.UpdatesGetState()
	bot.Use(blacklistMessageMiddleware)
	assistants.ForEach(func(a *core.Assistant) {
		a.Client.UpdatesGetState()
	})

	for _, h := range handlers {
		bot.AddCommandHandler(h.Pattern, SafeMessageHandler(h.Handler), h.Filters...).
			SetGroup(100)
	}

	for _, h := range cbHandlers {
		bot.AddCallbackHandler(h.Pattern, WithBlacklistCallback(SafeCallbackHandler(h.Handler)), h.Filters...).
			SetGroup(90)
	}

	bot.On("edit:/eval", evalHandle).SetGroup(80)
	bot.On("edit:/ev", evalCommandHandler).SetGroup(80)

	bot.On("participant", handleParticipantUpdate).SetGroup(70)
	bot.On("action", handleChatAction).SetGroup(70)

	bot.AddActionHandler(handleActions).SetGroup(60)

	bot.AddRawHandler(&telegram.UpdateReadChannelOutbox{}, cleanModeReadHandler)
	cleanScheduler.start()

	assistants.ForEach(func(a *core.Assistant) {
		a.Ntg.OnStreamEnd(streamEndHandler)
	})

	go MonitorRooms()

	if is, _ := database.AutoLeave(); is {
		go startAutoLeave()
	}

	if config.SetCmds && config.OwnerID != 0 {
		go setBotCommands(bot)
	}

	cplayCommands := []string{
		"/cfplay", "/vcplay", "/fvcplay",
		"/cpause", "/cresume", "/cskip", "/cstop",
		"/cmute", "/cunmute", "/cseek", "/cseekback",
		"/cjump", "/cremove", "/cclear", "/cmove",
		"/cspeed", "/creplay", "/cposition", "/cshuffle",
		"/cloop", "/cqueue", "/creload",
		"/crestore",
	}

	for _, cmd := range cplayCommands {
		baseCmd := "/" + cmd[2:] // Remove 'c' prefix
		if baseHelp, exists := helpTexts[baseCmd]; exists {
			helpTexts[cmd] = fmt.Sprintf(`<i>Variante de channel play de %s</i>

<b>⚙️ Requer:</b>
Primeiro configura o canal usando: <code>/channelplay --set [channel_id]</code>

%s

<b>💡 Observação:</b>
Esse comando afeta o chat de voz do canal vinculado, não o grupo atual.`, baseCmd, baseHelp)
		}
	}
}

func setBotCommands(bot *telegram.Client) {
	// Set commands for normal users in private chats
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopeUsers{}, "", AllCommands.PrivateUserCommands); err != nil {
		gologging.Error("Failed to set PrivateUserCommands " + err.Error())
	}

	// Set commands for normal users in group chats
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopeChats{}, "", AllCommands.GroupUserCommands); err != nil {
		gologging.Error("Failed to set GroupUserCommands " + err.Error())
	}

	// Set commands for chat admins
	if _, err := bot.BotsSetBotCommands(
		&telegram.BotCommandScopeChatAdmins{},
		"",
		append(AllCommands.GroupUserCommands, AllCommands.GroupAdminCommands...),
	); err != nil {
		gologging.Error("Failed to set GroupAdminCommands " + err.Error())
	}

	// Set commands for sudo users in their private chat
	sudoers, err := database.Sudoers()
	if err != nil {
		log.Printf("Failed to get sudoers for setting commands: %v", err)
	} else {
		sudoCommands := append(AllCommands.PrivateUserCommands, AllCommands.PrivateSudoCommands...)
		for _, sudoer := range sudoers {
			if _, err := bot.BotsSetBotCommands(
				&telegram.BotCommandScopePeer{
					Peer: &telegram.InputPeerUser{UserID: sudoer, AccessHash: 0},
				},
				"",
				sudoCommands,
			); err != nil {
				gologging.Error("Failed to set PrivateSudoCommands " + err.Error())
			}
		}
	}

	ownerCommands := append(
		AllCommands.PrivateUserCommands,
		AllCommands.PrivateSudoCommands...,
	)
	ownerCommands = append(ownerCommands, AllCommands.PrivateOwnerCommands...)
	if _, err := bot.BotsSetBotCommands(&telegram.BotCommandScopePeer{
		Peer: &telegram.InputPeerUser{UserID: config.OwnerID, AccessHash: 0},
	}, "", ownerCommands); err != nil {
		gologging.Error("Failed to set PrivateOwnerCommands " + err.Error())
	}
}
