# build
FROM golang:1.20-alpine as builder

ARG ldflags

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags $ldflags -o ghostedbot ./

# test
FROM builder as test
RUN CGO_ENABLED=0 go test -v ./...

# release
FROM alpine:latest as release

WORKDIR /app

COPY --from=builder /app/ghostedbot ./

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/ghostedbot"]
