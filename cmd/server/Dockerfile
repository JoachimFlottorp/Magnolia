FROM jf/magnolia.deps as deps

FROM quay.io/goswagger/swagger:latest as swag
WORKDIR /src
COPY . .
RUN make docs

FROM golang:1.19.2 as builder
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

RUN chmod +x /app/server

ENTRYPOINT ["/app/server"]