package news

import (
	"go.uber.org/zap"
)

// slog - package global "sugared" logger.
// hopefully, it is safe to modify it before any code that sees slog actually runs
var slog *zap.SugaredLogger

func init() {
	SetLoggerDefault()
}

func SetLogger(l *zap.Logger) { //nolint:golint
	if l == nil {
		l = zap.L()
	}
	slog = l.Sugar()
}

func SetLoggerDefault() { //nolint:golint
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	SetLogger(l)
}

func SetLoggerVerbose() { //nolint:golint
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	SetLogger(l)
}
