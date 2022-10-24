FROM quay.io/goswagger/swagger:latest as swag
WORKDIR /app
COPY . .
RUN make docs

FROM golang:1.19.1 as builder
ENV GO111MODULE=on
WORKDIR /app
COPY . .

RUN go mod download
RUN go build -v -ldflags '-extldflags "-static"' -o server /app/cmd/server/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=builder /app/server /app/server
COPY --from=swag /app/web /app/web

RUN chmod +x /app/server

ENTRYPOINT ["/app/server"]