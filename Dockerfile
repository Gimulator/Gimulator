FROM golang:1-alpine as builder

ENV GOOS=linux GOARCH=amd64

RUN apk add g++

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -a -ldflags "-extldflags '-static -O3' -s -w" -o gimulator cmd/gimulator/main.go

FROM busybox

# Since BusyBox does not come with default & pre-defined CAs and certs, secure connections might not get validated and therefore, established.
# To resolve this, available CAs and certs will be copied to busybox.
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

WORKDIR /app
COPY --from=builder /build/gimulator gimulator

CMD ["./gimulator"]