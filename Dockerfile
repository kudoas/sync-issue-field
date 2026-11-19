FROM golang:1.27.4-alpine3.24@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

WORKDIR /app

# 依存関係のレイヤーを分離
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
