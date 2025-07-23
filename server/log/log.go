package log

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var once sync.Once

var logger zerolog.Logger

func Logger() *zerolog.Logger {
	once.Do(func() {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return &logger
}