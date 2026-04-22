package modules

import (
	"encoding/base64"
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/core"
	state "main/internal/core/models"
	"main/internal/tomato"
)

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
		return renderHome(nil, cb)
	case "s": // anime:s:<b64query>:<page>
		if len(parts) < 4 {
			cb.Answer("🤔 Busca inválida")
			return tg.ErrEndGroup
		}
		query, _ := decodeB64(parts[2])
		page, _ := strconv.Atoi(parts[3])
		return renderSearch(nil, cb, query, page)
	case "v": // anime:v:<animeID>
		id, _ := strconv.Atoi(parts[2])
		return renderAnime(cb, id)
	case "ep": // anime:ep:<animeID>:<seasonID>:<page>
		if len(parts) < 5 {
			cb.Answer("🤔 Temporada inválida")
			return tg.ErrEndGroup
		}
		animeID, _ := strconv.Atoi(parts[2])
		seasonID, _ := strconv.Atoi(parts[3])
		page, _ := strconv.Atoi(parts[4])
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
		return renderAnime(cb, id)
	case "x": // anime:x — close
		cb.Answer("")
		cb.Delete()
		return tg.ErrEndGroup
	}

	cb.Answer("")
	return tg.ErrEndGroup
}

// ---------- renderers ----------

// renderHome shows the feed with rails.
// Either m (fresh message) or cb (callback) is provided.
func renderHome(m *tg.NewMessage, cb *tg.CallbackQuery) error {
	feed, err := tomato.Default().Feed()
	if err != nil {
		return replyErr(m, cb, "❌ Erro ao carregar feed: "+err.Error())
	}

	var b strings.Builder
	b.WriteString("🍅 <b>TomatoAnimes — Início</b>\n")
	b.WriteString("<i>Escolha um anime pra começar.</i>\n")

	kb := tg.NewKeyboard()
	rowsAdded := 0

	// Show trending + newly added + recommended (types 3 and 5), capped
	for _, sec := range feed.Data {
		if sec.Type != 3 && sec.Type != 5 {
			continue
		}
		if len(sec.Data) == 0 {
			continue
		}
		b.WriteString("\n<b>▸ ")
		b.WriteString(html.EscapeString(sec.Title))
		b.WriteString("</b>")

		// One row with up to 3 animes per section
		max := len(sec.Data)
		if max > 3 {
			max = 3
		}
		row := make([]tg.KeyboardButton, 0, max)
		for i := 0; i < max; i++ {
			a := sec.Data[i]
			if a.AnimeID == 0 {
				continue
			}
			label := fmt.Sprintf("#%d", a.AnimeID)
			row = append(row, tg.Button.Data(
				label,
				fmt.Sprintf("anime:v:%d", a.AnimeID),
			))
		}
		if len(row) > 0 {
			kb.AddRow(row...)
			rowsAdded++
		}
		if rowsAdded >= 5 {
			break
		}
	}

	b.WriteString("\n\n<b>🔎 Pra buscar:</b> <code>/anime &lt;nome&gt;</code>")

	kb.AddRow(
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(m, cb, b.String(), kb.Build())
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
		return replyErr(m, cb, "❌ Erro na busca: "+err.Error())
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

	// pagination
	navRow := []tg.KeyboardButton{}
	if page > 0 {
		navRow = append(navRow, tg.Button.Data(
			"« Anterior",
			fmt.Sprintf("anime:s:%s:%d", encodeB64(query), page-1),
		))
	}
	if len(res.Result) >= 40 { // heurística: se vieram muitos, provavelmente tem mais
		navRow = append(navRow, tg.Button.Data(
			"Próximo »",
			fmt.Sprintf("anime:s:%s:%d", encodeB64(query), page+1),
		))
	}
	if len(navRow) > 0 {
		kb.AddRow(navRow...)
	}

	kb.AddRow(
		tg.Button.Data("🏠 Início", "anime:h"),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(m, cb, b.String(), kb.Build())
}

// renderAnime shows the anime detail with seasons.
func renderAnime(cb *tg.CallbackQuery, animeID int) error {
	a, err := tomato.Default().Anime(animeID)
	if err != nil {
		return replyErr(nil, cb, "❌ Erro ao abrir anime: "+err.Error())
	}

	d := a.AnimeDetails
	var b strings.Builder
	fmt.Fprintf(&b, "🎬 <b>%s</b>  <i>(%s)</i>\n\n", html.EscapeString(d.AnimeName), html.EscapeString(d.AnimeDate))
	if d.AnimeGenre != "" {
		fmt.Fprintf(&b, "🏷 <b>Gêneros:</b> %s\n", html.EscapeString(d.AnimeGenre))
	}
	if d.AnimeParentalRating != "" {
		fmt.Fprintf(&b, "🔞 <b>Classificação:</b> %s\n", html.EscapeString(d.AnimeParentalRating))
	}
	fmt.Fprintf(&b, "🎞 <b>Episódios:</b> %d\n", d.AnimeEpisodes)
	if d.DubAvailable {
		b.WriteString("🗣 <b>Dublagem:</b> disponível ✅\n")
	}
	if d.AnimeDescription != "" {
		fmt.Fprintf(
			&b, "\n<i>%s</i>\n",
			html.EscapeString(truncate(d.AnimeDescription, 400)),
		)
	}
	b.WriteString("\n<b>📅 Escolha a temporada:</b>")

	kb := tg.NewKeyboard()
	for _, s := range a.AnimeSeasons {
		icon := "📺"
		if s.SeasonDubbed == 1 {
			icon = "🗣"
		}
		label := fmt.Sprintf("%s %s", icon, truncate(s.SeasonName, 30))
		kb.AddRow(tg.Button.Data(
			label,
			fmt.Sprintf("anime:ep:%d:%d:0", animeID, s.SeasonID),
		))
	}

	kb.AddRow(
		tg.Button.Data("🏠 Início", "anime:h"),
		tg.Button.Data("❌ Fechar", "anime:x"),
	)

	return sendOrEdit(nil, cb, b.String(), kb.Build())
}

const episodesPerPage = 10

// renderEpisodes lists episodes for a season (paginated).
func renderEpisodes(
	cb *tg.CallbackQuery,
	animeID, seasonID, page int,
) error {
	res, err := tomato.Default().SeasonEpisodes(seasonID, page, "ASC")
	if err != nil {
		return replyErr(nil, cb, "❌ Erro ao carregar episódios: "+err.Error())
	}
	if len(res.Data) == 0 {
		return replyErr(nil, cb, "😕 Nenhum episódio encontrado nessa temporada.")
	}

	var b strings.Builder
	fmt.Fprintf(
		&b,
		"📺 <b>Episódios</b>  <i>(total: %d)</i>\n\n",
		res.Episodes,
	)

	kb := tg.NewKeyboard()
	start := page * episodesPerPage
	end := start + episodesPerPage
	if end > len(res.Data) {
		end = len(res.Data)
	}

	// API might return the entire season — paginar local se for o caso
	slice := res.Data
	if end-start > 0 && end <= len(res.Data) && start < len(res.Data) {
		slice = res.Data[start:end]
	} else if start >= len(res.Data) {
		slice = nil
	}

	for _, ep := range slice {
		fmt.Fprintf(
			&b,
			"<b>Ep %d.</b> %s\n",
			ep.EpNumber,
			html.EscapeString(truncate(ep.EpName, 60)),
		)
		label := fmt.Sprintf("▶️ Ep %d — %s", ep.EpNumber, truncate(ep.EpName, 30))
		kb.AddRow(tg.Button.Data(
			label,
			fmt.Sprintf("anime:p:%d", ep.EpID),
		))
	}

	// Navegação de páginas local (a API parece devolver tudo numa página)
	totalPages := (len(res.Data) + episodesPerPage - 1) / episodesPerPage
	nav := []tg.KeyboardButton{}
	if page > 0 {
		nav = append(nav, tg.Button.Data(
			"« Anterior",
			fmt.Sprintf("anime:ep:%d:%d:%d", animeID, seasonID, page-1),
		))
	}
	if page < totalPages-1 {
		nav = append(nav, tg.Button.Data(
			"Próximo »",
			fmt.Sprintf("anime:ep:%d:%d:%d", animeID, seasonID, page+1),
		))
	}
	if len(nav) > 0 {
		kb.AddRow(nav...)
	}

	kb.AddRow(
		tg.Button.Data("🔙 Temporadas", fmt.Sprintf("anime:back:%d", animeID)),
		tg.Button.Data("🏠 Início", "anime:h"),
	)

	return sendOrEdit(nil, cb, b.String(), kb.Build())
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
		return replyErr(nil, cb, "❌ Erro ao gerar stream: "+err.Error())
	}

	url, label := pickStream(s.Streams, quality)
	if url == "" {
		return replyErr(nil, cb, "😕 Essa qualidade não está disponível.")
	}

	ass, err := core.Assistants.ForChat(chatID)
	if err != nil {
		return replyErr(nil, cb,
			"❌ Não consegui resolver o assistente pra esse grupo:\n<code>"+
				html.EscapeString(err.Error())+"</code>")
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
			"❌ Não consegui iniciar a reprodução.\n\n"+
				"Certifica que o <b>chat de vídeo</b> está ativo e o assistente "+
				"(<b>@"+ass.Self.Username+"</b>) é membro do grupo.\n\n"+
				"<i>Detalhe:</i> <code>"+html.EscapeString(err.Error())+"</code>")
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

	return sendOrEdit(nil, cb, b.String(), kb.Build())
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

func sendOrEdit(
	m *tg.NewMessage,
	cb *tg.CallbackQuery,
	text string,
	markup tg.ReplyMarkup,
) error {
	opts := &tg.SendOptions{
		ParseMode:   tg.HTML,
		ReplyMarkup: markup,
		LinkPreview: false,
	}
	if cb != nil {
		cb.Answer("")
		if _, err := cb.Edit(text, opts); err != nil {
			gologging.DebugF("anime: cb.Edit failed: %v", err)
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

