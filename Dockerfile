FROM golang:latest

WORKDIR /go
COPY  main.go .
RUN go get github.com/docker/docker/api/ && \
    go get github.com/prometheus/client_golang/prometheus && \
    GOOS=linux CGO_ENABLED=0 go build main.go

FROM alpine:latest
WORKDIR /go
COPY --from=0 /go/main .
CMD ["./main"]

