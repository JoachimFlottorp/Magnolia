FROM jf/magnolia.deps as deps

FROM node:16 AS builder
WORKDIR /src/markov-generator

COPY [ "markov-generator/package.json", "markov-generator/package-lock.json*", "./markov-generator/" ]

RUN npm ci

COPY ./markov-generator .
COPY --from=deps /src/protobuf/*.ts /src/markov-generator/src/protobuf/

ENV NODE_ENV=production

FROM alpine:latest
WORKDIR /app
COPY --from=builder /src/markov-generator /app

ENTRYPOINT [ "/app/node_modules/.bin/ts-node", "/app/src/index.ts" ]
