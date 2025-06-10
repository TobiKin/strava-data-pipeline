FROM golang:1.22-alpine as builder

# Install git and certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/strava-pipeline ./cmd/server/

# Use a small image for the final container
FROM alpine:latest

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary and config
COPY --from=builder /app/strava-pipeline /strava-pipeline
COPY --from=builder /app/config.yaml /config.yaml

# Use an unprivileged user
USER appuser:appuser

# Port exposure
EXPOSE 8080

# Run the application
ENTRYPOINT ["/strava-pipeline", "--config", "/config.yaml"]
