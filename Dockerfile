FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 go build -o dove-dashboard ./cmd/dove-dashboard

FROM alpine:3

LABEL org.opencontainers.image.title="Dove Dashboard" \
      org.opencontainers.image.description="A lightweight and peaceful web-based system monitor written in Go." \
      org.opencontainers.image.source="https://github.com/PoProstuWitold/dove-dashboard" \
      org.opencontainers.image.url="https://hub.docker.com/r/poprostuwitold/dove-dashboard" \
      org.opencontainers.image.documentation="https://github.com/PoProstuWitold/dove-dashboard#readme" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="Witold Zawada (PoProstuWitold)"

RUN apk add --no-cache lm-sensors util-linux

RUN addgroup -g 10001 dove && adduser -D -u 10001 -G dove dove

USER dove

WORKDIR /home/dove

COPY --from=builder /app/dove-dashboard ./dove-dashboard
COPY --from=builder /app/internal/web ./web

EXPOSE 2137

ENTRYPOINT ["./dove-dashboard"]