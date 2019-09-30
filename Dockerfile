FROM golang:latest

WORKDIR /go
ADD . /go
RUN  go get github.com/docker/docker/api/


CMD ["go", "run", "main.go"]
