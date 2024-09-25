FROM golang:1.23.0-alpine@sha256:d0b31558e6b3e4cc59f6011d79905835108c919143ebecc58f35965bf79948f4 as builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:7eccd4d5d574a3c889b0a6d97b2cdd0308c8e1afc2bba8d467c2b87d879b0c1c

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
