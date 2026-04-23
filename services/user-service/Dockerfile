# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum if they exist
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o main .

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY .env* ./

EXPOSE 8081

CMD ["./main"]
