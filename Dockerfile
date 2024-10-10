# Build stage
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code to the container
COPY . .

# Build the Go service
RUN go build -o hub-service

# Final stage: minimal image
FROM alpine:3.15

# Set the working directory
WORKDIR /app

# Copy the Go binary from the build stage
COPY --from=builder /app/hub-service .

# Expose the port your service listens on (if applicable)
EXPOSE 8080

# Run the service
CMD ["./hub-service", "api"]