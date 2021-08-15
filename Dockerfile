FROM golang:buster as builder
# FROM xushikuan/alpine-build:2.0 AS builder

ENV GO111MODULE=on 
ENV GOOS=linux
ENV GOARCH=amd64
#ENV CGO_ENABLED=1

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o gimulator cmd/gimulator/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /build/gimulator gimulator

CMD ["./gimulator"]