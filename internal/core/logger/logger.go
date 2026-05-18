package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Options for logger configuration
type Options struct {
	Debug    bool
	LogFile  string
	UseColor bool
}

// Init initializes the global logger
func Init(opts Options) error {
	var handler slog.Handler

	if opts.LogFile != "" {
		dir := filepath.Dir(opts.LogFile)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return err
		}
	}

	consoleHandler := NewConsoleHandler(os.Stdout, os.Stderr, opts)

	if opts.LogFile != "" {
		f, err := os.OpenFile(opts.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		handler = NewMultiHandler(consoleHandler, fileHandler)
	} else {
		handler = consoleHandler
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}

// ConsoleHandler handles UI output
type ConsoleHandler struct {
	out   io.Writer
	err   io.Writer
	opts  Options
	attrs []slog.Attr
	group string
}

func NewConsoleHandler(out, err io.Writer, opts Options) *ConsoleHandler {
	return &ConsoleHandler{
		out:  out,
		err:  err,
		opts: opts,
	}
}

func (h *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if h.opts.Debug {
		return true
	}
	return level >= slog.LevelInfo
}

func (h *ConsoleHandler) Handle(ctx context.Context, r slog.Record) error {
	w := h.out
	if r.Level >= slog.LevelError {
		w = h.err
	}

	level := r.Level.String()
	if r.Level >= slog.LevelError {
		level = "ERROR"
	} else if r.Level >= slog.LevelWarn {
		level = "WARN"
	} else if r.Level >= slog.LevelInfo {
		level = "INFO"
	} else {
		level = "DEBUG"
	}

	msg := fmt.Sprintf("[%s] %s", level, r.Message)
	
	_, err := io.WriteString(w, msg+"\n")
	return err
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ConsoleHandler{
		out:   h.out,
		err:   h.err,
		opts:  h.opts,
		attrs: append(h.attrs, attrs...),
		group: h.group,
	}
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		out:   h.out,
		err:   h.err,
		opts:  h.opts,
		attrs: h.attrs,
		group: name,
	}
}

// MultiHandler dispatches to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(handlers...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewMultiHandler(handlers...)
}

// Log levels
func Debug(msg string, args ...interface{}) {
	slog.Debug(fmt.Sprintf(msg, args...))
}

func Info(msg string, args ...interface{}) {
	slog.Info(fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	slog.Warn(fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...interface{}) {
	slog.Error(fmt.Sprintf(msg, args...))
}

func Log(ctx context.Context, level slog.Level, msg string, args ...interface{}) {
	slog.Log(ctx, level, fmt.Sprintf(msg, args...))
}

// LogWithTime logs with timestamp
func LogWithTime(level slog.Level, msg string, args ...interface{}) {
	slog.Log(context.Background(), level, fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), fmt.Sprintf(msg, args...)))
}
