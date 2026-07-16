package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cam/cine-cli/internal/core"
	"github.com/spf13/cobra"
)

func (a *App) searchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for movies and TV shows",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			return a.runSearch(cmd.Context(), query)
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) watchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [query]",
		Short: "Quick search and play",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			return a.runQuickPlay(cmd.Context(), query)
		},
	}
}

func (a *App) browseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "browse",
		Short: "Launch interactive TUI browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runTUI(cmd.Context())
		},
	}
}

func (a *App) historyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show watch history",
		Long:  `View, search, filter, and sort watch history.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			query, _ := cmd.Flags().GetString("search")
			sortBy, _ := cmd.Flags().GetString("sort")
			sortOrder, _ := cmd.Flags().GetString("order")
			mediaType, _ := cmd.Flags().GetString("type")
			limit, _ := cmd.Flags().GetInt("limit")

			if query != "" || sortBy != "" || mediaType != "" {
				filter := core.HistoryFilter{
					Query:     query,
					MediaType: core.MediaType(mediaType),
					SortBy:    sortBy,
					SortOrder: sortOrder,
					Limit:     limit,
				}
				if filter.Limit <= 0 {
					filter.Limit = 50
				}
				entries, err := a.DB.ListWithFilters(cmd.Context(), filter)
				if err != nil {
					return err
				}
				return a.printHistory(entries)
			}

			entries, err := a.DB.List(cmd.Context(), limit, 0)
			if err != nil {
				return err
			}
			return a.printHistory(entries)
		},
	}

	cmd.Flags().StringP("search", "s", "", "Search history by title")
	cmd.Flags().StringP("sort", "o", "date", "Sort by: date, title, duration, progress")
	cmd.Flags().StringP("order", "r", "desc", "Sort order: asc, desc")
	cmd.Flags().StringP("type", "t", "", "Filter by type: movie, series")
	cmd.Flags().IntP("limit", "l", 20, "Number of entries")
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) printHistory(entries []core.HistoryEntry) error {
	if len(entries) == 0 {
		fmt.Println("No watch history yet.")
		return nil
	}
	if a.jsonOut {
		a.printJSON(entries)
		return nil
	}
	fmt.Println("\nWatch History")
	for _, e := range entries {
		pos := int(e.Position) / 60
		label := e.Title
		if e.MediaType == "series" && e.Season > 0 {
			label = fmt.Sprintf("%s S%02dE%02d", e.Title, e.Season, e.Episode)
		}
		provider := e.Provider
		if provider == "" {
			provider = "-"
		}
		fmt.Printf("  %s (%dm) [%s]\n", label, pos, provider)
	}
	return nil
}

func (a *App) configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.jsonOut {
				a.printJSON(a.Config)
				return nil
			}
			fmt.Printf("Provider:    %s\n", a.Config.Provider)
			fmt.Printf("Player:      %s\n", a.Config.Player)
			fmt.Printf("Language:    %s\n", a.Config.Language)
			fmt.Printf("Data Dir:    %s\n", a.Config.DataDir)
			fmt.Printf("Config Path: %s\n", a.Config.ConfigPath())
			fmt.Printf("DB Path:     %s\n", a.Config.DBPath())
			fmt.Printf("TMDB Key:    %s\n", maskKey(a.Config.TMDBAPIKey))
			fmt.Printf("Theme:       %s\n", a.Config.Theme)
			fmt.Printf("Quality:     %s\n", a.Config.DefaultQuality)
			fmt.Printf("Cache TTL:   %ds\n", a.Config.CacheTTL)
			fmt.Printf("Max DL:      %d\n", a.Config.MaxConcurrentDownloads)
			fmt.Printf("Download Dir: %s\n", a.Config.DownloadDir)

			providers := a.Manager.ListProviders()
			fmt.Printf("\nProviders:   %s\n", strings.Join(providers, ", "))
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) trendingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trending",
		Short: "Show trending movies and TV shows",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runTrending(cmd.Context())
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) popularCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "popular",
		Short: "Show popular movies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runTrendingCmd(cmd.Context(), core.MediaTypeMovie)
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) recommendationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommend [id]",
		Short: "Show recommendations for a TMDB ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmdbID := parseInt(args[0])
			return a.runRecommendations(cmd.Context(), tmdbID)
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) providersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "List available stream providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := a.Manager.ListProviders()
			if a.jsonOut {
				a.printJSON(providers)
				return nil
			}
			fmt.Println("Available stream providers:")
			for _, p := range providers {
				fmt.Printf("  - %s\n", p)
			}
			fmt.Printf("\nDefault: %s\n", a.Config.Provider)
			fmt.Println("\nConfigure with: cine config set provider <name>")
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func (a *App) runSearch(ctx context.Context, query string) error {
	if a.Metadata == nil {
		return fmt.Errorf("TMDB API key not configured")
	}

	results, err := a.Metadata.Search(ctx, core.SearchFilter{Query: query})
	if err != nil {
		return err
	}

	if a.jsonOut {
		a.printJSON(results)
		return nil
	}

	for _, m := range results {
		mt := "M"
		if m.MediaType == "series" {
			mt = "TV"
		}
		stars := ""
		if m.Rating > 0 {
			stars = fmt.Sprintf(" ★%.1f", m.Rating)
		}
		fmt.Printf("  [%s] %s (%d)%s — tmdb:%d\n", mt, m.Title, m.Year, stars, m.TMDBID)
	}
	return nil
}

func (a *App) runTrending(ctx context.Context) error {
	if a.Metadata == nil {
		return fmt.Errorf("TMDB API key not configured")
	}

	movies, _ := a.Metadata.GetTrending(ctx, core.MediaTypeMovie, 1)
	series, _ := a.Metadata.GetTrending(ctx, core.MediaTypeSeries, 1)

	if a.jsonOut {
		a.printJSON(map[string]interface{}{
			"movies": movies,
			"series": series,
		})
		return nil
	}

	fmt.Println("\nTrending Movies")
	for _, m := range movies {
		fmt.Printf("  [M] %s (%d) ★%.1f — tmdb:%d\n", m.Title, m.Year, m.Rating, m.TMDBID)
	}

	fmt.Println("\nTrending TV Shows")
	for _, s := range series {
		fmt.Printf("  [TV] %s (%d) ★%.1f — tmdb:%d\n", s.Title, s.Year, s.Rating, s.TMDBID)
	}

	return nil
}

func (a *App) runTrendingCmd(ctx context.Context, mediaType core.MediaType) error {
	if a.Metadata == nil {
		return fmt.Errorf("TMDB API key not configured")
	}

	results, err := a.Metadata.GetTrending(ctx, mediaType, 1)
	if err != nil {
		return err
	}

	if a.jsonOut {
		a.printJSON(results)
		return nil
	}

	for _, m := range results {
		mt := "M"
		if m.MediaType == "series" {
			mt = "TV"
		}
		fmt.Printf("  [%s] %s (%d) ★%.1f — tmdb:%d\n", mt, m.Title, m.Year, m.Rating, m.TMDBID)
	}
	return nil
}

func (a *App) runRecommendations(ctx context.Context, tmdbID int) error {
	if a.Metadata == nil {
		return fmt.Errorf("TMDB API key not configured")
	}

	results, err := a.Metadata.GetRecommendations(ctx, tmdbID, core.MediaTypeMovie)
	if err != nil {
		return err
	}

	if a.jsonOut {
		a.printJSON(results)
		return nil
	}

	fmt.Printf("\nRecommendations for TMDB %d:\n", tmdbID)
	for _, m := range results {
		fmt.Printf("  [M] %s (%d) ★%.1f — tmdb:%d\n", m.Title, m.Year, m.Rating, m.TMDBID)
	}
	return nil
}

func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "********"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
