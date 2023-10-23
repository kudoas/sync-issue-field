FROM golang:1.21.3-alpine as builder

WORKDIR /app
COPY . /app

RUN go get -d -v
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -v -o app .

FROM alpine:latest

COPY --from=builder /app/app /app

ENTRYPOINT ["/app"]
