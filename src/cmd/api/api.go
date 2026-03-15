package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"suscord/internal/app"
	"syscall"
	"time"

	"github.com/avast/retry-go/v4"
)

func main() {
	var (
		a   *app.App
		err error
	)

	err = retry.Do(func() error {
		a, err = app.NewApp()
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3), retry.Delay(time.Second))
	if err != nil {
		log.Fatalf("init new app error: %+v\n", err)
	}

	go func() {
		if err = a.RunApi(); err != nil {
			log.Fatalf("run api: %+v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.Shutdown(ctx)
}
