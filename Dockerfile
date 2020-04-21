FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o gimulator cmd/gimulator/main.go

FROM scratch
COPY --from=builder /build/gimulator /app/gimulator
WORKDIR /app
CMD ["./gimulator", "-ip=localhost:3030", "-config-file=/configs/roles.yaml"]
