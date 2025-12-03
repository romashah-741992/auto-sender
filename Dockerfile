FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auto-sender ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/auto-sender .

ENV PORT=8080

EXPOSE 8080

CMD ["./auto-sender"]
