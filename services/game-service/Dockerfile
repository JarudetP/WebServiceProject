FROM golang:alpine

WORKDIR /app

# Install build dependencies for any C-based image libraries if needed
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8083

CMD ["./main"]
