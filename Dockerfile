FROM golang:alpine as builder

ENV GO111MODULE=on GOOS=linux GOARCH=amd64

RUN apk add g++

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o gimulator cmd/gimulator/main.go

FROM busybox:musl

WORKDIR /app

COPY --from=builder /build/gimulator gimulator

CMD ["./gimulator"]
