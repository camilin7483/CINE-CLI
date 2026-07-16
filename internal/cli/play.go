package cli

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/cam/cine-cli/internal/core"
)

func (a *App) runQuickPlay(ctx context.Context, query string) error {
	if a.Metadata == nil {
		return fmt.Errorf("TMDB API key not configured. Set tmdb_api_key in ~/.config/cine-cli/config.yaml")
	}

	results, err := a.Metadata.Search(ctx, core.SearchFilter{Query: query})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no results found for %q", query)
	}

	selected := results[0]
	fmt.Printf("\n  %s (%d) [%s]\n", selected.Title, selected.Year, selected.MediaType)

	tmdbID := fmt.Sprintf("%d", selected.TMDBID)
	providerID := tmdbID
	isSeries := selected.MediaType == core.MediaTypeSeries

	if isSeries {
		seasons, err := a.Metadata.GetSeasons(ctx, selected.TMDBID)
		if err == nil && len(seasons) > 0 {
			episodes, err := a.Metadata.GetEpisodes(ctx, selected.TMDBID, seasons[0].SeasonNumber)
			if err == nil && len(episodes) > 0 {
				providerID = fmt.Sprintf("%s/%d/%d", tmdbID, 1, 1)
				fmt.Printf("  Season 1 Episode 1: %s\n", episodes[0].Name)
			}
		}
	}

	// Check for continue watching
	var resumePosition float64
	progress, err := a.DB.GetProgress(ctx, selected.ID, 0, 0)
	if err == nil && progress != nil && !progress.Completed {
		resumePosition = progress.Position
		pct := int(progress.Percentage)
		fmt.Printf("  Continue from %s (%d%%)\n", formatDuration(progress.Position), pct)
	}

	providers := a.Manager.ListProviders()
	if a.Config.Provider != "" {
		providers = append([]string{a.Config.Provider}, providers...)
	}

	var stream *core.Stream
	var usedProvider string

	for _, pname := range providers {
		ref := core.MediaRef{
			ProviderName: pname,
			ProviderID:   providerID,
			Title:        selected.Title,
			MediaType:    selected.MediaType,
		}

		s, err := a.Manager.ResolveStream(ctx, ref)
		if err == nil && s != nil && s.URL != "" {
			stream = s
			usedProvider = pname
			fmt.Printf("  Stream resolved via %s\n", usedProvider)
			break
		}
	}

	if stream == nil || stream.URL == "" {
		fmt.Printf("  No stream available, opening in browser...\n")
		tmdbIDStr := fmt.Sprintf("%d", selected.TMDBID)
		url := fmt.Sprintf("https://vidsrc.to/embed/movie/%s", tmdbIDStr)
		exec.Command("xdg-open", url).Start()
		return nil
	}

	opts := core.PlayOptions{
		StreamURL:     stream.URL,
		Referer:       stream.Referer,
		UserAgent:     stream.UserAgent,
		Subtitles:     stream.Subtitles,
		Title:         selected.Title,
		Player:        a.Config.Player,
		PreferredLang: a.Config.SmartSelection.PreferredLang,
		SubsLang:      a.Config.SubtitlesLanguage,
	}

	if resumePosition > 0 {
		opts.ExtraArgs = append(opts.ExtraArgs, fmt.Sprintf("--start=%f", resumePosition))
	}

	if err := a.Player.Play(ctx, opts); err != nil {
		fmt.Printf("  Playback failed: %v\n  Opening in browser...\n", err)
		exec.Command("xdg-open", stream.URL).Start()
		return nil
	}

	a.DB.Add(ctx, core.HistoryEntry{
		MediaID:   selected.ID,
		Title:     selected.Title,
		MediaType: selected.MediaType,
		Provider:  usedProvider,
		StreamURL: stream.URL,
	})

	// Save progress tracking
	season, episode := 0, 0
	if isSeries {
		season = 1
		episode = 1
	}
	a.DB.SaveProgress(ctx, selected.ID, selected.Title, selected.MediaType,
		season, episode, 0, 0, usedProvider, stream.URL)

	return nil
}

func (a *App) runTUI(ctx context.Context) error {
	tui := NewTUI(a)
	_, err := tui.Run()
	return err
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
