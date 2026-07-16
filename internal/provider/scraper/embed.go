package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/cam/cine-cli/internal/core"
	"github.com/cam/cine-cli/internal/provider/resolver"
)

var embedUA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

type embedProvider struct {
	name     string
	priority int
	resolver *resolver.Chain
	ytdlp    *resolver.YTDlpStream
}

func (p *embedProvider) Name() string  { return p.name }
func (p *embedProvider) Priority() int { return p.priority }

func (p *embedProvider) Search(ctx context.Context, query string) ([]core.MediaRef, error) {
	return nil, nil
}

func (p *embedProvider) buildEmbedURL(ref core.MediaRef) string {
	id := ref.ProviderID
	mediaType := ref.MediaType

	switch p.name {
	case "vidsrc":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://vidsrc.to/embed/tv/%s/%d/%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://vidsrc.to/embed/movie/%s", id)
	case "2embed":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://www.2embed.cc/embedtv/%s&s=%d&e=%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://www.2embed.cc/embed/%s", id)
	case "vidlink":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://vidlink.pro/tv/%s/%d/%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://vidlink.pro/movie/%s", id)
	case "vidsrcme":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://vidsrc.me/embed/tv?tmdb=%s&season=%d&episode=%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://vidsrc.me/embed/movie?tmdb=%s", id)
	case "superembed":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://multiembed.mov/direct.php?video_id=%s&s=%d&e=%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://multiembed.mov/direct.php?video_id=%s", id)
	case "multiem":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://multiembed.mov/embed/tv/%s/%d/%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://multiembed.mov/embed/movie/%s", id)
	case "autoembed":
		if mediaType == core.MediaTypeSeries {
			s, ep := parseSE(id)
			return fmt.Sprintf("https://autoembed.cc/tv/%s/%d/%d", tmdbPart(id), s, ep)
		}
		return fmt.Sprintf("https://autoembed.cc/movie/%s", id)
	default:
		return fmt.Sprintf("https://vidsrc.to/embed/movie/%s", id)
	}
}

func (p *embedProvider) GetStream(ctx context.Context, ref core.MediaRef) (*core.Stream, error) {
	embedURL := p.buildEmbedURL(ref)

	result, err := p.resolver.Resolve(ctx, embedURL)
	if err == nil && result != nil && result.URL != "" {
		return &core.Stream{
			URL:       result.URL,
			Referer:   result.Referer,
			UserAgent: result.UserAgent,
			Subtitles: toSubtitles(result.Subtitles),
			Provider:  p.name,
			IsM3U8:    isM3U8(result.URL),
		}, nil
	}

	// yt-dlp as last resort
	if p.ytdlp != nil && p.ytdlp.Available() {
		stream, yterr := p.ytdlp.ResolveStream(ctx, embedURL)
		if yterr == nil && stream != nil && stream.URL != "" {
			stream.Provider = p.name
			return stream, nil
		}
	}

	// Return nil + error — don't return a fake stream URL
	return nil, fmt.Errorf("provider %s: could not resolve stream for %s", p.name, embedURL)
}

func newEmbedProvider(name string, priority int, chain *resolver.Chain, yt *resolver.YTDlpStream) *embedProvider {
	return &embedProvider{
		name:     name,
		priority: priority,
		resolver: chain,
		ytdlp:    yt,
	}
}

func RegisterAll(registry core.ProviderRegistry) {
	registerEmbedProviders(registry)
}

func registerEmbedProviders(registry core.ProviderRegistry) {
	vidsrcChain := resolver.NewVidsrcResolver()
	httpFb := resolver.NewHTTP()
	chain := resolver.NewChain(vidsrcChain, httpFb)
	yt := resolver.NewYTDlpStream(20 * time.Second)

	registry.Register(newEmbedProvider("vidsrc", 1, chain, yt))
	registry.Register(newEmbedProvider("2embed", 2, chain, yt))
	registry.Register(newEmbedProvider("vidlink", 3, chain, yt))
	registry.Register(newEmbedProvider("vidsrcme", 4, chain, yt))
	registry.Register(newEmbedProvider("superembed", 5, chain, yt))
	registry.Register(newEmbedProvider("multiem", 6, chain, yt))
	registry.Register(newEmbedProvider("autoembed", 7, chain, yt))
}

func parseSE(providerID string) (int, int) {
	s, ep := 1, 1
	parts := splitSlash(providerID)
	if len(parts) >= 3 {
		s = atoiOr(parts[1], 1)
		ep = atoiOr(parts[2], 1)
	}
	return s, ep
}

func tmdbPart(providerID string) string {
	parts := splitSlash(providerID)
	if len(parts) > 0 {
		return parts[0]
	}
	return providerID
}

func splitSlash(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == '/' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func atoiOr(s string, def int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return def
	}
	return n
}

func isM3U8(url string) bool {
	for i := 0; i < len(url)-4; i++ {
		if url[i] == '.' && url[i+1] == 'm' && url[i+2] == '3' && url[i+3] == 'u' && i+4 < len(url) && url[i+4] == '8' {
			return true
		}
	}
	return false
}

func toSubtitles(subs []resolver.SubData) []core.Subtitle {
	var result []core.Subtitle
	for _, s := range subs {
		result = append(result, core.Subtitle{URL: s.URL, Lang: s.Lang})
	}
	return result
}
