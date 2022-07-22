package main

import (
	"flag"
	"fmt"
	"legocerthub-backend/pkg/domain/app"
	"log"

	"net/http"
	"time"
)

// http server config options
type webConfig struct {
	host string
	port int
}

func main() {
	// parse command line for config options
	var webCfg webConfig
	var appCfg app.Configuration

	flag.StringVar(&webCfg.host, "host", "localhost", "hostname to listen on")
	flag.IntVar(&webCfg.port, "port", 4050, "port number for API to listen on")
	flag.IntVar(&appCfg.Http01Port, "http01port", 4060, "port number for http01 challenge server to listen on")
	// TODO: change default to false
	flag.BoolVar(&appCfg.DevMode, "development", true, "run the server in development mode")
	flag.Parse()

	// configure the app
	app, err := app.CreateAndConfigure(appCfg)
	if err != nil {
		log.Panicf("panic: failed to configure app: %s", err)
	}
	defer app.CloseStorage()

	// configure webserver
	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	// allow longer timeouts when in development
	if appCfg.DevMode {
		readTimeout = 15 * time.Second
		writeTimeout = 30 * time.Second
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", webCfg.host, webCfg.port),
		Handler:      app.Routes(),
		IdleTimeout:  1 * time.Minute,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// launch webserver
	app.GetLogger().Infof("starting lego-certhub on %s:%d", webCfg.host, webCfg.port)
	err = srv.ListenAndServe()
	if err != nil {
		app.GetLogger().Panicf("panic: failed to start http server: %s", err)
	}
}
