FROM golang:1.12-stretch as builder
ADD . /go/src/github.com/iov-one/chatbot
WORKDIR /go/src/github.com/iov-one/chatbot/cmd/bot/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:3.9

ARG KUBE_VERSION="1.11.1"

RUN apk add --update ca-certificates && \
    apk add --update -t deps curl && \
    curl -L https://storage.googleapis.com/kubernetes-release/release/v$KUBE_VERSION/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    apk del --purge deps && \
    rm /var/cache/apk/*
WORKDIR /app
COPY --from=builder /go/src/github.com/iov-one/chatbot/cmd/bot/app .
ENTRYPOINT ["./app"]
