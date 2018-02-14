package news

import (
	"math/rand"
	"time"

	"go.uber.org/zap"
)

var slog *zap.SugaredLogger

func init() {
	SetLoggerDefault()
	rand.Seed(time.Now().UTC().UnixNano())
}

func SetLogger(l *zap.Logger) {
	if l == nil {
		l = zap.L()
	}
	slog = l.Sugar()
}

func SetLoggerDefault() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	SetLogger(l)
}

func SetLoggerVerbose() {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	SetLogger(l)
}
