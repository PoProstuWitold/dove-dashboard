FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN CGO_ENABLED=0 go build -o dovedashboard ./cmd/dovedashboard

FROM alpine:3

RUN apk add --no-cache lm-sensors

RUN addgroup -g 10001 dove && adduser -D -u 10001 -G dove dove

USER dove

WORKDIR /home/dove

COPY --from=builder /app/dovedashboard ./dovedashboard
COPY --from=builder /app/internal/web ./web

EXPOSE 2137

ENTRYPOINT ["./dovedashboard"]