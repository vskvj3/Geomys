# Use an official lightweight Go image as the base
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the server executable
RUN go version
RUN go build -v -o server ./cmd/server/main.go

# Build the client executable (optional)
# RUN go build -o client ./cmd/client/main.go

# Create a minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy built executables from the builder stage
COPY --from=builder /app/server /app/server
# COPY --from=builder /app/client /app/client

# Expose the default port
EXPOSE 5000 6000

# Default command to run the server
CMD ["./server"]
