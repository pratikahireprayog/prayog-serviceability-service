FROM golang:1.24.2-alpine

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o serviceability ./cmd/api

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./serviceability"] 