# Use the official Go image as the base image
FROM golang:1.19 as builder

# Set the working directory in the container
WORKDIR /app

# Copy the application files into the working directory
COPY . /app

# Build the application
RUN go mod tidy
RUN go build -o main ./cmd/main.go

FROM ubuntu:latest

# Set the working directory in the container
WORKDIR /app

COPY --from=builder /app/main .

# Expose port 8080
EXPOSE 8080

# executable
ENTRYPOINT [ "./main" ]
