# build
FROM golang:1.20-alpine as builder

ARG ldflags

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags $ldflags -o ghostedbot ./

# test
FROM builder as test
RUN go test -v ./...

# release
FROM alpine:latest as release

WORKDIR /app

COPY --from=builder /app/ghostedbot ./

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/ghostedbot"]
