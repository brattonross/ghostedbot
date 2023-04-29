# build
FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.buildHash $(git rev-parse --short HEAD)" -o /ghostedbot

# test
FROM builder as test
RUN go test -v ./...

# release
FROM alpine:latest as release

WORKDIR /

COPY --from=builder /ghostedbot /ghostedbot

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/ghostedbot"]
