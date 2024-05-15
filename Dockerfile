FROM golang:1.22.3-alpine@sha256:df479aa7d1298917c7de7fd56c7231ec450149f1e63d960ea96aebf9ff9240c5 as builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
