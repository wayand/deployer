# Step 1: Build the Go app
FROM golang:1.23-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Initialize Go modules (this only runs if go.mod doesn't already exist)
RUN go mod init github.com/wayand/deployer || true

# Download the dependencies (will work even if the go.mod already exists)
RUN go mod tidy

# Build the Go binary
RUN go build -o webhook .

# Step 2: Create a lightweight image to run the app
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app/

# Install bash, docker-cli, and Docker Compose
RUN apk --no-cache add git bash docker-cli curl 
# openssh

# Add GitHub to known hosts (to avoid "authenticity of host" prompt)
# RUN mkdir -p /root/.ssh && \
#     ssh-keyscan github.com >> /root/.ssh/known_hosts

# Install Docker Compose as a standalone binary
RUN curl -L "https://github.com/docker/compose/releases/download/v2.29.6/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

# Copy the built binary from the previous build stage
COPY --from=build /app/webhook .

# Expose port 9000 for the webhook listener
EXPOSE 9000

# Run the Go webhook application
CMD ["./webhook"]
