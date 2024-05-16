FROM golang:1.22.3-alpine@sha256:f1fe698725f6ed14eb944dc587591f134632ed47fc0732ec27c7642adbe90618 as builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
