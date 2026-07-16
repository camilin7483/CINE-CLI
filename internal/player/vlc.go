package player

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/cam/cine-cli/internal/core"
)

type VLC struct {
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	mu      sync.Mutex
	running bool
	args    []string
}

func NewVLC(extraArgs []string) *VLC {
	return &VLC{args: extraArgs}
}

func (p *VLC) Name() string { return "vlc" }

func (p *VLC) Play(ctx context.Context, opts core.PlayOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("player already running")
	}

	args := []string{
		"--play-and-exit",
		"--no-video-title-show",
	}

	if opts.Referer != "" {
		args = append(args, fmt.Sprintf("--http-referrer=%s", opts.Referer))
	}
	if opts.PreferredLang != "" {
		args = append(args, fmt.Sprintf("--audio-language=%s", opts.PreferredLang))
	}
	if opts.SubsLang != "" {
		args = append(args, fmt.Sprintf("--sub-language=%s", opts.SubsLang))
	}

	args = append(args, p.args...)
	args = append(args, opts.StreamURL)

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	cmd := exec.CommandContext(ctx, "vlc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil

	p.cmd = cmd
	p.running = true

	go func() {
		cmd.Run()
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}()

	return nil
}

func (p *VLC) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cancel != nil {
		p.cancel()
	}
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Signal(os.Interrupt)
	}
	return nil
}

func (p *VLC) Pause() error                     { return nil }
func (p *VLC) Resume() error                    { return nil }
func (p *VLC) Position() (time.Duration, error) { return 0, nil }
func (p *VLC) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

type PlayerSwitch struct {
	mpv *MPV
	vlc *VLC
}

func NewPlayerSwitch(mpvArgs, vlcArgs []string) *PlayerSwitch {
	return &PlayerSwitch{
		mpv: NewMPV(mpvArgs),
		vlc: NewVLC(vlcArgs),
	}
}

func (ps *PlayerSwitch) Play(ctx context.Context, opts core.PlayOptions) error {
	switch opts.Player {
	case "vlc":
		return ps.vlc.Play(ctx, opts)
	default:
		return ps.mpv.Play(ctx, opts)
	}
}

func (ps *PlayerSwitch) Stop() error {
	ps.mpv.Stop()
	ps.vlc.Stop()
	return nil
}

func (ps *PlayerSwitch) Name() string                     { return "player" }
func (ps *PlayerSwitch) Pause() error                     { return ps.mpv.Pause() }
func (ps *PlayerSwitch) Resume() error                    { return ps.mpv.Resume() }
func (ps *PlayerSwitch) Position() (time.Duration, error) { return ps.mpv.Position() }
func (ps *PlayerSwitch) Running() bool                    { return ps.mpv.Running() || ps.vlc.Running() }
