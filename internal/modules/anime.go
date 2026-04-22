package modules

import (
	"encoding/base64"
	"fmt"
	"html"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
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
	case "p": // anime:p:<episodeID>
		id, _ := strconv.Atoi(parts[2])
		return renderStream(cb, id)
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

// renderStream fetches stream URLs and shows ready-to-paste commands.
func renderStream(cb *tg.CallbackQuery, episodeID int) error {
	cb.Answer("🎬 Gerando stream...")

	s, err := tomato.Default().EpisodeStream(episodeID)
	if err != nil {
		return replyErr(nil, cb, "❌ Erro ao gerar stream: "+err.Error())
	}

	best := s.Streams.Best()
	if best == "" {
		return replyErr(nil, cb, "😕 Nenhum stream disponível pra esse episódio.")
	}

	var b strings.Builder
	fmt.Fprintf(
		&b,
		"🎬 <b>Pronto pra tocar!</b>\n\n<b>%s</b>\n\n",
		html.EscapeString(s.EpisodeName),
	)
	b.WriteString("📋 <b>Toque no comando pra copiar, depois cole no chat de vídeo:</b>\n\n")

	fmt.Fprintf(&b, "<code>/vplay %s</code>\n\n", best)

	if s.Streams.FHD != "" && s.Streams.SHD != "" && s.Streams.FHD != s.Streams.SHD {
		b.WriteString("<i>Se o 1080p travar, use o 480p:</i>\n")
		fmt.Fprintf(&b, "<code>/vplay %s</code>\n\n", s.Streams.SHD)
	}

	if s.EpisodeHasNext {
		fmt.Fprintf(
			&b,
			"➡️ <b>Próximo:</b> Ep %d — %s",
			s.EpisodeNumber+1,
			html.EscapeString(truncate(s.NextEpisodeTitle, 50)),
		)
	}

	kb := tg.NewKeyboard()
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

// shush unused imports in case of future expansion
var (
	_ = time.Now
	_ = sync.Mutex{}
	_ = config.BetomatoToken
)
