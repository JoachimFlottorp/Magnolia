FROM jf/magnolia.deps as deps

FROM node:16 AS builder
WORKDIR /src

COPY [ "package.json", "package-lock.json*", "./" ]

RUN npm ci

COPY . .
COPY --from=deps /src/protobuf /src/markov-generator/src/protobuf

ENV NODE_ENV=production

FROM alpine:latest
WORKDIR /app
COPY --from=builder /src /app

ENTRYPOINT [ "/app/node_modules/.bin/ts-node", "/app/markov-generator/src/index.ts" ]
