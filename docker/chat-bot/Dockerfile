FROM jf/magnolia.deps AS deps

FROM golang:1.19.2 AS builder
ENV GO111MODULE=on
ENV GOCACHE /go/cache
WORKDIR /src

COPY [ "go.mod", "go.sum", "./" ]
RUN go mod download

COPY . .
COPY --from=deps /src/protobuf /src/protobuf
COPY --from=deps /go/cache /go/cache

RUN go build -v -ldflags '-extldflags "-static"' -o server /src/cmd/chat-bot/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=builder /src/server /app/server

RUN chmod +x /app/server

ENTRYPOINT ["/app/server"]
