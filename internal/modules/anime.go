package modules

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/tomato"
)

// ---------- in-memory caches ----------
//
// We cache episode thumbnails and anime cover URLs so the now-playing
// screen can repaint with a rich visual without hitting the API again.
// Both maps are small (a few dozen entries per active chat) and are
// never flushed — memory cost is negligible.
var (
	epThumbCache    = make(map[int]string)
	epThumbMu       sync.RWMutex
	animeCoverCache = make(map[int]string)
	animeCoverMu    sync.RWMutex
	animeNameCache  = make(map[int]string)
	animeNameMu     sync.RWMutex
)

func cacheAnimeName(id int, name string) {
	if id == 0 || name == "" {
		return
	}
	animeNameMu.Lock()
	animeNameCache[id] = name
	animeNameMu.Unlock()
}

func lookupAnimeName(id int) string {
	animeNameMu.RLock()
	defer animeNameMu.RUnlock()
	return animeNameCache[id]
}

// preloadAnimeNames fetches every anime's name in parallel (capped)
// and caches them. Used to turn "#1234" into real titles before
// rendering a rail.
func preloadAnimeNames(ids []int) {
	const maxConcurrent = 6
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for _, id := range ids {
		if id == 0 || lookupAnimeName(id) != "" {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()
			a, err := tomato.Default().Anime(id)
			if err != nil {
				gologging.DebugF("anime: preload %d failed: %v", id, err)
				return
			}
			cacheAnimeName(id, a.AnimeDetails.AnimeName)
			cover := a.AnimeDetails.AnimeCapeURL
			if cover == "" {
				cover = a.AnimeDetails.AnimeBannerURL
			}
			cacheAnimeCover(id, cover)
		}(id)
	}
	wg.Wait()
}

func cacheEpisodeThumb(id int, thumb string) {
	if id == 0 || thumb == "" {
		return
	}
	epThumbMu.Lock()
	epThumbCache[id] = thumb
	epThumbMu.Unlock()
}

func lookupEpisodeThumb(id int) string {
	epThumbMu.RLock()
	defer epThumbMu.RUnlock()
	return epThumbCache[id]
}

func cacheAnimeCover(id int, cover string) {
	if id == 0 || cover == "" {
		return
	}
	animeCoverMu.Lock()
	animeCoverCache[id] = cover
	animeCoverMu.Unlock()
}

func lookupAnimeCover(id int) string {
	animeCoverMu.RLock()
	defer animeCoverMu.RUnlock()
	return animeCoverCache[id]
}

// friendlyErr translates a tomato client error into a user-facing
// Telegram message. The raw error is logged separately — this never
// leaks the upstream URL or HTTP details to end users.
func friendlyErr(err error) string {
	switch {
	case errors.Is(err, tomato.ErrNotConfigured):
		return "⚙️ <b>TomatoAnimes não configurado.</b>\n" +
			"Define <code>BETOMATO_TOKEN</code> no <code>.env</code> e reinicia o bot."
	case errors.Is(err, tomato.ErrAuth):
		return "🔒 <b>Sessão do TomatoAnimes expirou.</b>\n" +
			"Atualiza o <code>BETOMATO_TOKEN</code> no <code>.env</code> e reinicia."
	case errors.Is(err, tomato.ErrNotFound):
		return "🔍 <b>Conteúdo não encontrado.</b>"
	case errors.Is(err, tomato.ErrTimeout):
		return "⏱ <b>O servidor demorou pra responder.</b>\nTenta de novo em instantes."
	case errors.Is(err, tomato.ErrUnavailable):
		return "⚠️ <b>Servidor indisponível.</b>\nVolta mais tarde."
	case errors.Is(err, tomato.ErrTransport):
		return "🌐 <b>Sem conexão com o servidor.</b>\nCheca a rede e tenta de novo."
	case errors.Is(err, tomato.ErrProtocol):
		return "🤖 <b>Resposta inesperada do servidor.</b>\nTenta de novo."
	}
	return "❌ <b>Algo deu errado.</b> Tenta de novo."
}

// ---------- help ----------

