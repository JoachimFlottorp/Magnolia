FROM quay.io/goswagger/swagger:latest as swag
WORKDIR /app
COPY . .
RUN make docs

FROM golang:1.19.1-alpine3.16 as proto
RUN apk add --no-cache make protobuf
WORKDIR /app
COPY . .
RUN make proto

FROM golang:1.18.1 as builder
ENV GO111MODULE=on
WORKDIR /src
COPY . .

WORKDIR /src
RUN go mod download

RUN go build -v -ldflags '-extldflags "-static"' /src/cmd/server/main.go

from golang:1.18.1 as collector
ENV GO111MODULE=on
WORKDIR /src
COPY . .

WORKDIR /src
RUN go mod download

RUN go build -v -ldflags '-extldflags "-static"' /src/collector/main.go

FROM alpine:latest
WORKDIR /app

COPY ./entrypoint/ /app/entrypoint

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=builder /src/main /app/
COPY --from=proto /app/protobuf/collector/ /app/protobuf/collector/
COPY --from=swag /app/web/public/ /app/web/public/
COPY --from=collector /src/main /app/collector

RUN chmod +x /app/entrypoint/server.sh
RUN chmod +x /app/entrypoint/collector.sh
