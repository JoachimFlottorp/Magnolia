FROM jf/magnolia.deps as deps

FROM node:18 AS builder
WORKDIR /src/markov-generator

COPY [ "markov-generator/package.json", "markov-generator/package-lock.json*", "./" ]

RUN npm ci

COPY ./markov-generator .
COPY --from=deps /src/markov-generator/src/protobuf /src/markov-generator/src/protobuf

ENV NODE_ENV=production

FROM denoland/deno:1.29.1
WORKDIR /app
COPY --from=builder /src/markov-generator /app

ENTRYPOINT [ "deno", "run", "--allow-read", "--allow-net", "src/index.ts", "/app/config.toml" ]
