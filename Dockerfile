FROM golang:latest as builder
WORKDIR /go/src/github.com/zate/incorgnito/
ADD main.go .
RUN go mod init github.com/Zate/incorgnito
RUN GO111MODULE=on go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o incorgnito main.go

FROM scratch
WORKDIR /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/zate/incorgnito/incorgnito .

CMD ["/corgibot"]
