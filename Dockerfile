FROM golang:1.27.3-alpine@sha256:d200d2dd76b03a99b52651fd1325aff7dcd35422c35c01bac56571b0507f3deb as builder

WORKDIR /app

# 依存関係のレイヤーを分離
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
