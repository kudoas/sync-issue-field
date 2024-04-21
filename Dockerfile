FROM golang:1.22.2-alpine@sha256:cdc86d9f363e8786845bea2040312b4efa321b828acdeb26f393faa864d887b0 as builder

WORKDIR /app
COPY . /app

RUN go get -d -v
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

FROM alpine:latest@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
