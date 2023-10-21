// Package main implements the rabbitping tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/udhos/snsping/internal/env"
	"github.com/udhos/snsping/internal/snsclient"
	"github.com/udhos/snsping/internal/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const version = "1.1.0"

type application struct {
	me           string
	conf         config
	serverHealth *http.Server
	snsClient    *sns.Client
	tracer       trace.Tracer
}

func longVersion(me string) string {
	return fmt.Sprintf("%s runtime=%s GOOS=%s GOARCH=%s GOMAXPROCS=%d",
		me, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
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
		v := longVersion(me + " version=" + version)
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

	app.snsClient = snsclient.NewSnsClient(me, app.conf.topicArn, app.conf.topicRoleArn, app.conf.endpointURL)

	//
	// initialize tracing
	//

	jaegerURL := env.String("JAEGER_URL", "http://jaeger-collector:14268/api/traces")

	{
		tp, errTracer := tracing.TracerProvider(me, jaegerURL)
		if errTracer != nil {
			log.Fatalf("tracer provider: %v", errTracer)
		}

		// Register our TracerProvider as the global so any imported
		// instrumentation in the future will default to using it.
		otel.SetTracerProvider(tp)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Cleanly shutdown and flush telemetry when the application exits.
		defer func(ctx context.Context) {
			// Do not make the application hang when it is shutdown.
			ctx, cancel = context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				log.Fatalf("trace shutdown: %v", err)
			}
		}(ctx)

		tracing.TracePropagation()

		app.tracer = tp.Tracer(me)
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

		mux.HandleFunc(app.conf.healthPath, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "health ok", 200)
		})

		go func() {
			log.Printf("health server: listening on %s %s", app.conf.healthAddr, app.conf.healthPath)
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
