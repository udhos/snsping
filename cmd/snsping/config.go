package main

import (
	"time"

	"github.com/udhos/boilerplate/envconfig"
)

type config struct {
	interval     time.Duration
	healthAddr   string
	healthPath   string
	endpointURL  string
	topicArn     string
	topicRoleArn string
	debug        bool
	attributes   int
}

func getConfig(roleSessionName string) config {

	env := envconfig.NewSimple(roleSessionName)

	return config{
		interval:     env.Duration("INTERVAL", 10*time.Second),
		healthAddr:   env.String("HEALTH_ADDR", ":8888"),
		healthPath:   env.String("HEALTH_PATH", "/health"),
		topicArn:     env.String("TOPIC_ARN", ""),
		topicRoleArn: env.String("TOPIC_ROLE_ARN", ""),
		debug:        env.Bool("DEBUG", false),
		attributes:   env.Int("ATTRIBUTES", 1),
	}
}
