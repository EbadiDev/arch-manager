# syntax=docker/dockerfile:1

## Build
FROM golang:1.24.2-bookworm AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o arch-manager
RUN tar -zcf web.tar.gz web

## Deploy
FROM ghcr.io/ebadidev/debian:bookworm-slim

WORKDIR /app

COPY --from=build /app/arch-manager arch-manager
COPY --from=build /app/web.tar.gz web.tar.gz
COPY --from=build /app/resources/ed25519_public_key.txt resources/ed25519_public_key.txt
COPY --from=build /app/configs/main.defaults.json configs/main.defaults.json
COPY --from=build /app/storage/app/.gitignore storage/app/.gitignore
COPY --from=build /app/storage/database/.gitignore storage/app/.gitignore
COPY --from=build /app/storage/logs/.gitignore storage/logs/.gitignore
COPY --from=build /app/third_party/xray-linux-64/xray third_party/xray-linux-64/xray

RUN tar -xvf web.tar.gz
RUN rm web.tar.gz

ENTRYPOINT ["./arch-manager", "start"]
