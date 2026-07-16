# 🎬 cine-cli

> **Watch movies and TV shows directly from your terminal.**
>
> Developed by **CamiloDev**

---

## ✨ What is cine-cli?

A modern, lightning-fast CLI application for discovering and streaming movies and TV series. Built in Go with a clean hexagonal architecture, rich metadata from TMDB, and intelligent stream resolution from multiple providers.

*No browser required. Just your terminal and mpv.*

```bash
cine               # Launch the interactive TUI
cine watch "dune"  # Quick search and play
cine trending      # Browse trending content
cine search "oppenheimer"  # Search only
```

---

## 🚀 Quick Start

### 1. Dependencies

```bash
# Arch Linux (btw)
sudo pacman -S mpv yt-dlp

# macOS
brew install mpv yt-dlp
```

### 2. Get TMDB API Key (free)

1. Go to https://www.themoviedb.org/settings/api
2. Create a free account and get your API key
3. Save it:

```bash
mkdir -p ~/.config/cine-cli
cat > ~/.config/cine-cli/config.yaml << EOF
provider: vidsrc
player: mpv
tmdb_api_key: "your-api-key-here"
EOF
```

### 3. Install

```bash
cd cine-cli
go build -o ~/.local/bin/cine ./cmd/cine
```

Make sure `~/.local/bin` is in your `$PATH`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

---

## 🎮 Interactive TUI

```
┌─────────────────────────────────────────────────┐
│  🎬  cine-cli                    (TMDB)         │
│     developed by CamiloDev                      │
│                                                 │
│  Search: dune_                                  │
│                                                 │
│  enter to search · backspace · esc to quit      │
└─────────────────────────────────────────────────┘
```

### Keybindings

| Key | Action |
|-----|--------|
| `type + enter` | Search for a title |
| `↑ ↓` / `j k` | Navigate list |
| `enter` | Select / Play |
| `backspace` | Delete in search |
| `esc` | Go back / Quit |
| `tab space` | Toggle detail view |
| `b` | Open in browser |
| `t` | Trending |
| `?` | Help |
| `ctrl+c` | Quit |

### Screens

- **Search** — Type to search. Results appear instantly.
- **Results** — Browse search results with ratings and years.
- **Seasons** — Pick a season for TV shows.
- **Episodes** — Pick an episode to play.
- **Playing** — Shows current playback status.
- **Trending** — Browse trending movies and TV shows.

---

## 📡 Architecture

```
cine-cli/
├── cmd/cine/main.go              # Entry point
├── internal/
│   ├── cli/                      # Cobra CLI + Bubble Tea TUI
│   │   ├── tui.go                # TUI model & events
│   │   ├── tui_views.go          # TUI views (search, results, etc.)
│   │   └── tui_styles.go         # Lipgloss styling
│   ├── core/                     # Domain interfaces (hexagonal ports)
│   │   ├── media.go              # Media, Stream, History types
│   │   ├── provider.go           # Provider interface
│   │   ├── metadata.go           # MetadataProvider interface
│   │   ├── player.go             # Player interface
│   │   └── repository.go         # Store interfaces
│   ├── metadata/tmdb/            # TMDB API integration
│   ├── provider/
│   │   ├── registry.go           # Provider registry + manager
│   │   ├── scraper/embed.go      # Embed providers (vidsrc, 2embed, ...)
│   │   └── resolver/             # Stream resolvers
│   │       ├── vidsrc.go         # Vidsrc HTTP chain resolver
│   │       ├── browser.go        # Chrome headless resolver
│   │       ├── http.go           # Generic HTTP scraper
│   │       ├── ytdlp.go          # yt-dlp resolver
│   │       └── megacloud.go      # Megacloud AES decryptor
│   ├── player/                   # MPV/VLC player integration
│   ├── database/                 # SQLite (history, favorites, watchlist)
│   ├── config/                   # YAML configuration
│   └── cache/                    # Two-layer cache
├── pkg/types/                    # Shared types for plugins
└── config.yaml.example
```

---

## 🔗 Stream Resolution Pipeline

When you select a title, cine-cli resolves the stream URL through this pipeline:

```
TMDB ID
   │
   ▼
┌──────────────────┐
│ VidsrcResolver   │  ← Pure HTTP chain (fast, no browser)
│ vidsrc.to        │
│   → vsembed.ru   │
│   → cloudorchs-  │
│     tra.com/rcp  │
│   → .../prorcp   │
│   → m3u8 + token │  ← JWT from generate.php
├──────────────────┤
│ HTTP Resolver    │  ← Generic scraper (fallback)
├──────────────────┤
│ yt-dlp Resolver  │  ← yt-dlp extractors (last resort)
├──────────────────┤
│ xdg-open Browser │  ← Failsafe: open in desktop browser
└──────────────────┘
```

---

## 🧩 Providers

| # | Provider | Method | Status |
|---|----------|--------|--------|
| 1 | **vidsrc** | Vidsrc HTTP chain → token → m3u8 | ✅ |
| 2 | **2embed** | Embed API + iframe chain | 🟡 |
| 3 | **vidlink** | Embed API + JS decode | 🟡 |
| 4 | **vidsrcme** | Vidsrc HTTP chain (IMDB) | 🟡 |
| 5 | **superembed** | Multi-source aggregator | 🟡 |

---

## 🗄️ Database (SQLite)

```sql
history     — Watch history with resume positions
favorites   — Saved movies and series
watchlist   — Plan to watch list
cache       — Metadata cache with TTL
```

---

## 🛠️ Commands

```bash
cine                     # Launch TUI (default)
cine watch <query>       # Quick search & play
cine search <query>      # Search only
cine trending            # Show trending movies + TV
cine popular             # Popular movies
cine recommend <tmdb_id> # Recommendations by TMDB ID
cine browse              # Launch TUI explicitly
cine history             # Watch history
cine providers           # List available providers
cine config              # Show current configuration
cine --help              # Show all commands
```

---

## ⚙️ Configuration

```yaml
# ~/.config/cine-cli/config.yaml
provider: vidsrc          # Default stream provider
player: mpv               # Media player (mpv or vlc)
quality: ""               # Empty = auto, or "best", "1080p"
language: en-US           # Metadata language
tmdb_api_key: ""          # Your TMDB API key
data_dir: ~/.local/share/cine-cli
mpv_args:                 # Extra mpv arguments
  - "--hwdec=auto"
  - "--volume=70"
cache_ttl: 3600           # Cache TTL in seconds
max_results: 50           # Max search results
```

---

## 🧪 Requirements

- **Go 1.24+**
- **mpv** — Media player (primary)
- **VLC** — Alternative player (optional)
- **TMDB API key** — Free metadata enrichment
- **yt-dlp** — Additional site support (optional)
- **Google Chrome** — Headless stream extraction fallback (optional)

---

## 📝 License

MIT — Built with ❤️ by **CamiloDev**