func init() {
	helpTexts["/anime"] = `<i>Navega e toca animes do TomatoAnimes direto no chat de voz.</i>

<u>Uso:</u>
<b>/anime</b> — Abre a tela inicial com "Em alta" e "Novos episódios"
<b>/anime &lt;busca&gt;</b> — Procura um anime/mangá por nome

<b>🎬 Fluxo:</b>
1. Escolhe um anime nos resultados
2. Seleciona a temporada desejada
3. Escolhe o episódio
4. O bot te entrega o comando <code>/vplay</code> pronto — clica no texto pra copiar e manda no chat de voz

<b>🔑 Requisitos:</b>
• Variável <code>BETOMATO_TOKEN</code> configurada no <code>.env</code>
• Chat de vídeo ativo no grupo (pro <code>/vplay</code>)

<b>💡 Dica:</b> usa <code>/a</code> como atalho.`
	helpTexts["/a"] = helpTexts["/anime"]
	helpTexts["/tomato"] = helpTexts["/anime"]
}

// ---------- entry handler ----------

func animeHandler(m *tg.NewMessage) error {
	if !tomato.Default().Configured() {
		m.Reply(
			"🍅 <b>TomatoAnimes não configurado</b>\n\n"+
				"Define a variável <code>BETOMATO_TOKEN</code> no teu <code>.env</code> "+
				"e reinicia o bot pra usar o <code>/anime</code>.",
			&tg.SendOptions{ParseMode: tg.HTML},
		)
		return tg.ErrEndGroup
	}

	query := strings.TrimSpace(m.Args())
	if query == "" {
		return renderHome(m, nil)
	}
	return renderSearch(m, nil, query, 0)
}

// ---------- callback router ----------

func animeCB(cb *tg.CallbackQuery) error {
	data := cb.DataString()
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		cb.Answer("")
		return tg.ErrEndGroup
	}

	switch parts[1] {
	case "h": // anime:h — home
		cb.Answer("⏳ Carregando…")
		return renderHome(nil, cb)
	case "rail": // anime:rail:<b64title>
		if len(parts) < 3 {
			cb.Answer("🤔 Categoria inválida")
			return tg.ErrEndGroup
		}
		title, _ := decodeB64(parts[2])
		cb.Answer("🎞 Carregando títulos…")
		return renderRail(cb, title)
	case "s": // anime:s:<b64query>:<page>
		if len(parts) < 4 {
			cb.Answer("🤔 Busca inválida")
			return tg.ErrEndGroup
		}
		query, _ := decodeB64(parts[2])
		page, _ := strconv.Atoi(parts[3])
		cb.Answer("🔎 Buscando…")
		return renderSearch(nil, cb, query, page)
	case "v": // anime:v:<animeID>
		id, _ := strconv.Atoi(parts[2])
		cb.Answer("🎬 Abrindo…")
		return renderAnime(cb, id)
	case "ep": // anime:ep:<animeID>:<seasonID>:<page>
		if len(parts) < 5 {
			cb.Answer("🤔 Temporada inválida")
			return tg.ErrEndGroup
		}
		animeID, _ := strconv.Atoi(parts[2])
		seasonID, _ := strconv.Atoi(parts[3])
		page, _ := strconv.Atoi(parts[4])
		cb.Answer("📺 Carregando episódios…")
		return renderEpisodes(cb, animeID, seasonID, page)
	case "p": // anime:p:<episodeID> — auto-play in best quality
		id, _ := strconv.Atoi(parts[2])
		return performPlay(cb, id, "best")
	case "pq": // anime:pq:<episodeID>:<fhd|shd>
		if len(parts) < 4 {
			cb.Answer("🤔 Qualidade inválida")
			return tg.ErrEndGroup
		}
		id, _ := strconv.Atoi(parts[2])
		return performPlay(cb, id, parts[3])
	case "back": // anime:back:<animeID>
		id, _ := strconv.Atoi(parts[2])
		cb.Answer("🔙 Voltando…")
		return renderAnime(cb, id)
	case "x": // anime:x — close
		cb.Answer("")
		cb.Delete()
		return tg.ErrEndGroup
	case "noop": // anime:noop — pagination counter / dimmed arrows
		cb.Answer("")
		return tg.ErrEndGroup
	}

	cb.Answer("")
	return tg.ErrEndGroup
}

