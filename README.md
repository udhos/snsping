[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/snsping/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/snsping)](https://goreportcard.com/report/github.com/udhos/snsping)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/snsping.svg)](https://pkg.go.dev/github.com/udhos/snsping)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/snsping)](https://artifacthub.io/packages/search?repo=snsping)
[![Docker Pulls snsping](https://img.shields.io/docker/pulls/udhos/snsping)](https://hub.docker.com/r/udhos/snsping)

# snsping

# Build

```
git clone https://github.com/udhos/snsping

cd snsping

go install ./...
```

# Run

```
# optional
export DEBUG=true
export TOPIC_ROLE_ARN=arn:aws:iam::100010001000:role/publisher

# mandatory
export TOPIC_ARN=arn:aws:sns:us-east-1:100010001000:topicname

snsping
```

# Docker images

https://hub.docker.com/r/udhos/snsping


# Helm chart

https://udhos.github.io/snsping
