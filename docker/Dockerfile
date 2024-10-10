# Start from a Golang base image
FROM golang:1.23.1-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o bstore .

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates openssl

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/bstore .

# Create .bstore directory
RUN mkdir -p /root/.bstore

# Copy configuration file
COPY conf.yml /root/.bstore/conf.yml

# Expose the port the app runs on
EXPOSE 8080

# Environment variables for keys (will be provided at runtime)
ENV BSTORE_READ_WRITE_KEY=""
ENV BSTORE_ENC_KEY=""

# Command to run the executable
CMD ["./bstore"]
