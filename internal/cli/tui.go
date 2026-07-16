package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cam/cine-cli/internal/core"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenSearch screen = iota
	screenResults
	screenSeasons
	screenEpisodes
	screenPlaying
	screenHelp
	screenBrowser
	screenTrending
	screenResolving
	screenFavorites
	screenWatchlist
	screenHistory
)

const visibleItems = 15

var chars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type model struct {
	app         *App
	screen      screen
	prevScreen  screen
	err         string
	search      string
	results     []core.Media
	trending    []core.Media
	favs        []core.Favorite
	watchlist   []core.WatchlistItem
	history     []core.HistoryEntry
	continueW   []core.ContinueWatching
	selected    *core.Media
	seasons     []core.Season
	episodes    []core.Episode
	season      int
	episode     int
	cursor      int
	scrollIdx   int
	loading     bool
	detail      bool
	width       int
	height      int
	resolveMsg  string
	isFav       bool
	sidebarOpen bool
	sidebarIdx  int
	spinnerIdx  int
	viewMode    string
	themeSet    bool
	audioLang   string
	subsLang    string
}

var audioLangs = []string{"en", "es", "ja", "pt", "fr", "de", "it", "ko", "zh", "ar", "ru", "hi", "original"}

func NewTUI(app *App) *tea.Program {
	initStyles(app.Config.ThemeMode)
	lang := app.Config.SmartSelection.PreferredLang
	if lang == "" {
		lang = "en"
	}
	subs := app.Config.SubtitlesLanguage
	if subs == "" {
		subs = lang
	}
	m := &model{
		app:        app,
		screen:     screenSearch,
		viewMode:   "list",
		sidebarIdx: 0,
		audioLang:  lang,
		subsLang:   subs,
	}
	return tea.NewProgram(m, tea.WithAltScreen())
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.loadTrending(), m.loadContinueWatching(), m.spinnerTick())
}

type trendingMsg []core.Media
type continueWatchingMsg []core.ContinueWatching
type resolvingMsg struct{ err string }
type spinnerTick struct{}
type favoritesMsg []core.Favorite
type watchlistMsg []core.WatchlistItem
type historyMsg []core.HistoryEntry
type downloadResultMsg struct {
	title string
	err   string
}

func (m *model) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTick{}
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = minInt(msg.Width, 120)
		m.height = msg.Height
		if !m.themeSet {
			initStyles(m.app.Config.ThemeMode)
			m.themeSet = true
		}
		return m, nil

	case trendingMsg:
		m.trending = msg
		return m, nil

	case continueWatchingMsg:
		m.continueW = msg
		return m, nil

	case favoritesMsg:
		m.favs = msg
		return m, nil

	case watchlistMsg:
		m.watchlist = msg
		return m, nil

	case historyMsg:
		m.history = msg
		return m, nil

	case downloadResultMsg:
		if msg.err != "" {
			m.err = msg.err
		} else {
			m.err = ""
		}
		return m, nil

	case spinnerTick:
		m.spinnerIdx = (m.spinnerIdx + 1) % len(chars)
		if m.loading {
			return m, m.spinnerTick()
		}
		return m, nil

	case resolvingMsg:
		if msg.err != "" {
			m.err = msg.err
			m.screen = screenResults
			m.loading = false
			return m, nil
		}
		m.loading = false
		m.screen = screenPlaying
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *model) loadTrending() tea.Cmd {
	return func() tea.Msg {
		if m.app.Metadata == nil {
			return trendingMsg(nil)
		}
		movies, _ := m.app.Metadata.GetTrending(context.Background(), core.MediaTypeMovie, 1)
		series, _ := m.app.Metadata.GetTrending(context.Background(), core.MediaTypeSeries, 1)
		all := append(movies, series...)
		return trendingMsg(deduplicate(all))
	}
}

func (m *model) loadContinueWatching() tea.Cmd {
	return func() tea.Msg {
		items, err := m.app.DB.ListContinueWatching(context.Background(), 5)
		if err != nil || items == nil {
			return continueWatchingMsg(nil)
		}
		return continueWatchingMsg(items)
	}
}

func (m *model) loadFavorites() tea.Cmd {
	return func() tea.Msg {
		favs, err := m.app.DB.ListFavorites(context.Background())
		if err != nil || favs == nil {
			return favoritesMsg(nil)
		}
		return favoritesMsg(favs)
	}
}

