FROM golang:1.12-stretch as builder
ADD . /go/src/github.com/iov-one/chatbot
WORKDIR /go/src/github.com/iov-one/chatbot/cmd/bot/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


FROM everpeace/kubectl
WORKDIR /app
COPY --from=builder /go/src/github.com/iov-one/chatbot/cmd/bot/app .
ENTRYPOINT ["./app"]
