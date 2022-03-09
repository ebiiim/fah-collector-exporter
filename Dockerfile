FROM golang:1.17-buster as builder
WORKDIR /go/src/app
COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 go build "-ldflags=-s -w" -trimpath -o main

FROM alpine:3.15
LABEL org.opencontainers.image.source https://github.com/ebiiim/fah-sidecar-exporter
COPY --from=builder /go/src/app/main .
ENTRYPOINT [ "./main" ]