// ---------- renderers ----------

// sectionIcon picks a nice emoji for each feed rail title.
func sectionIcon(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "alta"):
		return "🔥"
	case strings.Contains(lower, "recém") || strings.Contains(lower, "novo"):
		return "🆕"
	case strings.Contains(lower, "dublagem") || strings.Contains(lower, "dublado"):
		return "🗣"
	case strings.Contains(lower, "aventura"):
		return "🗺"
	case strings.Contains(lower, "comédia") || strings.Contains(lower, "comedia"):
		return "😂"
	case strings.Contains(lower, "romance"):
		return "❤️"
	case strings.Contains(lower, "slice"):
		return "🌿"
	case strings.Contains(lower, "recomen"):
		return "✨"
	case strings.Contains(lower, "curtido"):
		return "👍"
	case strings.Contains(lower, "talvez"):
		return "🎯"
	}
	return "🎞"
}

// renderHome shows a category picker built from the feed rails.
// Each rail becomes a button; clicking it opens renderRail with
// preloaded anime names.
func renderHome(m *tg.NewMessage, cb *tg.CallbackQuery) error {
	feed, err := tomato.Default().Feed()
	if err != nil {
		gologging.ErrorF("anime: feed failed: %v", err)
		return replyErr(m, cb, friendlyErr(err))
	}

	var b strings.Builder
	b.WriteString("🍅 <b>TomatoAnimes</b>\n")
	b.WriteString("<i>Escolha uma categoria ou busque um anime.</i>\n\n")
	b.WriteString("<b>🔎 Pra buscar:</b> <code>/anime &lt;nome&gt;</code>")

	kb := tg.NewKeyboard()

	// 2-column grid of rails that actually have content.
	row := make([]tg.KeyboardButton, 0, 2)
	for _, sec := range feed.Data {
		if (sec.Type != 3 && sec.Type != 5) || len(sec.Data) == 0 || sec.Title == "" {
			continue
		}
		label := fmt.Sprintf("%s %s", sectionIcon(sec.Title), truncate(sec.Title, 22))
		row = append(row, tg.Button.Data(
			label,
			"anime:rail:"+encodeB64(sec.Title),
		))
		if len(row) == 2 {
			kb.AddRow(row...)
			row = nil
		}
	}
	if len(row) > 0 {
		kb.AddRow(row...)
	}

	kb.AddRow(
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(m, cb, b.String(), "", kb.Build())
}

// renderRail shows all animes from a given feed rail (identified by its
// title). Preloads names in parallel so the buttons display real titles.
func renderRail(cb *tg.CallbackQuery, title string) error {
	feed, err := tomato.Default().Feed()
	if err != nil {
		gologging.ErrorF("anime: feed failed: %v", err)
		return replyErr(nil, cb, friendlyErr(err))
	}

	var rail *tomato.FeedSection
	for i := range feed.Data {
		if feed.Data[i].Title == title {
			rail = &feed.Data[i]
			break
		}
	}
	if rail == nil || len(rail.Data) == 0 {
		return replyErr(nil, cb, "😕 Categoria não encontrada.")
	}

	// Collect up to 12 anime IDs and preload their names in parallel.
	ids := make([]int, 0, 12)
	for _, a := range rail.Data {
		if a.AnimeID == 0 {
			continue
		}
		ids = append(ids, a.AnimeID)
		if len(ids) >= 12 {
			break
		}
	}
	preloadAnimeNames(ids)

	var b strings.Builder
	fmt.Fprintf(
		&b,
		"%s <b>%s</b>\n<i>Escolha um anime pra abrir os detalhes.</i>\n",
		sectionIcon(rail.Title),
		html.EscapeString(rail.Title),
	)

	// 2-column grid of anime buttons with real names (cache-backed).
	kb := tg.NewKeyboard()
	row := make([]tg.KeyboardButton, 0, 2)
	for _, id := range ids {
		name := lookupAnimeName(id)
		if name == "" {
			name = fmt.Sprintf("Anime #%d", id)
		}
		row = append(row, tg.Button.Data(
			truncate(name, 22),
			fmt.Sprintf("anime:v:%d", id),
		))
		if len(row) == 2 {
			kb.AddRow(row...)
			row = nil
		}
	}
	if len(row) > 0 {
		kb.AddRow(row...)
	}

	kb.AddRow(
		tg.Button.Data("🏠 Início", "anime:h"),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(nil, cb, b.String(), "", kb.Build())
}

// renderSearch shows search results.
func renderSearch(
	m *tg.NewMessage,
	cb *tg.CallbackQuery,
	query string,
	page int,
) error {
	if query == "" {
		return replyErr(m, cb, "🤔 Busca vazia")
	}
	res, err := tomato.Default().Search(query, page)
	if err != nil {
		gologging.ErrorF("anime: search %q failed: %v", query, err)
		return replyErr(m, cb, friendlyErr(err))
	}
	if len(res.Result) == 0 {
		return replyErr(
			m, cb,
			fmt.Sprintf(
				"😕 Nenhum resultado pra <b>%s</b>.",
				html.EscapeString(query),
			),
		)
	}

	var b strings.Builder
	fmt.Fprintf(
		&b,
		"🔎 <b>Resultados pra:</b> <code>%s</code>\n",
		html.EscapeString(query),
	)
	b.WriteString("<i>Toque num item pra abrir.</i>\n\n")

	kb := tg.NewKeyboard()
	count := 0
	for _, r := range res.Result {
		// Only anime entries (streaming isn't supported for mangas)
		if r.Type != "anime" {
			continue
		}
		if count >= 10 {
			break
		}
		count++

		fmt.Fprintf(
			&b,
			"<b>%d.</b> <b>%s</b>  <i>(%s)</i>\n  🎞 %d eps • %s\n\n",
			count,
			html.EscapeString(truncate(r.Name, 60)),
			html.EscapeString(r.Date),
			r.Episodes,
			html.EscapeString(truncate(r.Tags, 50)),
		)
		kb.AddRow(
			tg.Button.Data(
				fmt.Sprintf("▶️ %d. %s", count, truncate(r.Name, 35)),
				fmt.Sprintf("anime:v:%d", r.ID),
			),
		)
	}

	// pagination row — API doesn't expose total pages, so we assume
	// "there's more" whenever the response is full. Dim the forward arrow
	// otherwise. The counter is approximate.
	hasMore := len(res.Result) >= 40
	approxTotal := page + 1
	if hasMore {
		approxTotal = page + 2
	}
	kb.AddRow(paginationRow(
		page, approxTotal,
		func(p int) string {
			return fmt.Sprintf("anime:s:%s:%d", encodeB64(query), p)
		},
	)...)

	kb.AddRow(
		tg.Button.Data("🏠 Início", "anime:h"),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(m, cb, b.String(), "", kb.Build())
}

// renderAnime shows the anime detail with seasons.
// The screen is repainted as a photo message (banner) with caption so
// the user gets a cinematic card.
func renderAnime(cb *tg.CallbackQuery, animeID int) error {
	a, err := tomato.Default().Anime(animeID)
	if err != nil {
		gologging.ErrorF("anime: detail %d failed: %v", animeID, err)
		return replyErr(nil, cb, friendlyErr(err))
	}

	d := a.AnimeDetails
	// Cache cover URL for the now-playing screen to reuse later.
	cover := d.AnimeBannerURL
	if cover == "" {
		cover = d.AnimeCapeURL
	}
	if cover == "" {
		cover = d.AnimeCoverURL
	}
	cacheAnimeCover(animeID, cover)

	var b strings.Builder
	fmt.Fprintf(
		&b, "🎬 <b>%s</b>  <i>(%s)</i>\n\n",
		html.EscapeString(d.AnimeName),
		html.EscapeString(d.AnimeDate),
	)
	if d.AnimeGenre != "" {
		fmt.Fprintf(&b, "🏷 <b>Gêneros:</b> %s\n", html.EscapeString(d.AnimeGenre))
	}
	if d.AnimeParentalRating != "" {
		fmt.Fprintf(&b, "🔞 <b>Classificação:</b> %s\n", html.EscapeString(d.AnimeParentalRating))
	}
	fmt.Fprintf(&b, "🎞 <b>Episódios:</b> %d", d.AnimeEpisodes)
	if d.DubAvailable {
		b.WriteString("  •  🗣 Dublagem ✅")
	}
	b.WriteString("\n")
	if d.AnimeDescription != "" {
		fmt.Fprintf(
			&b, "\n<i>%s</i>\n",
			html.EscapeString(truncate(d.AnimeDescription, 380)),
		)
	}
	b.WriteString("\n<b>📅 Escolha a temporada:</b>")

	// Seasons in 2-column grid. Dubbed and legendada com ícones distintos.
	kb := tg.NewKeyboard()
	row := make([]tg.KeyboardButton, 0, 2)
	for _, s := range a.AnimeSeasons {
		icon := "📺"
		if s.SeasonDubbed == 1 {
			icon = "🗣"
		}
		label := fmt.Sprintf("%s %s", icon, truncate(s.SeasonName, 22))
		row = append(row, tg.Button.Data(
			label,
			fmt.Sprintf("anime:ep:%d:%d:0", animeID, s.SeasonID),
		))
		if len(row) == 2 {
			kb.AddRow(row...)
			row = nil
		}
	}
	if len(row) > 0 {
		kb.AddRow(row...)
	}

	kb.AddRow(
		tg.Button.Data("🏠 Início", "anime:h"),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(nil, cb, b.String(), cover, kb.Build())
}

const episodesPerPage = 10

// renderEpisodes lists episodes for a season as a 2-column grid with
// numbered pagination. Thumbnails are cached for the now-playing screen.
func renderEpisodes(
	cb *tg.CallbackQuery,
	animeID, seasonID, page int,
) error {
	res, err := tomato.Default().SeasonEpisodes(seasonID, page, "ASC")
	if err != nil {
		gologging.ErrorF("anime: season %d page %d failed: %v", seasonID, page, err)
		return replyErr(nil, cb, friendlyErr(err))
	}
	if len(res.Data) == 0 {
		return replyErr(nil, cb, "😕 Nenhum episódio encontrado nessa temporada.")
	}

	// Cache every episode thumbnail so performPlay can paint a rich card.
	for _, ep := range res.Data {
		cacheEpisodeThumb(ep.EpID, ep.EpThumbnail)
	}

	start := page * episodesPerPage
	end := start + episodesPerPage
	if end > len(res.Data) {
		end = len(res.Data)
	}
	var slice []tomato.Episode
	if start < len(res.Data) {
		slice = res.Data[start:end]
	}

	totalPages := (len(res.Data) + episodesPerPage - 1) / episodesPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	var b strings.Builder
	fmt.Fprintf(
		&b,
		"📺 <b>Episódios</b>  <i>(%d no total • página %d/%d)</i>\n\n",
		res.Episodes, page+1, totalPages,
	)
	for _, ep := range slice {
		fmt.Fprintf(
			&b,
			"<b>Ep %d.</b> %s\n",
			ep.EpNumber,
			html.EscapeString(truncate(ep.EpName, 60)),
		)
	}

	// Episode buttons in a 2-column grid.
	kb := tg.NewKeyboard()
	row := make([]tg.KeyboardButton, 0, 2)
	for _, ep := range slice {
		label := fmt.Sprintf("▶️ Ep %d", ep.EpNumber)
		row = append(row, tg.Button.Data(
			label,
			fmt.Sprintf("anime:p:%d", ep.EpID),
		))
		if len(row) == 2 {
			kb.AddRow(row...)
			row = nil
		}
	}
	if len(row) > 0 {
		kb.AddRow(row...)
	}

	// Smart pagination: [◂] [ · N/M · ] [▸]
	kb.AddRow(paginationRow(
		page, totalPages,
		func(p int) string {
			return fmt.Sprintf("anime:ep:%d:%d:%d", animeID, seasonID, p)
		},
	)...)

	kb.AddRow(
		tg.Button.Data("🔙 Temporadas", fmt.Sprintf("anime:back:%d", animeID)),
		tg.Button.Data("🏠 Início", "anime:h"),
	)

	return sendOrEdit(nil, cb, b.String(), lookupAnimeCover(animeID), kb.Build())
}

// paginationRow builds a Telegram pagination row with a centered counter
// and prev/next arrows. Buttons outside the valid range are dimmed
// (rendered as a no-op).
func paginationRow(
	current, total int,
	hrefFn func(page int) string,
) []tg.KeyboardButton {
	if total <= 1 {
		return nil
	}

	prev := tg.Button.Data("·", "anime:noop")
	if current > 0 {
		prev = tg.Button.Data("◂", hrefFn(current-1))
	}
	counter := tg.Button.Data(
		fmt.Sprintf("· %d/%d ·", current+1, total),
		"anime:noop",
	)
	next := tg.Button.Data("·", "anime:noop")
	if current < total-1 {
		next = tg.Button.Data("▸", hrefFn(current+1))
	}
	return []tg.KeyboardButton{prev, counter, next}
}

// performPlay resolves the episode stream, picks a quality and injects it
// straight into the chat's voice chat via core.RoomState.Play. It then
// edits the callback message with a "now playing" card.
//
// quality: "fhd" | "shd" | "best"
func performPlay(cb *tg.CallbackQuery, episodeID int, quality string) error {
	chatID := cb.ChannelID()

	// Auto-play requires a group with an active video chat.
	// User DMs (positive ID) don't have voice chats.
	if chatID > 0 {
		cb.Answer(
			"⚠️ O /anime só toca em grupos com chat de vídeo ativo.",
			&tg.CallbackOptions{Alert: true},
		)
		return tg.ErrEndGroup
	}

	cb.Answer("🎬 Preparando o player...")

	s, err := tomato.Default().EpisodeStream(episodeID)
	if err != nil {
		gologging.ErrorF("anime: stream ep %d failed: %v", episodeID, err)
		return replyErr(nil, cb, friendlyErr(err))
	}

	url, label := pickStream(s.Streams, quality)
	if url == "" {
		return replyErr(nil, cb, "😕 Essa qualidade não está disponível.")
	}

	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		gologging.ErrorF("anime: ForChat(%d) failed: %v", chatID, err)
		return replyErr(nil, cb,
			"❌ <b>Não consegui resolver o assistente pra esse grupo.</b>\n\n"+
				"Confere se pelo menos uma <code>STRING_SESSIONS</code> está ativa.")
	}
	room, _ := core.GetRoom(chatID, ass, true)

	requester := "🍅 <i>via /anime</i>"
	if cb.Sender != nil {
		requester = fmt.Sprintf(
			`<a href="tg://user?id=%d">%s</a>`,
			cb.Sender.ID,
			html.EscapeString(cb.Sender.FirstName),
		)
	}

	track := &state.Track{
		ID:        fmt.Sprintf("tomato_%d", episodeID),
		Title:     s.EpisodeName,
		URL:       url,
		Video:     true,
		Source:    "TomatoAnimes",
		Requester: requester,
	}

	if err := room.Play(track, url, false); err != nil {
		gologging.ErrorF("anime: room.Play failed for ep %d: %v", episodeID, err)
		return replyErr(nil, cb,
			"❌ <b>Não consegui iniciar a reprodução.</b>\n\n"+
				"Confere se:\n"+
				"• O <b>chat de vídeo</b> está ativo no grupo\n"+
				"• O assistente <b>@"+ass.Self.Username+"</b> é membro\n"+
				"• O grupo permite streaming de vídeo")
	}

	// ---------- Now-playing card ----------
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"🎬 <b>Tocando no chat de vídeo!</b>\n\n"+
			"<b>%s</b>\n\n"+
			"🍿 Qualidade: <b>%s</b>\n",
		html.EscapeString(s.EpisodeName),
		label,
	)
	if s.EpisodeHasNext {
		fmt.Fprintf(
			&b,
			"\n➡️ <b>Próximo:</b> Ep %d — <i>%s</i>",
			s.EpisodeNumber+1,
			html.EscapeString(truncate(s.NextEpisodeTitle, 50)),
		)
	}

	kb := tg.NewKeyboard()
	// quality switcher
	qRow := []tg.KeyboardButton{}
	if quality != "fhd" && s.Streams.FHD != "" {
		qRow = append(qRow, tg.Button.Data(
			"🎬 1080p",
			fmt.Sprintf("anime:pq:%d:fhd", episodeID),
		))
	}
	if quality != "shd" && s.Streams.SHD != "" {
		qRow = append(qRow, tg.Button.Data(
			"📱 480p",
			fmt.Sprintf("anime:pq:%d:shd", episodeID),
		))
	}
	if len(qRow) > 0 {
		kb.AddRow(qRow...)
	}

	if s.EpisodeHasNext && s.NextEpisodeID > 0 {
		kb.AddRow(tg.Button.Data(
			"▶️ Próximo episódio",
			fmt.Sprintf("anime:p:%d", s.NextEpisodeID),
		))
	}
	kb.AddRow(
		tg.Button.Data(
			"🔙 Voltar",
			fmt.Sprintf("anime:v:%d", s.EpisodeAnimeID),
		),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	// Prefer the episode thumbnail (cached from the season listing),
	// otherwise fall back to the anime cover we've seen before.
	thumb := lookupEpisodeThumb(episodeID)
	if thumb == "" {
		thumb = lookupAnimeCover(s.EpisodeAnimeID)
	}
	return sendOrEdit(nil, cb, b.String(), thumb, kb.Build())
}

// pickStream chooses a URL based on requested quality and returns
// the URL along with a human label for the now-playing card.
func pickStream(
	q tomato.StreamQualities,
	quality string,
) (url, label string) {
	switch quality {
	case "fhd":
		return q.FHD, "1080p"
	case "shd":
		return q.SHD, "480p"
	}
	// best: pick highest available
	switch {
	case q.FHD != "":
		return q.FHD, "1080p"
	case q.MHD != "":
		return q.MHD, "720p"
	default:
		return q.SHD, "480p"
	}
}

// ---------- helpers ----------

// sendOrEdit replies with a fresh message (m) or edits the callback
// message (cb). media is an optional photo URL. Since Telegram can't
// always edit a text-only message into a photo message (or vice versa),
// we fall back to delete+repost when Edit fails.
func sendOrEdit(
	m *tg.NewMessage,
	cb *tg.CallbackQuery,
	text, media string,
	markup tg.ReplyMarkup,
) error {
	opts := &tg.SendOptions{
		ParseMode:   tg.HTML,
		ReplyMarkup: markup,
		LinkPreview: false,
	}
	if media != "" {
		opts.Media = media
	}
	if cb != nil {
		if _, err := cb.Edit(text, opts); err != nil {
			gologging.DebugF("anime: cb.Edit failed (%v) — reposting", err)
			cb.Delete()
			if _, err := cb.Client.SendMessage(cb.ChannelID(), text, opts); err != nil {
				gologging.ErrorF("anime: repost failed: %v", err)
			}
		}
		return tg.ErrEndGroup
	}
	if m != nil {
		if _, err := m.Reply(text, opts); err != nil {
			gologging.DebugF("anime: m.Reply failed: %v", err)
		}
	}
	return tg.ErrEndGroup
}

func replyErr(m *tg.NewMessage, cb *tg.CallbackQuery, text string) error {
	opts := &tg.SendOptions{ParseMode: tg.HTML}
	if cb != nil {
		cb.Answer(stripTags(text))
		return tg.ErrEndGroup
	}
	if m != nil {
		m.Reply(text, opts)
	}
	return tg.ErrEndGroup
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len([]rune(s)) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func stripTags(s string) string {
	// crude tag stripper for callback toasts
	out := s
	for strings.Contains(out, "<") {
		lt := strings.Index(out, "<")
		gt := strings.Index(out[lt:], ">")
		if gt < 0 {
			break
		}
		out = out[:lt] + out[lt+gt+1:]
	}
	return out
}

func encodeB64(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func decodeB64(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
