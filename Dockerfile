FROM golang:1.23.0-alpine@sha256:d0b31558e6b3e4cc59f6011d79905835108c919143ebecc58f35965bf79948f4 as builder

WORKDIR /app

# 依存関係のレイヤーを分離
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