func (m *model) loadWatchlist() tea.Cmd {
	return func() tea.Msg {
		items, err := m.app.DB.ListWatchlist(context.Background())
		if err != nil || items == nil {
			return watchlistMsg(nil)
		}
		return watchlistMsg(items)
	}
}

func (m *model) loadHistory() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.app.DB.List(context.Background(), 50, 0)
		if err != nil || entries == nil {
			return historyMsg(nil)
		}
		return historyMsg(entries)
	}
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "ctrl+c" {
		return m, tea.Quit
	}
	if key == "esc" {
		if m.sidebarOpen {
			m.sidebarOpen = false
			return m, nil
		}
		if m.screen == screenSearch {
			return m, tea.Quit
		}
		m.goBack()
		return m, nil
	}

	if m.sidebarOpen {
		return m.handleSidebarKey(key)
	}

	if m.screen == screenSearch {
		switch key {
		case "enter":
			if len(strings.TrimSpace(m.search)) > 0 {
				return m, m.doSearch()
			}
			return m, nil
		case "backspace":
			if len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
			}
			return m, nil
		case "?":
			m.prevScreen = m.screen
			m.screen = screenHelp
			return m, nil
		case "T":
			m.screen = screenTrending
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadTrending()
		case "S":
			m.sidebarOpen = true
			return m, nil
		case "L":
			m.cycleLanguage()
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.search += strings.ToLower(key)
			}
			return m, nil
		}
	}

	// Normalizar mayúsculas para soportar Shift+key
	lkey := strings.ToLower(key)

	switch lkey {
	case "enter":
		return m.handleEnter()

	case "up", "k":
		m.moveCursor(-1)
		return m, nil

	case "down", "j":
		m.moveCursor(1)
		return m, nil

	case "tab", " ":
		if m.screen == screenResults && m.selected != nil {
			m.detail = !m.detail
			if m.detail {
				m.viewMode = "split"
			} else {
				m.viewMode = "list"
			}
		}
		return m, nil

	case "f":
		if m.screen == screenResults && m.selected != nil {
			return m, m.toggleFavorite()
		}
		return m, nil

	case "d":
		if m.screen == screenResults && m.selected != nil {
			return m, m.startDownload()
		}
		return m, nil

	case "b":
		if m.screen == screenResults || m.screen == screenSeasons || m.screen == screenEpisodes || m.screen == screenPlaying {
			return m, m.openBrowser()
		}
		return m, nil

	case "t":
		if m.screen == screenResults || m.screen == screenSeasons || m.screen == screenEpisodes {
			m.screen = screenTrending
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadTrending()
		}
		return m, nil

	case "s":
		m.sidebarOpen = !m.sidebarOpen
		return m, nil

	case "v":
		if m.screen == screenFavorites {
			if len(m.favs) == 0 || m.cursor >= len(m.favs) {
				return m, nil
			}
			fav := m.favs[m.cursor]
			m.selected = &core.Media{
				ID:        fav.MediaID,
				Title:     fav.Title,
				MediaType: fav.MediaType,
				PosterURL: fav.PosterURL,
			}
			return m, m.startResolve()
		}
		return m, nil

	case "l":
		m.cycleLanguage()
		return m, nil

	case "?":
		if m.screen == screenHelp {
			m.screen = m.prevScreen
		} else {
			m.prevScreen = m.screen
			m.screen = screenHelp
		}
		return m, nil

	case "backspace":
		return m, nil

	default:
		if m.screen == screenPlaying || m.screen == screenBrowser {
			m.screen = screenResults
			m.cursor = 0
			m.scrollIdx = 0
			return m, nil
		}
	}
	return m, nil
}

func (m *model) handleSidebarKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.sidebarIdx--
		if m.sidebarIdx < 0 {
			m.sidebarIdx = 0
		}
		return m, nil
	case "down", "j":
		maxIdx := 5
		if m.sidebarIdx < maxIdx {
			m.sidebarIdx++
		}
		return m, nil
	case "enter":
		m.sidebarOpen = false
		switch m.sidebarIdx {
		case 0:
			m.screen = screenSearch
			m.cursor = 0
			m.scrollIdx = 0
		case 1:
			m.screen = screenTrending
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadTrending()
		case 2:
			m.screen = screenFavorites
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadFavorites()
		case 3:
			m.screen = screenWatchlist
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadWatchlist()
		case 4:
			m.screen = screenHistory
			m.cursor = 0
			m.scrollIdx = 0
			return m, m.loadHistory()
		case 5:
			m.screen = screenHelp
			m.prevScreen = screenSearch
		}
		return m, nil
	case "esc":
		m.sidebarOpen = false
		return m, nil
	}
	return m, nil
}

