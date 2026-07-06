// Package logger builds the application-wide zap logger.
package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"messenger/internal/core/config"
)

// New creates a zap.Logger that writes JSON logs to both stdout and a rotating
// log file (well, an append-only file managed by the OS) under out/logs.
func New(cfg config.Log) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	if cfg.File != "" {
		if dir := filepath.Dir(cfg.File); dir != "" {
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				return nil, fmt.Errorf("create log dir %q: %w", dir, mkErr)
			}
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderCfg)

	syncers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if cfg.File != "" {
		file, openErr := os.OpenFile(cfg.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if openErr != nil {
			return nil, fmt.Errorf("open log file %q: %w", cfg.File, openErr)
		}
		syncers = append(syncers, zapcore.AddSync(file))
	}

	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(syncers...), level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}
