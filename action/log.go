package main

import (
	"os"

	"github.com/rs/zerolog"
)

func newLogger(cfg *Config) zerolog.Logger {
	return zerolog.New(
		zerolog.NewConsoleWriter(
			func(w *zerolog.ConsoleWriter) {
				w.Out = os.Stdout
			},
		),
	).
		With().
		Timestamp().
		Str("github-repo", cfg.GithubRepository).
		Str("ref-name", cfg.GithubRefName).
		Str("provider-name", cfg.TFProviderName).
		Str("provider-version", cfg.providerVersion()).
		Logger()
}