func (m *model) moveCursor(delta int) {
	maxLen := m.listLen()
	if maxLen == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= maxLen {
		m.cursor = maxLen - 1
	}
	if m.cursor < m.scrollIdx {
		m.scrollIdx = m.cursor
	}
	if m.cursor >= m.scrollIdx+visibleItems {
		m.scrollIdx = m.cursor - visibleItems + 1
	}
}

func (m *model) goBack() {
	switch m.screen {
	case screenResults, screenHelp, screenTrending, screenFavorites, screenWatchlist, screenHistory:
		m.screen = screenSearch
		m.detail = false
		m.viewMode = "list"
		m.err = ""
		m.cursor = 0
		m.scrollIdx = 0
	case screenSeasons:
		m.screen = screenResults
		m.cursor = 0
		m.scrollIdx = 0
	case screenEpisodes:
		m.screen = screenSeasons
		m.cursor = 0
		m.scrollIdx = 0
	case screenPlaying, screenBrowser:
		m.screen = screenResults
		m.cursor = 0
		m.scrollIdx = 0
	}
}

func (m *model) listLen() int {
	switch m.screen {
	case screenResults:
		return len(m.results)
	case screenSeasons:
		return len(m.seasons)
	case screenEpisodes:
		return len(m.episodes)
	case screenTrending:
		return len(m.trending)
	case screenFavorites:
		return len(m.favs)
	case screenWatchlist:
		return len(m.watchlist)
	case screenHistory:
		return len(m.history)
	}
	return 0
}

func (m *model) currentList() []core.Media {
	switch m.screen {
	case screenResults:
		return m.results
	case screenTrending:
		return m.trending
	}
	return nil
}

func (m *model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenSearch:
		if len(strings.TrimSpace(m.search)) > 0 {
			return m, m.doSearch()
		}
		return m, nil

	case screenResults:
		if len(m.results) == 0 || m.cursor >= len(m.results) {
			return m, nil
		}
		sel := &m.results[m.cursor]
		m.selected = sel
		m.isFav = m.checkFavorite(sel.ID)
		m.detail = false
		m.viewMode = "list"
		if sel.MediaType == core.MediaTypeSeries {
			return m, m.loadSeasons()
		}
		return m, m.startResolve()

	case screenTrending:
		if len(m.trending) == 0 || m.cursor >= len(m.trending) {
			return m, nil
		}
		sel := &m.trending[m.cursor]
		m.selected = sel
		m.isFav = m.checkFavorite(sel.ID)
		m.cursor = 0
		m.scrollIdx = 0
		if sel.MediaType == core.MediaTypeSeries {
			return m, m.loadSeasons()
		}
		return m, m.startResolve()

	case screenSeasons:
		if len(m.seasons) == 0 || m.cursor >= len(m.seasons) {
			return m, nil
		}
		m.season = m.seasons[m.cursor].SeasonNumber
		m.cursor = 0
		m.scrollIdx = 0
		return m, m.loadEpisodes()

	case screenEpisodes:
		if len(m.episodes) == 0 || m.cursor >= len(m.episodes) {
			return m, nil
		}
		m.episode = m.episodes[m.cursor].EpisodeNumber
		return m, m.startResolve()

	case screenFavorites:
		if len(m.favs) == 0 || m.cursor >= len(m.favs) {
			return m, nil
		}
		fav := m.favs[m.cursor]
		m.selected = &core.Media{
			ID:        fav.MediaID,
			Title:     fav.Title,
			MediaType: fav.MediaType,
			PosterURL: fav.PosterURL,
		}
		return m, m.startResolve()

	case screenWatchlist:
		if len(m.watchlist) == 0 || m.cursor >= len(m.watchlist) {
			return m, nil
		}
		item := m.watchlist[m.cursor]
		m.selected = &core.Media{
			ID:        item.MediaID,
			Title:     item.Title,
			MediaType: item.MediaType,
		}
		return m, m.startResolve()

	case screenHistory:
		if len(m.history) == 0 || m.cursor >= len(m.history) {
			return m, nil
		}
		entry := m.history[m.cursor]
		m.selected = &core.Media{
			ID:        entry.MediaID,
			Title:     entry.Title,
			MediaType: entry.MediaType,
		}
		return m, m.startResolve()
	}
	return m, nil
}

func (m *model) checkFavorite(mediaID string) bool {
	exists, _ := m.app.DB.FavoriteExists(context.Background(), mediaID)
	return exists
}

