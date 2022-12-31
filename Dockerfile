ARG DENO_VERSION=1.29.1
ARG PBKIT_VERSION=0.0.57

FROM denoland/deno:alpine-${DENO_VERSION} as deno

FROM golang:1.19.2-alpine3.16 as proto

# https://github.com/denoland/deno_docker/issues/240#issuecomment-1205550359 # 
COPY --from=deno /bin/deno /bin/deno
COPY --from=deno /usr/glibc-compat /usr/glibc-compat
COPY --from=deno /lib/* /lib/
COPY --from=deno /lib64/* /lib64/
COPY --from=deno /usr/lib/* /usr/lib/

RUN apk add --no-cache make protobuf bash curl git
ARG PBKIT_VERSION

RUN git clone -b "v$PBKIT_VERSION" https://github.com/pbkit/pbkit.git pbkit \
    && cd pbkit \
    && /bin/deno install -n pb -A --unstable --root / cli/pb/entrypoint.ts 

ENV GOMODULE111=on
WORKDIR /src
COPY ./protobuf ./protobuf

ENV PB_BINARY=/bin/pb

WORKDIR /src/protobuf
RUN /bin/deno run --allow-run --allow-read --allow-env --allow-write generate.ts

ARG DENO_VERSION
FROM denoland/deno:alpine-${DENO_VERSION} as deno_deps_base
WORKDIR /app
COPY markov-generator/src/deps.ts /app/deps.ts
RUN deno cache deps.ts

FROM golang:1.19.2-alpine3.16 as golang_deps_base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY --from=proto /src/protobuf /app/protobuf
COPY internal /app/internal
COPY cmd /app/cmd
COPY pkg /app/pkg
COPY external /app/external

FROM golang_deps_base as server_deps
COPY cmd/server/main.go /app/cmd/server/main.go
RUN cd cmd/server && go build -ldflags '-extldflags "-static"' -o ./out ./main.go

FROM alpine:3.16 as server
COPY --from=server_deps /app/cmd/server/out /app/server
ENTRYPOINT ["/app/server", "-config", "/app/config.toml"]

FROM golang_deps_base as twitch_reader_deps
COPY cmd/twitch-reader/main.go /app/cmd/twitch-reader/main.go
RUN cd cmd/twitch-reader && go build -ldflags '-extldflags "-static"' -o ./out ./main.go

FROM alpine:3.16 as twitch_reader
COPY --from=twitch_reader_deps /app/cmd/twitch-reader/out /app/twitch-reader
ENTRYPOINT ["/app/twitch-reader", "-config", "/app/config.toml"]

FROM golang_deps_base as chat_bot_deps
COPY cmd/chat-bot/main.go /app/cmd/chat-bot/main.go
RUN cd cmd/chat-bot && go build -ldflags '-extldflags "-static"' -o ./out ./main.go

FROM alpine:3.16 as chat_bot
COPY --from=chat_bot_deps /app/cmd/chat-bot/out /app/chat-bot
ENTRYPOINT ["/app/chat-bot", "-config", "/app/config.toml"]

FROM deno_deps_base as markov_generator
COPY markov-generator/src /app/src
COPY --from=proto /src/markov-generator/src/protobuf /app/src/protobuf
ENTRYPOINT [ "deno", "run", "--allow-read", "--allow-net", "/app/src/index.ts", "/app/config.toml" ]
