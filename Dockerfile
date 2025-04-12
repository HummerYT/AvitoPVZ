FROM golang:1.23-alpine AS builder

WORKDIR /usr/local/src

COPY ["../go.mod", "../go.sum", "./"]
RUN go mod tidy
RUN go mod download

COPY .. /usr/local/src

RUN go build -o ./bin/app cmd/main.go

FROM alpine:latest AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY /internal/migrations /internal/migrations
COPY /test/integration /test/integration

CMD ["/app"]