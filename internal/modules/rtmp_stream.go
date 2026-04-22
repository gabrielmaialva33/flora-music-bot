package modules

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

var (
	rtmpStreams   = make(map[int64]*tg.RTMPStream)
	rtmpStreamsMu sync.RWMutex
)

func init() {
	helpTexts["stream"] = `<i>Inicia streaming RTMP ao vivo pro servidor configurado.</i>

<u>Uso:</u>
<b>/stream &lt;query/URL&gt;</b> — Começa a transmitir uma faixa
<b>/stream [responda a áudio/vídeo]</b> — Transmite a mídia respondida

<b>🎥 Features:</b>
• Streaming ao vivo pro seu servidor RTMP
• Suporta áudio e vídeo
• Suporte a fila (tipo /play)
• Monitoramento de status em tempo real

<b>⚙️ Setup necessário:</b>
Antes de usar esse comando, um admin precisa configurar o RTMP:
1. Abre o DM do bot
2. Manda: <code>/setrtmp &lt;chat_id&gt; &lt;rtmp_url&gt;</code>

<b>📝 Exemplo de setup:</b>
No DM do bot:
<code>/setrtmp -1001234567890 rtmps://dc5-1.rtmp.t.me/s/123:key</code>

Depois no seu chat:
<code>/stream never gonna give you up</code>

<b>⚠️ Observações importantes:</b>
• Streams RTMP têm delay de buffering de ~15-30s
• Setup SÓ funciona no DM do bot (por segurança)
• Usa <code>/streamstop</code> pra terminar o stream
• Só admins/autorizados podem controlar streams
• NÃO usamos a API RTMP do Telegram - você usa seu próprio servidor`

	helpTexts["streamstop"] = `<i>Para o stream RTMP atual.</i>

<u>Uso:</u>
<b>/streamstop</b> — Para o stream ativo

<b>⚠️ Observação:</b>
Só admins/autorizados podem parar streams.`

	helpTexts["streamstatus"] = `<i>Checa o status atual do stream RTMP.</i>

<u>Uso:</u>
<b>/streamstatus</b> — Mostra informações do stream

<b>📊 Mostra:</b>
• Estado do stream (tocando/parado)
• Posição atual
• Servidor RTMP (mascarado por segurança)
• Status da configuração`

	helpTexts["setrtmp"] = `<i>Configura o servidor de streaming RTMP (só no DM).</i>

<u>Uso:</u>
<b>/setrtmp &lt;chat_id&gt; &lt;rtmp_url&gt;</b> — Define RTMP pra um chat

<b>🔒 Segurança:</b>
• <b>Esse comando SÓ funciona no DM</b> (chat privado com o bot)
• NUNCA compartilha credenciais RTMP em grupos
• Credenciais são armazenadas com segurança no banco
• NÃO usamos a API RTMP do Telegram

<b>📋 Formato da URL:</b>
<code>rtmp://servidor/app/streamkey</code>
ou
<code>rtmps://servidor/s/streamkey</code>

<b>📝 Exemplos:</b>

<b>Chat de Voz do Telegram:</b>
1. Inicia o chat de voz no seu canal
2. O Telegram te dá: <code>rtmps://dc5-1.rtmp.t.me/s/123:key</code>
3. No DM do bot manda: <code>/setrtmp -1001234567890 rtmps://dc5-1.rtmp.t.me/s/123:key</code>

<b>Servidor RTMP Customizado:</b>
<code>/setrtmp -1001234567890 rtmp://live.example.com/stream/mykey</code>

<b>🔍 Como pegar o Chat ID:</b>
• Encaminha uma mensagem do chat pro @userinfobot
• Ou usa o comando <code>/id</code> no chat

<b>⚠️ Requisitos:</b>
• Você precisa ser admin do chat alvo
• O bot precisa ser membro do chat alvo
• Comando só funciona no chat privado do bot

<b>💡 Por que só no DM?</b>
Stream keys RTMP são tipo senhas. Configurar no DM evita vazamento acidental em grupos.`
}

// Get or create RTMP stream for chat
func getOrCreateRTMPStream(chatID int64) (*tg.RTMPStream, error) {
	rtmpStreamsMu.Lock()
	defer rtmpStreamsMu.Unlock()

	if stream, exists := rtmpStreams[chatID]; exists {
		return stream, nil
	}

	url, key, err := database.RTMP(chatID)
	if err != nil || url == "" || key == "" {
		return nil, fmt.Errorf(
			"RTMP not configured. Admin must use /setrtmp in bot DM first",
		)
	}

	stream, err := core.Bot.NewRTMPStream(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create RTMP stream: %w", err)
	}

	stream.SetLoopCount(0)
	stream.SetURL(url)
	stream.SetKey(key)

	stream.OnError(func(chatID int64, err error) {
		gologging.ErrorF("RTMP error in chat %d: %v", chatID, err)
		core.Bot.SendMessage(
			chatID,
			"⚠️ RTMP stream encountered an error. Check logs for details.",
		)
	})

	rtmpStreams[chatID] = stream
	return stream, nil
}

