FROM golang:1.21.6-alpine@sha256:a6a7f1fcf12f5efa9e04b1e75020931a616cd707f14f62ab5262bfbe109aa84a as builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
