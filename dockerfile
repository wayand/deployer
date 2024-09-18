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

# Copy the built binary from the previous build stage
COPY --from=build /app/.env .
COPY --from=build /app/webhook .

# Expose port 9000 for the webhook listener
EXPOSE 9000

# Run the Go webhook application
CMD ["./webhook"]
