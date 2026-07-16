package player

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/cam/cine-cli/internal/core"
)

type MPV struct {
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	mu      sync.Mutex
	running bool
	args    []string
}

func NewMPV(extraArgs []string) *MPV {
	return &MPV{args: extraArgs}
}

func (p *MPV) Name() string { return "mpv" }

func (p *MPV) Play(ctx context.Context, opts core.PlayOptions) error {
	p.mu.Lock()
	if p.running {
		p.stopLocked()
		time.Sleep(300 * time.Millisecond)
	}

	args := make([]string, 0)

	if opts.Referer != "" {
		args = append(args, "--referrer="+opts.Referer)
	}
	if opts.UserAgent != "" {
		args = append(args, "--user-agent="+opts.UserAgent)
	}
	if opts.Title != "" {
		args = append(args, "--title="+opts.Title)
	}
	if opts.PreferredLang != "" {
		args = append(args, "--alang="+opts.PreferredLang)
	}
	if opts.SubsLang != "" {
		args = append(args, "--slang="+opts.SubsLang)
	}

	args = append(args, p.args...)
	args = append(args, opts.StreamURL)

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	cmd := exec.CommandContext(ctx, "mpv", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	p.cmd = cmd
	p.running = true

	err := cmd.Start()
	if err != nil {
		p.running = false
		p.mu.Unlock()
		return fmt.Errorf("mpv start: %w", err)
	}
	p.mu.Unlock()

	go func() {
		cmd.Wait()
		p.mu.Lock()
		p.running = false
		p.mu.Unlock()
	}()

	time.Sleep(500 * time.Millisecond)

	if !p.Running() {
		return fmt.Errorf("mpv exited immediately — stream may be invalid or geo-blocked")
	}

	return nil
}

func (p *MPV) stopLocked() {
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (p *MPV) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
	return nil
}

func (p *MPV) Pause() error  { return nil }
func (p *MPV) Resume() error { return nil }
func (p *MPV) Position() (time.Duration, error) { return 0, nil }
func (p *MPV) Running() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