func (m *model) cycleLanguage() {
	for i, l := range audioLangs {
		if l == m.audioLang {
			m.audioLang = audioLangs[(i+1)%len(audioLangs)]
			m.subsLang = m.audioLang
			m.app.Config.SmartSelection.PreferredLang = m.audioLang
			m.app.Config.SubtitlesLanguage = m.subsLang
			m.app.Config.Save()
			return
		}
	}
	m.audioLang = audioLangs[0]
	m.subsLang = audioLangs[0]
}

func (m *model) toggleFavorite() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return nil
		}
		exists, _ := m.app.DB.FavoriteExists(context.Background(), m.selected.ID)
		if exists {
			m.app.DB.RemoveFavorite(context.Background(), m.selected.ID)
		} else {
			m.app.DB.AddFavorite(context.Background(), core.Favorite{
				MediaID:   m.selected.ID,
				Title:     m.selected.Title,
				MediaType: m.selected.MediaType,
				PosterURL: m.selected.PosterURL,
			})
		}
		m.isFav = !exists
		return nil
	}
}

func (m *model) startDownload() tea.Cmd {
	return func() tea.Msg {
		if m.selected == nil {
			return downloadResultMsg{err: "no media selected"}
		}

		tmdbID := fmt.Sprintf("%d", m.selected.TMDBID)
		providerID := tmdbID
		if m.selected.MediaType == core.MediaTypeSeries {
			providerID = fmt.Sprintf("%s/%d/%d", tmdbID, m.season, m.episode)
		}

		providers := m.app.Manager.ListProviders()
		if m.app.Config.Provider != "" {
			providers = append([]string{m.app.Config.Provider}, providers...)
		}

		var streamURL, referer, ua, usedProvider string
		for _, pname := range providers {
			ref := core.MediaRef{
				ProviderName: pname,
				ProviderID:   providerID,
				Title:        m.selected.Title,
				MediaType:    m.selected.MediaType,
			}
			s, err := m.app.Manager.ResolveStream(context.Background(), ref)
			if err == nil && s != nil && s.URL != "" && isStreamURL(s.URL) {
				streamURL = s.URL
				referer = s.Referer
				ua = s.UserAgent
				usedProvider = pname
				break
			}
		}

		if streamURL == "" {
			return downloadResultMsg{err: "Could not resolve stream for download"}
		}

		dl := core.Download{
			MediaID:   m.selected.ID,
			Title:     m.selected.Title,
			MediaType: m.selected.MediaType,
			Season:    m.season,
			Episode:   m.episode,
			URL:       streamURL,
			Referer:   referer,
			UserAgent: ua,
			Quality:   m.app.Config.DefaultQuality,
			Provider:  usedProvider,
		}
		if err := m.app.Downloads.Enqueue(context.Background(), dl); err != nil {
			return downloadResultMsg{err: fmt.Sprintf("Download error: %v", err)}
		}
		return downloadResultMsg{title: m.selected.Title}
	}
}

func (m *model) doSearch() tea.Cmd {
	return func() tea.Msg {
		query := strings.TrimSpace(m.search)
		if len(query) == 0 {
			return nil
		}
		m.loading = true
		results, err := m.app.Metadata.Search(context.Background(), core.SearchFilter{Query: query})
		if err != nil {
			m.err = err.Error()
			m.loading = false
			return nil
		}
		m.results = deduplicate(results)
		m.cursor = 0
		m.scrollIdx = 0
		m.loading = false
		m.screen = screenResults
		return nil
	}
}

func (m *model) loadSeasons() tea.Cmd {
	return func() tea.Msg {
		m.loading = true
		seasons, err := m.app.Metadata.GetSeasons(context.Background(), m.selected.TMDBID)
		if err != nil {
			m.err = err.Error()
			m.loading = false
			return nil
		}
		var filtered []core.Season
		for _, s := range seasons {
			if s.SeasonNumber > 0 {
				filtered = append(filtered, s)
			}
		}
		m.seasons = filtered
		m.cursor = 0
		m.scrollIdx = 0
		m.loading = false
		m.screen = screenSeasons
		return nil
	}
}

func (m *model) loadEpisodes() tea.Cmd {
	return func() tea.Msg {
		m.loading = true
		episodes, err := m.app.Metadata.GetEpisodes(context.Background(), m.selected.TMDBID, m.season)
		if err != nil {
			m.err = err.Error()
			m.loading = false
			return nil
		}
		m.episodes = episodes
		m.cursor = 0
		m.scrollIdx = 0
		m.loading = false
		m.screen = screenEpisodes
		return nil
	}
}

