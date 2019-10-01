FROM golang:latest

WORKDIR /go
ADD . /go
RUN  go get github.com/docker/docker/api/ && go get github.com/prometheus/client_golang/prometheus


CMD ["go", "run", "main.go"]
