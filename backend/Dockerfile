# Use the official Golang image as the base image
FROM golang:1.23-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first for dependency caching
COPY go.mod go.sum ./

# Download and install dependencies
RUN go mod download

# Copy the entire codebase into the container
COPY . .

# Build the Go app
RUN go build -o go-chat-app main.go

# Expose the port on which the app will run
EXPOSE 8080

# Command to run the executable
CMD ["./go-chat-app"]