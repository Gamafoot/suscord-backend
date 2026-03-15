package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	FilePath   string
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	Level      string
}

func NewSugaredLogger(cfg Config) (*zap.SugaredLogger, func(), error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileWriter, level),
		zapcore.NewCore(jsonEncoder, consoleWriter, level),
	)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger.Sugar(), func() {
		_ = logger.Sync()
	}, nil
}
