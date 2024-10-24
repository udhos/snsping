// Package main implements the rabbitping tool.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/otelconfig/oteltrace"
	"go.opentelemetry.io/otel/trace"
)

const version = "1.2.7"

type application struct {
	me           string
	conf         config
	serverHealth *http.Server
	tracer       trace.Tracer
}

func main() {

	//
	// parse cmd line
	//

	var showVersion bool
	flag.BoolVar(&showVersion, "version", showVersion, "show version")
	flag.Parse()

	//
	// show version
	//

	me := filepath.Base(os.Args[0])

	{
		v := boilerplate.LongVersion(me + " version=" + version)
		if showVersion {
			fmt.Println(v)
			return
		}
		log.Print(v)
	}

	//
	// application
	//

	app := &application{
		me:   me,
		conf: getConfig(me),
	}

	//
	// initialize tracing
	//

	{
		options := oteltrace.TraceOptions{
			DefaultService:     me,
			NoopTracerProvider: false,
			Debug:              true,
		}

		tracer, cancel, errTracer := oteltrace.TraceStart(options)

		if errTracer != nil {
			log.Fatalf("tracer: %v", errTracer)
		}

		defer cancel()

		app.tracer = tracer
	}

	//
	// start health server
	//

	{
		mux := http.NewServeMux()
		app.serverHealth = &http.Server{
			Addr:    app.conf.healthAddr,
			Handler: mux,
		}

		mux.HandleFunc(app.conf.healthPath, func(w http.ResponseWriter,
			_ /*r*/ *http.Request) {
			http.Error(w, "health ok", 200)
		})

		go func() {
			log.Printf("health server: listening on %s %s",
				app.conf.healthAddr, app.conf.healthPath)
			err := app.serverHealth.ListenAndServe()
			log.Fatalf("health server: exited: %v", err)
		}()
	}

	//
	// start pinger
	//

	go pinger(app)

	<-make(chan struct{}) // wait forever
}