func streamHandler(m *tg.NewMessage) error {
	return handleStream(m, false)
}

func handleStream(m *tg.NewMessage, force bool) error {
	chatID := m.ChannelID()

	url, key, err := database.RTMP(chatID)
	if err != nil || url == "" || key == "" {
		m.Reply(F(chatID, "rtmp_not_configured", locales.Arg{
			"cmd": "/setrtmp",
		}))
		return tg.ErrEndGroup
	}

	parts := strings.SplitN(m.Text(), " ", 2)
	query := ""
	if len(parts) > 1 {
		query = strings.TrimSpace(parts[1])
	}

	if query == "" && !m.IsReply() {
		m.Reply(F(chatID, "no_song_query", locales.Arg{
			"cmd": getCommand(m),
		}))
		return tg.ErrEndGroup
	}

	stream, err := getOrCreateRTMPStream(chatID)
	if err != nil {
		m.Reply(F(chatID, "rtmp_init_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	if stream.State() == tg.StreamStatePlaying && !force {
		m.Reply(F(chatID, "rtmp_already_streaming"))
		return tg.ErrEndGroup
	}

	searchStr := ""
	if query != "" {
		searchStr = F(chatID, "searching_query", locales.Arg{
			"query": utils.EscapeHTML(query),
		})
	} else {
		searchStr = F(chatID, "searching")
	}

	replyMsg, err := m.Reply(searchStr)
	if err != nil {
		gologging.ErrorF("Failed to send searching message: %v", err)
		return tg.ErrEndGroup
	}

	tracks, err := safeGetTracks(m, replyMsg, chatID, false)
	if err != nil {
		utils.EOR(replyMsg, err.Error())
		return tg.ErrEndGroup
	}

	if len(tracks) == 0 {
		utils.EOR(replyMsg, F(chatID, "no_song_found"))
		return tg.ErrEndGroup
	}

	track := tracks[0]
	mention := utils.MentionHTML(m.Sender)
	track.Requester = mention

	// Download track
	downloadingText := F(chatID, "play_downloading_song", locales.Arg{
		"title": utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
	})
	replyMsg, _ = utils.EOR(replyMsg, downloadingText)

	ctx, cancel := context.WithCancel(context.Background())
	downloadCancels[chatID] = cancel
	defer func() {
		if _, ok := downloadCancels[chatID]; ok {
			delete(downloadCancels, chatID)
			cancel()
		}
	}()

	filePath, err := safeDownload(ctx, track, replyMsg, chatID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.EOR(replyMsg, F(chatID, "play_download_canceled", locales.Arg{
				"user": mention,
			}))
		} else {
			utils.EOR(replyMsg, F(chatID, "play_download_failed", locales.Arg{
				"title": utils.EscapeHTML(utils.ShortTitle(track.Title, 25)),
				"error": utils.EscapeHTML(err.Error()),
			}))
		}
		return tg.ErrEndGroup
	}

	// Start streaming
	utils.EOR(replyMsg, F(chatID, "rtmp_starting_stream"))

	if err := stream.Play(filePath); err != nil {
		utils.EOR(replyMsg, F(chatID, "rtmp_play_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	// Success message
	title := utils.EscapeHTML(utils.ShortTitle(track.Title, 25))
	msgText := F(chatID, "rtmp_now_streaming", locales.Arg{
		"url":      track.URL,
		"title":    title,
		"duration": formatDuration(track.Duration),
		"by":       mention,
	})

	opt := &tg.SendOptions{
		ParseMode: "HTML",
	}

	if track.Artwork != "" {
		opt.Media = utils.CleanURL(track.Artwork)
	}

	utils.EOR(replyMsg, msgText, opt)
	return tg.ErrEndGroup
}

// /streamstop - Stop RTMP stream
func streamStopHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	rtmpStreamsMu.RLock()
	stream, exists := rtmpStreams[chatID]
	rtmpStreamsMu.RUnlock()

	if !exists || stream.State() != tg.StreamStatePlaying {
		m.Reply(F(chatID, "rtmp_not_streaming"))
		return tg.ErrEndGroup
	}

	if err := stream.Stop(); err != nil {
		m.Reply(F(chatID, "rtmp_stop_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	m.Reply(F(chatID, "rtmp_stopped", locales.Arg{
		"user": utils.MentionHTML(m.Sender),
	}))

	return tg.ErrEndGroup
}

// /streamstatus - Check RTMP status
func streamStatusHandler(m *tg.NewMessage) error {
	chatID := m.ChannelID()

	// Check if RTMP is configured (without exposing credentials)
	url, _, err := database.RTMP(chatID)
	if err != nil || url == "" {
		m.Reply(F(chatID, "rtmp_not_configured", locales.Arg{
			"cmd": "/setrtmp",
		}))
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.RLock()
	stream, exists := rtmpStreams[chatID]
	rtmpStreamsMu.RUnlock()

	if !exists {
		// RTMP configured but not initialized yet
		m.Reply(F(chatID, "rtmp_configured_not_started", locales.Arg{
			"server": maskRTMPURL(url),
		}))
		return tg.ErrEndGroup
	}

	state := stream.State()
	pos := stream.CurrentPosition()

	var statusText string
	switch state {
	case tg.StreamStatePlaying:
		statusText = F(chatID, "rtmp_status_playing", locales.Arg{
			"position": formatDuration(int(pos.Seconds())),
			"server":   maskRTMPURL(url),
		})
	case tg.StreamStatePaused:
		statusText = F(chatID, "rtmp_status_paused", locales.Arg{
			"server": maskRTMPURL(url),
		})
	default:
		statusText = F(chatID, "rtmp_status_idle", locales.Arg{
			"server": maskRTMPURL(url),
		})
	}

	m.Reply(statusText)
	return tg.ErrEndGroup
}

// /setrtmp - Configure RTMP (DM only for security)
func setRTMPHandler(m *tg.NewMessage) error {
	if !filterChannel(m) {
		return tg.ErrEndGroup
	}

	m.Delete()

	switch m.ChatType() {
	case tg.EntityChat:
		m.Reply(F(m.ChannelID(), "rtmp_dm_only", locales.Arg{
			"cmd":          "/setrtmp",
			"bot_username": m.Client.Me().Username,
		}))
		return tg.ErrEndGroup
	case tg.EntityUser:
	default:
		return tg.ErrEndGroup
	}

	args := strings.Fields(m.Text())

	if len(args) < 3 {
		m.Reply(F(m.ChannelID(), "rtmp_setup_usage"))
		return tg.ErrEndGroup
	}

	cid := args[1]
	raw := args[2]

	idx := strings.LastIndex(raw, "/")
	if idx <= 0 || idx == len(raw)-1 {
		m.Reply(F(m.ChannelID(), "rtmp_parse_failed", locales.Arg{
			"error": "invalid RTMP format",
		}))
		return tg.ErrEndGroup
	}

	url := raw[:idx+1]
	key := raw[idx+1:]

	if url == "" || key == "" {
		m.Reply(F(m.ChannelID(), "rtmp_parse_failed", locales.Arg{
			"error": "empty url or key",
		}))
		return tg.ErrEndGroup
	}

	targetChatID, err := strconv.ParseInt(cid, 10, 64)
	if err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_invalid_chat_id"))
		return tg.ErrEndGroup
	}

	isAdmin, err := utils.IsChatAdmin(m.Client, targetChatID, m.SenderID())
	if err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_check_admin_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}
	if !isAdmin {
		m.Reply(F(m.ChannelID(), "rtmp_not_admin"))
		return tg.ErrEndGroup
	}

	if err := database.SetRTMP(targetChatID, url, key); err != nil {
		m.Reply(F(m.ChannelID(), "rtmp_init_failed", locales.Arg{
			"error": err.Error(),
		}))
		return tg.ErrEndGroup
	}

	rtmpStreamsMu.Lock()
	if stream, exists := rtmpStreams[targetChatID]; exists {
		stream.SetURL(url)
		stream.SetKey(key)
	}
	rtmpStreamsMu.Unlock()

	m.Reply(F(m.ChannelID(), "rtmp_configured_success", locales.Arg{
		"chat_id": targetChatID,
		"url":     url,
		"key":     maskKey(key),
	}))

	return tg.ErrEndGroup
}

func clearRTMPState(chatID int64) {
	rtmpStreamsMu.Lock()
	defer rtmpStreamsMu.Unlock()

	if stream, ok := rtmpStreams[chatID]; ok {
		_ = stream.Stop()
		delete(rtmpStreams, chatID)
	}
}

func maskRTMPURL(url string) string {
	if idx := strings.Index(url, "://"); idx != -1 {
		proto := url[:idx+3]
		rest := url[idx+3:]
		if len(rest) > 10 {
			return proto + rest[:10] + "***"
		}
	}
	return url
}

func maskKey(k string) string {
	l := len(k)
	if l <= 4 {
		return "****"
	}
	if l <= 8 {
		return k[:2] + "****" + k[l-2:]
	}
	return k[:4] + "****" + k[l-4:]
}
