// Package snsclient creates sns client.
package snsclient

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/udhos/boilerplate/awsconfig"
)

// NewSnsClient creates an SNS client.
func NewSnsClient(sessionName, topicArn, roleArn, endpointURL string) *sns.Client {
	const me = "NewSnsClient"

	topicRegion, errTopic := GetTopicRegion(topicArn)
	if errTopic != nil {
		log.Fatalf("%s: topic region error: %v", me, errTopic)
	}

	awsConfOptions := awsconfig.Options{
		Region:          topicRegion,
		RoleArn:         roleArn,
		RoleSessionName: sessionName,
		EndpointURL:     endpointURL,
	}

	awsConf, errAwsConf := awsconfig.AwsConfig(awsConfOptions)
	if errAwsConf != nil {
		log.Fatalf("%s: aws config error: %v", me, errAwsConf)
	}

	return sns.NewFromConfig(awsConf.AwsConfig)
}

// GetTopicRegion extracts region from SNS topic ARN.
// arn:aws:sns:us-east-1:123456789012:mytopic
func GetTopicRegion(topicArn string) (string, error) {
	const me = "getTopicRegion"
	fields := strings.SplitN(topicArn, ":", 5)
	if len(fields) < 5 {
		return "", fmt.Errorf("%s: bad topic arn=[%s]", me, topicArn)
	}
	region := fields[3]
	//log.Printf("%s: topicRegion=[%s]", me, region)
	return region, nil
}
