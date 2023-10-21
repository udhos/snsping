package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/udhos/opentelemetry-trace-sqs/otelsns"
	"gopkg.in/yaml.v3"
)

func pinger(app *application) {
	const me = "pinger"

	var topics []string

	if errYaml := yaml.Unmarshal([]byte(app.conf.topicArn), &topics); errYaml != nil {
		log.Fatalf("%s: parse yaml topics: %v", me, errYaml)
	}

	size := len(topics)

	if size < 1 {
		log.Fatalf("%s: bad number of topics: %d", me, size)
	}

	countOk := make([]int, size)
	countErrors := make([]int, size)

	for {
		for i, t := range topics {
			if errPub := publish(app, t); errPub == nil {
				countOk[i]++
			} else {
				countErrors[i]++
			}
			if app.conf.debug {
				log.Printf("%s: %s: success=%d error=%d",
					me, t, countOk[i], countErrors[i])
			}
		}
		if app.conf.debug {
			log.Printf("%s: sleeping for %v",
				me, app.conf.interval)
		}
		time.Sleep(app.conf.interval)
	}
}

func publish(app *application, topicArn string) error {
	const me = "publish"

	ctx, span := app.tracer.Start(context.TODO(), me)
	defer span.End()

	message := "snsping"

	input := &sns.PublishInput{
		Message:           aws.String(message),
		TopicArn:          aws.String(topicArn),
		MessageAttributes: make(map[string]types.MessageAttributeValue),
	}

	otelsns.NewCarrier().Inject(ctx, input.MessageAttributes)

	_, errPub := app.snsClient.Publish(context.TODO(), input)
	if errPub != nil {
		log.Printf("%s: error: %v",
			me, errPub)
		return errPub
	}

	return nil
}
