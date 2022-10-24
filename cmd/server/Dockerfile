FROM jf/magnolia.deps AS deps

FROM quay.io/goswagger/swagger:latest AS swag
WORKDIR /src
COPY . .
RUN make docs

FROM golang:1.19.2 AS builder
ENV GO111MODULE=on
WORKDIR /src
COPY . .
COPY --from=deps /src/protobuf /src/protobuf

RUN go mod download
RUN go build -v -ldflags '-extldflags "-static"' -o server /src/cmd/server/main.go

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates libc6-compat make

COPY --from=builder /src/server /app/server
COPY --from=swag /src/web/public/ /app/web/public/

RUN chmod +x /app/server

ENTRYPOINT ["/app/server"]
