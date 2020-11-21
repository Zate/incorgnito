FROM golang:latest as builder
WORKDIR /go/src/github.com/zate/corgibot/
ADD main.go .
RUN GO111MODULE=on go mod init github.com/Zate/corgibot
RUN GO111MODULE=on go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o corgibot main.go

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/zate/corgibot/corgibot .

CMD ["/corgibot"]
