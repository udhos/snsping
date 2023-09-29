package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func pinger(app *application) {
	const me = "pinger"

	var countOk int
	var countErrors int

	for {
		errPub := publish(app)
		if errPub == nil {
			countOk++
		} else {
			countErrors++
		}
		if app.conf.debug {
			log.Printf("%s: success=%d error=%d: sleeping for %v",
				me, countOk, countErrors, app.conf.interval)
		}
		time.Sleep(app.conf.interval)
	}
}

func publish(app *application) error {
	const me = "publish"

	message := "sns-ping"

	input := &sns.PublishInput{
		Message:  aws.String(message),
		TopicArn: aws.String(app.conf.topicArn),
	}

	_, errPub := app.snsClient.Publish(context.TODO(), input)
	if errPub != nil {
		log.Printf("%s: error: %v",
			me, errPub)
		return errPub
	}

	return nil
}
