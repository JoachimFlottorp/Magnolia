FROM jf/yeahapi.deps AS deps

FROM golang:1.19.1 AS collector
ENV GO111MODULE=on
WORKDIR /app
COPY . .
COPY --from=deps /app/protobuf /app/protobuf

RUN go mod download

RUN go build -v -ldflags '-extldflags "-static"' -o main /app/collector/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=deps /app/web /app/web
COPY --from=collector /app/main /app/main

RUN chmod +x /app/main

EXPOSE 50051

ENTRYPOINT ["/app/main"]