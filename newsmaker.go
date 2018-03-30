package main

import (
	"os"
	"os/signal"

	"flag"

	"github.com/BurntSushi/toml"
	"github.com/dlepex/newsmaker/news"
	"go.uber.org/zap"
)

func main() {

	flag.Parse()

	log, _ := zap.NewDevelopment()
	news.SetLogger(log)
	slog := log.Sugar()
	cfgpath := flag.Arg(0)
	slog.Infow("starting newsmaker", "cfgpath", cfgpath)
	var conf config
	_, err := toml.DecodeFile(cfgpath, &conf)
	if err != nil {
		slog.Fatalf("config parse err: %s", err)
	}

	pl, ers := conf.newPipeline()

	if len(ers) != 0 {
		slog.Fatalf("pipeline config errors: %v", ers)
	}

	err = pl.Run()
	if err != nil {
		slog.Fatalf("pipeline start error: %s", err)
	}
	setupSignalHandlers(log)

	pl.Wait()
}

func setupSignalHandlers(l *zap.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			l.Sync() //nolint:errcheck
			os.Exit(0)
		}
	}()
}
