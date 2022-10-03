FROM quay.io/goswagger/swagger:latest as swag
WORKDIR /app
COPY . .
RUN make docs

FROM alpine:3.7 as proto
RUN apk add --no-cache make
WORKDIR /app
COPY . .
RUN make proto

FROM golang:1.18.1 as builder
ENV GO111MODULE=on
WORKDIR /src
COPY . .

WORKDIR /src
RUN go mod download

WORKDIR /src

RUN go build -v -ldflags '-extldflags "-static"' /src/cmd/server/main.go
RUN ["chmod", "a+x", "./main"]

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates libc6-compat make
COPY --from=builder /src/main /app/
COPY --from=proto /app/protobuf/collector/ /app/protobuf/collector/
COPY --from=swag /app/web/public/ /app/web/public/

EXPOSE 3003

ENTRYPOINT [ "/app/main", "-config", "./config.json" ]
