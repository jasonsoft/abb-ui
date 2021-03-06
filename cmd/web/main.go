package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jasonsoft/abb-ui/routers"

	"github.com/jasonsoft/abb-ui/config"
	"github.com/jasonsoft/log"
	"github.com/jasonsoft/napnap"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// unknown error
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown error: %v", err)
			}
			log.Errorf("unknown error: %v", err)
		}
	}()

	_config := config.Config()

	// set up the log
	log.SetAppID("abb-ui") // unique id for the app

	// set up the napnap
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)

	nap := napnap.New()
	nap.SetRender("./templates")
	nap.Use(napnap.NewHealth())
	nap.Use(napnap.NewStatic("./public")) // use working directory
	//nap.Use(abb.NewErrorHandlingMiddleware())
	nap.Use(routers.NewRouter())
	nap.Use(napnap.NewNotfoundMiddleware())
	httpEngine := napnap.NewHttpEngine(_config.Bind)
	log.Info("abb ui starting")
	go func() {
		// service connections
		err := nap.Run(httpEngine)
		if err != nil {
			log.Error(err)
		}
	}()

	<-stopChan
	log.Info("Shutting down server...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	httpEngine.Shutdown(ctx)

	log.Info("gracefully stopped")
}
