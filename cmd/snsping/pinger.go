package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/udhos/opentelemetry-trace-sqs/otelsns"
	"github.com/udhos/snsping/internal/snsclient"
	"go.opentelemetry.io/otel/trace"
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

	clientPerRegion := map[string]*sns.Client{}

	for {
		for i, t := range topics {
			region, errRegion := snsclient.GetTopicRegion(t)
			if errRegion != nil {
				log.Fatalf("%s: region error: %v", me, errRegion)
			}
			client := clientPerRegion[region]
			if client == nil {
				client = snsclient.NewSnsClient(me, app.conf.topicArn, app.conf.topicRoleArn, app.conf.endpointURL)
				clientPerRegion[region] = client
			}
			if errPub := publish(client, t, app.tracer, app.conf.attributes); errPub == nil {
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

func publish(client *sns.Client, topicArn string, tracer trace.Tracer, attributes int) error {
	const me = "publish"

	ctx, span := tracer.Start(context.TODO(), me)
	defer span.End()

	message := "snsping"

	input := &sns.PublishInput{
		Message:           aws.String(message),
		TopicArn:          aws.String(topicArn),
		MessageAttributes: make(map[string]types.MessageAttributeValue),
	}

	for i := 0; i < attributes; i++ {
		str := fmt.Sprintf("%d", i)
		input.MessageAttributes[str] = types.MessageAttributeValue{
			StringValue: aws.String(str),
			DataType:    aws.String("String"),
		}
	}

	otelsns.NewCarrier().Inject(ctx, input.MessageAttributes)

	_, errPub := client.Publish(context.TODO(), input)
	if errPub != nil {
		log.Printf("%s: error: %v",
			me, errPub)
		return errPub
	}

	return nil
}
