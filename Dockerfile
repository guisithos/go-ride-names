# Use a different base image
FROM golang:1.21-bullseye

# Set the working directory
WORKDIR /app

# Set environment variables for Go modules and DNS
ENV GOPROXY=https://proxy.golang.org,direct
ENV GO111MODULE=on
ENV GOSUMDB=sum.golang.org
ENV GODEBUG=netdns=go
ENV DNS_RESOLVER=8.8.8.8
ENV PORT=8080

# Create a volume for token storage
VOLUME ["/app/tokens"]

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies with retry logic and specific DNS
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf && \
    echo "nameserver 8.8.4.4" >> /etc/resolv.conf && \
    for i in $(seq 1 3); do go mod download && break || sleep 5; done

# Create directories first
RUN mkdir -p /app/static/favicon /app/static/css /app/static/js /app/templates /app/tokens && \
    chmod 755 /app/tokens

# Copy the source code and static files
COPY . .
COPY static/ /app/static/
COPY templates/ /app/templates/

# Build the application
RUN go build -o /server ./cmd/main.go

# Expose port 8080
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/server"]