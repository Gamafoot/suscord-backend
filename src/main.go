package main

import (
	"suscord/pkg/logger"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/multierr"
)

func main() {
	logger, cleanup, err := logger.NewSugaredLogger(logger.Config{
		FilePath:   "logs/app.log",
		MaxSizeMB:  100,
		MaxBackups: 5,
		MaxAgeDays: 30,
		Compress:   true,
		Level:      "info",
	})
	if err != nil {
		panic(err)
	}
	defer cleanup()

	err = handler()

	logger.Errorw("test", "erraaa", err)
}

func handler() error {
	err1 := pkgerr.WithStack(pkgerr.New("call ONE"))
	err2 := pkgerr.WithStack(pkgerr.New("call TWO"))
	return multierr.Append(err1, err2)
}
