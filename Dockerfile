# syntax=docker/dockerfile:1

# Stage 1: Build
FROM golang:1.24.3 AS builder

WORKDIR /app

# Copy go files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o rappelconsommation main.go

# Stage 2: Run
FROM alpine:latest

WORKDIR /app

# Copy the built binary and static assets
COPY --from=builder /app/rappelconsommation .
COPY templates/ templates/
COPY static/ static/

# Run the app
CMD ["./rappelconsommation"]
