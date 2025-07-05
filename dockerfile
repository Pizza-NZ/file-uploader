FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/server cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY config.yml .

EXPOSE 2131

ENTRYPOINT [ "/app/server" ]

CMD [ "-config", "/app/config.yml" ]