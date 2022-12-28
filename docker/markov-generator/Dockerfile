FROM jf/magnolia.deps as deps

FROM node:18 AS builder
WORKDIR /src/markov-generator

COPY [ "markov-generator/package.json", "markov-generator/package-lock.json*", "./" ]

RUN npm ci

COPY ./markov-generator .
COPY --from=deps /src/protobuf/*.ts /src/markov-generator/src/protobuf/

ENV NODE_ENV=production

FROM node:18
WORKDIR /app
COPY --from=builder /src/markov-generator /app

ENTRYPOINT [ "node", "--no-warnings", "--loader", "ts-node/esm", "src/index.ts", "/app/config.toml" ]
