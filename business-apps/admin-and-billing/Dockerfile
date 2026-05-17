# Build stage for frontend
FROM oven/bun:1.3-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package files and install dependencies
COPY frontend/package.json frontend/package-lock.json* ./
RUN bun install

# Copy frontend source and build
COPY frontend/ ./
RUN bun run build

# Build stage for backend
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o admin-and-billing .

# Final runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy compiled backend binary
COPY --from=backend-builder /app/admin-and-billing .

# Copy frontend dist files
COPY --from=frontend-builder /app/dist ./dist

# Expose port
EXPOSE 8080

# Run the application
CMD ["./admin-and-billing"]