func (m *model) startResolve() tea.Cmd {
	m.screen = screenResolving
	m.loading = true
	m.err = ""
	return m.resolveStream()
}

func (m *model) resolveStream() tea.Cmd {
	return func() tea.Msg {
		tmdbID := fmt.Sprintf("%d", m.selected.TMDBID)
		providerID := tmdbID
		if m.selected.MediaType == core.MediaTypeSeries {
			providerID = fmt.Sprintf("%s/%d/%d", tmdbID, m.season, m.episode)
		}

		providers := m.app.Manager.ListProviders()
		if m.app.Config.Provider != "" {
			providers = append([]string{m.app.Config.Provider}, providers...)
		}

		var stream *core.Stream
		var usedProvider string
		for _, pname := range providers {
			ref := core.MediaRef{
				ProviderName: pname,
				ProviderID:   providerID,
				Title:        m.selected.Title,
				MediaType:    m.selected.MediaType,
			}
			s, err := m.app.Manager.ResolveStream(context.Background(), ref)
			if err == nil && s != nil && s.URL != "" && isStreamURL(s.URL) {
				stream = s
				usedProvider = pname
				break
			}
		}

		if stream == nil || stream.URL == "" {
			browserURL := buildBrowserURL(m.selected, m.season, m.episode)
			exec.Command("xdg-open", browserURL).Start()
			return resolvingMsg{err: fmt.Sprintf("Could not extract stream. Opened in browser: %s", browserURL)}
		}

		opts := core.PlayOptions{
			StreamURL:     stream.URL,
			Referer:       stream.Referer,
			UserAgent:     stream.UserAgent,
			Subtitles:     stream.Subtitles,
			Title:         m.selected.Title,
			Player:        m.app.Config.Player,
			ExtraArgs:     m.app.Config.MPVArgs,
			PreferredLang: m.audioLang,
			SubsLang:      m.subsLang,
		}

		progress, _ := m.app.DB.GetProgress(context.Background(), m.selected.ID, m.season, m.episode)
		if progress != nil && !progress.Completed && progress.Position > 0 {
			opts.ExtraArgs = append(opts.ExtraArgs, fmt.Sprintf("--start=%f", progress.Position))
		}

		if err := m.app.Player.Play(context.Background(), opts); err != nil {
			browserURL := buildBrowserURL(m.selected, m.season, m.episode)
			exec.Command("xdg-open", browserURL).Start()
			return resolvingMsg{err: fmt.Sprintf("Playback failed. Opened in browser: %s", browserURL)}
		}

		m.app.DB.Add(context.Background(), core.HistoryEntry{
			MediaID:   m.selected.ID,
			Title:     m.selected.Title,
			MediaType: m.selected.MediaType,
			Season:    m.season,
			Episode:   m.episode,
			Provider:  usedProvider,
			StreamURL: stream.URL,
		})

		m.app.DB.SaveProgress(context.Background(), m.selected.ID, m.selected.Title,
			m.selected.MediaType, m.season, m.episode, 0, 0, usedProvider, stream.URL)

		return resolvingMsg{}
	}
}

func isStreamURL(url string) bool {
	lower := strings.ToLower(url)
	return strings.Contains(lower, ".m3u8") ||
		strings.Contains(lower, ".mp4") ||
		strings.Contains(lower, ".mkv") ||
		strings.Contains(lower, "/video/") ||
		strings.Contains(lower, "/stream/")
}

func buildBrowserURL(media *core.Media, season, episode int) string {
	tmdbIDStr := fmt.Sprintf("%d", media.TMDBID)
	if media.MediaType == core.MediaTypeSeries {
		return fmt.Sprintf("https://vidsrc.to/embed/tv/%s/%d/%d", tmdbIDStr, season, episode)
	}
	return fmt.Sprintf("https://vidsrc.to/embed/movie/%s", tmdbIDStr)
}

func (m *model) openBrowser() tea.Cmd {
	return func() tea.Msg {
		url := buildBrowserURL(m.selected, m.season, m.episode)
		exec.Command("xdg-open", url).Start()
		m.screen = screenBrowser
		return nil
	}
}

func deduplicate(media []core.Media) []core.Media {
	seen := make(map[string]bool)
	var result []core.Media
	for _, m := range media {
		key := m.Title
		if m.Year > 0 {
			key = fmt.Sprintf("%s-%d", m.Title, m.Year)
		}
		if !seen[key] {
			seen[key] = true
			result = append(result, m)
		}
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
