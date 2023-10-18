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
	"github.com/udhos/snsping/snsclient"
)

type config struct {
	topicArn    string
	roleArn     string
	endpointURL string
	wg          sync.WaitGroup
}

const version = "1.0.0"

const batch = 10

func main() {

	me := filepath.Base(os.Args[0])

	conf := &config{}

	var count int
	var writers int
	var showVersion bool

	flag.IntVar(&count, "count", 10000, "total number of messages to send")
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

	messages := []types.PublishBatchRequestEntry{}

	body := "hello world"

	for i := 0; i < batch; i++ {
		id := strconv.Itoa(i)
		m := types.PublishBatchRequestEntry{
			Message: aws.String(body),
			Id:      aws.String(id),
		}
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
