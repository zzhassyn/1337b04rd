FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o 1337b04rd ./cmd/1337b04rd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/1337b04rd .
COPY --from=builder /app/template ./template
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./1337b04rd", "--port", "8080"]
