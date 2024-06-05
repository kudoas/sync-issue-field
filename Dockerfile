FROM golang:1.22.4-alpine@sha256:9bdd5692d39acc3f8d0ea6f81327f87ac6b473dd29a2b6006df362bff48dd1f8 as builder

WORKDIR /app
COPY . /app

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app ./cmd/run

FROM alpine:latest@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
