# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary in a minimal container
# -o /app/server builds the output to a specific path
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server/main.go

# Stage 2: Create the final, minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy the web assets to be served by the application
COPY ./web ./web

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
