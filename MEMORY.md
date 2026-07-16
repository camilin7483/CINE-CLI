# cine-cli — Memory

## 2026-07-16: Professional upgrade

### What was done
Complete overhaul of cine-cli adding ~15 new features while preserving existing architecture.

### New features implemented

**Core infrastructure:**
- i18n system (6 locales: ES, EN, PT, FR, DE, IT) with embedded JSON files
- Enhanced config with subtitles, download dir, smart selection, player detection, keybindings, proxy
- Database migration system (v1→v2→v3) with continue_watching + downloads tables
- Continue Watching: tracks position/duration/percentage per media_id+season+episode

**CLI commands (16 total):**
- `cine search|watch|browse|history|trending|popular|providers|config` — all enhanced
- `cine favorites` — list, add, remove, export/import JSON, backup, restore, dedup
- `cine watchlist` — same as favorites + status management
- `cine download` — list, pause, resume, cancel, progress, cleanup
- `cine plugin` — list, enable, disable, discover, doc
- `cine update` — check, download, verify checksum, replace binary, restart
- `cine setup` — interactive wizard for language, player, provider, quality, subtitles, TMDB key, cache, downloads
- `cine completion bash|zsh|fish|powershell` — shell completion

**JSON output:**
- `--json` flag on all query commands for scripting

**Parallel search:**
- Multi-provider search with goroutines, context, timeout, dedup, relevance sorting

**Player detection:**
- Auto-detect mpv, vlc, celluloid, iina, mpc-hc, potplayer with priority config

**Smart selection:**
- Quality scoring engine, bandwidth filtering, language preferences, custom rules

**TUI enhancements:**
- Continue Watching section on search screen
- Favorites toggle (f key) with ♥ indicator
- Download from TUI (d key)
- Enhanced detail block (genres, rating, long overview)
- Updated help screen with all keybindings

**Downloads system:**
- Concurrent download manager with semaphore
- HTTP Range support for pause/resume
- Progress tracking (bytes, speed, percentage)
- File naming: `Title (Year) [Quality].mp4`

**Plugin system:**
- Go plugin support (plugin.so with Provider symbol)
- External script plugins (script.sh)
- Manifest-based discovery (manifest.json)
- Enable/disable per plugin

### Files changed/created
- `internal/i18n/` — NEW (7 files: engine + 6 locale JSONs)
- `internal/core/download.go` — NEW (Download types + interfaces)
- `internal/core/selection.go` — NEW (Smart selection config + quality scoring)
- `internal/core/media.go` — ADDED (ContinueWatching, HistoryFilter types)
- `internal/core/repository.go` — ADDED (new store interfaces)
- `internal/database/store.go` — UPDATED (v3 migration + downloads table)
- `internal/database/history.go` — UPDATED (Search, ListWithFilters, UpdatePosition, etc.)
- `internal/database/favorites.go` — UPDATED (export/import/backup/restore/dedup)
- `internal/database/watchlist.go` — UPDATED (same + status/query)
- `internal/database/continue_watching.go` — NEW
- `internal/download/` — NEW (manager.go, store.go)
- `internal/plugin/plugin.go` — NEW (registry, Go plugin, external plugin, doc)
- `internal/update/update.go` — NEW (GitHub release checker, downloader, replacer)
- `internal/search/engine.go` — NEW (parallel search, dedup, sorting, filtering)
- `internal/player/detect/detect.go` — NEW (player auto-detection)
- `internal/config/config.go` — UPDATED (all new fields + methods)
- `internal/cli/root.go` — UPDATED (all new commands + services)
- `internal/cli/search.go` — UPDATED (JSON flags, enhanced history)
- `internal/cli/play.go` — UPDATED (continue watching integration)
- `internal/cli/tui.go` — UPDATED (continue watching, favorites, download)
- `internal/cli/tui_views.go` — UPDATED (continue watching section, fav indicator)
- `internal/cli/tui_styles.go` — UPDATED (subtitle style)
- `internal/cli/favorites.go` — NEW
- `internal/cli/watchlist.go` — NEW
- `internal/cli/download.go` — NEW
- `internal/cli/plugin.go` — NEW
- `internal/cli/update.go` — NEW
- `internal/cli/setup.go` — NEW
- `internal/cli/completion.go` — NEW

### Commands
- Build: `go build -o ~/.local/bin/cine ./cmd/cine/`
- Binary: 30MB at `~/.local/bin/cine`
- Config: `~/.config/cine-cli/config.yaml`

### Architecture
Clean hexagonal:
- `internal/core/` — domain interfaces + types
- `internal/database/` — SQLite persistence
- `internal/metadata/` — TMDB provider
- `internal/provider/` — 5 stream providers + registry + resolvers
- `internal/player/` — MPV + VLC + auto-detection
- `internal/config/` — YAML config
- `internal/cache/` — two-layer cache
- `internal/i18n/` — internationalization
- `internal/download/` — concurrent download manager
- `internal/plugin/` — plugin registry
- `internal/update/` — auto-updater
- `internal/search/` — parallel search engine
- `internal/cli/` — Cobra CLI + Bubble Tea TUI (all commands)

### Not yet implemented
- Fuzzy search (needs external lib or custom algo)
- Statistics dashboard (hours watched, genre stats, etc.)
- More providers (PrimeWire, LookMovie, etc.)
- Subtitle providers (OpenSubtitles, SubDL)
- Offline mode
- Bookmarks
- Tests, CI/CD, GoReleaser, linting pipeline
