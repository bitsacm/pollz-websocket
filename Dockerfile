FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy source code first
COPY . .

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

EXPOSE 1401

CMD ["./main"]