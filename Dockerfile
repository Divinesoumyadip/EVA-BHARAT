FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /ticket-system ./cmd/server

FROM alpine:3.20

RUN adduser -D -u 1000 appuser
COPY --from=build /ticket-system /usr/local/bin/ticket-system

USER appuser
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/ticket-system"]
