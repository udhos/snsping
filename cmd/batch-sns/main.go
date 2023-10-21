// Package main implements an utility.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/udhos/boilerplate/boilerplate"
	"github.com/udhos/opentelemetry-trace-sqs/otelsns"
	"github.com/udhos/snsping/internal/env"
	"github.com/udhos/snsping/internal/snsclient"
	"github.com/udhos/snsping/internal/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	topicArn    string
	roleArn     string
	endpointURL string
	wg          sync.WaitGroup
}

const version = "1.1.0"

const batch = 10

func main() {

	me := filepath.Base(os.Args[0])

	conf := &config{}

	var count int
	var writers int
	var showVersion bool

	flag.IntVar(&count, "count", 30, "total number of messages to send")
	flag.IntVar(&writers, "writers", 30, "number of concurrent writers")
	flag.StringVar(&conf.topicArn, "topicArn", "", "required topic ARN")
	flag.StringVar(&conf.roleArn, "roleArn", "", "optional role ARN")
	flag.StringVar(&conf.endpointURL, "endpointURL", "", "optional endpoint URL")
	flag.BoolVar(&showVersion, "version", showVersion, "show version")
	flag.Parse()

	{
		v := boilerplate.LongVersion(me + " version=" + version)
		if showVersion {
			fmt.Println(v)
			return
		}
		log.Print(v)
	}

	//
	// initialize tracing
	//

	jaegerURL := env.String("JAEGER_URL", "http://jaeger-collector:14268/api/traces")

	var tracer trace.Tracer

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

		tracer = tp.Tracer(me)
	}

	//
	// send
	//

	ctx, span := tracer.Start(context.TODO(), me)
	defer span.End()

	messages := []types.PublishBatchRequestEntry{}

	body := "batch-sns"

	for i := 0; i < batch; i++ {
		id := strconv.Itoa(i)
		m := types.PublishBatchRequestEntry{
			Message:           aws.String(body),
			Id:                aws.String(id),
			MessageAttributes: make(map[string]types.MessageAttributeValue),
		}

		otelsns.NewCarrier().Inject(ctx, m.MessageAttributes)

		messages = append(messages, m)
	}

	begin := time.Now()

	for i := 1; i <= writers; i++ {
		conf.wg.Add(1)
		go writer(i, writers, conf, count/writers, messages)
	}

	conf.wg.Wait()

	elap := time.Since(begin)

	rate := float64(count) / (float64(elap) / float64(time.Second))

	log.Printf("%s: sent=%d interval=%v rate=%v messages/sec",
		me, count, elap, rate)
}

func writer(id, total int, conf *config, count int, messages []types.PublishBatchRequestEntry) {
	defer conf.wg.Done()

	me := fmt.Sprintf("writer: [%d/%d]", id, total)

	log.Printf("%s: will send %d", me, count)

	begin := time.Now()

	const cooldown = 5 * time.Second

	snsClient := snsclient.NewSnsClient(me, conf.topicArn, conf.roleArn, conf.endpointURL)

	for sent := 0; sent < count; {
		input := &sns.PublishBatchInput{
			PublishBatchRequestEntries: messages,
			TopicArn:                   aws.String(conf.topicArn),
		}
		_, errPublish := snsClient.PublishBatch(context.TODO(), input)
		if errPublish != nil {
			log.Printf("%s: PublishBatch error: %v, sleeping %v",
				me, errPublish, cooldown)
			time.Sleep(cooldown)
			continue
		}
		sent += batch
	}

	elap := time.Since(begin)

	rate := float64(count) / float64(elap/time.Second)

	log.Printf("%s: sent=%d interval=%v rate=%v messages/sec",
		me, count, elap, rate)
}
