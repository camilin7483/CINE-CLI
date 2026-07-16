package cli

import "github.com/charmbracelet/lipgloss"

type styles struct {
	title      lipgloss.Style
	item       lipgloss.Style
	sel        lipgloss.Style
	dim        lipgloss.Style
	err        lipgloss.Style
	info       lipgloss.Style
	key        lipgloss.Style
	border     lipgloss.Style
	rating     lipgloss.Style
	subtitle   lipgloss.Style
	progress   lipgloss.Style
	progressBg lipgloss.Style
	sidebar    lipgloss.Style
	sidebarSel lipgloss.Style
	panel      lipgloss.Style
	panelTitle lipgloss.Style
	spinner    lipgloss.Style
	success    lipgloss.Style
	warn       lipgloss.Style
	helpKey    lipgloss.Style
	helpDesc   lipgloss.Style
	badge      lipgloss.Style
	separator  lipgloss.Style
}

var s styles

func initStyles(theme string) {
	if theme == "light" {
		s = lightTheme()
	} else {
		s = darkTheme()
	}
}

func darkTheme() styles {
	return styles{
		title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")).Padding(0, 1),
		item:       lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E2E8F0")),
		sel:        lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED")),
		dim:        lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")),
		err:        lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true),
		info:       lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")),
		key:        lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true),
		border:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7C3AED")),
		rating:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")),
		subtitle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA")),
		progress:   lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")),
		progressBg: lipgloss.NewStyle().Foreground(lipgloss.Color("#334155")),
		sidebar:    lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#CBD5E1")).Background(lipgloss.Color("#1E293B")),
		sidebarSel: lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#7C3AED")).Bold(true),
		panel:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#475569")).Padding(0, 1),
		panelTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F1F5F9")).Padding(0, 1),
		spinner:    lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true),
		success:    lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Bold(true),
		warn:       lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true),
		helpKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Bold(true),
		helpDesc:   lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0")),
		badge:      lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Background(lipgloss.Color("#1E1B4B")).Padding(0, 1).Bold(true),
		separator:  lipgloss.NewStyle().Foreground(lipgloss.Color("#475569")),
	}
}

func lightTheme() styles {
	return styles{
		title:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6D28D9")).Padding(0, 1),
		item:       lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#1E293B")),
		sel:        lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#6D28D9")),
		dim:        lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")),
		err:        lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626")).Bold(true),
		info:       lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")),
		key:        lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")).Bold(true),
		border:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6D28D9")),
		rating:     lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")),
		subtitle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED")),
		progress:   lipgloss.NewStyle().Foreground(lipgloss.Color("#6D28D9")),
		progressBg: lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1")),
		sidebar:    lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#1E293B")).Background(lipgloss.Color("#E2E8F0")),
		sidebarSel: lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#6D28D9")).Bold(true),
		panel:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#CBD5E1")).Padding(0, 1),
		panelTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1E293B")).Padding(0, 1),
		spinner:    lipgloss.NewStyle().Foreground(lipgloss.Color("#6D28D9")).Bold(true),
		success:    lipgloss.NewStyle().Foreground(lipgloss.Color("#16A34A")).Bold(true),
		warn:       lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")).Bold(true),
		helpKey:    lipgloss.NewStyle().Foreground(lipgloss.Color("#D97706")).Bold(true),
		helpDesc:   lipgloss.NewStyle().Foreground(lipgloss.Color("#1E293B")),
		badge:      lipgloss.NewStyle().Foreground(lipgloss.Color("#6D28D9")).Background(lipgloss.Color("#DDD6FE")).Padding(0, 1).Bold(true),
		separator:  lipgloss.NewStyle().Foreground(lipgloss.Color("#CBD5E1")),
	}
}

var camiloDevStyle = lipgloss.NewStyle().
	Bold(true).
	Italic(true).
	Foreground(lipgloss.Color("#A78BFA")).
	Background(lipgloss.Color("#1E1B4B")).
	Padding(0, 2)
