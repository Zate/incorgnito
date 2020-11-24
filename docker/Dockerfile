FROM golang:latest as builder
WORKDIR /go/src/github.com/zate/incorgnito/
ADD main.go bot.go purge.go ./
RUN go mod init github.com/Zate/incorgnito
RUN go mod download
RUN go mod vendor
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o incorgnito

FROM golang:latest
WORKDIR /
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/zate/incorgnito/incorgnito .
# COPY ChannelList.json .
CMD ["/incorgnito"]
