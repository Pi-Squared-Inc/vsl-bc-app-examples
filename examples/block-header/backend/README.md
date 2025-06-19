# Block Header Settlement Backend

This is the backend service for the block header settlement demo.

## Prerequisites

- [Golang](https://go.dev/doc/install): Version >= 1.24.1

## Getting Started

1. Prepare the environment variables

   ```bash
   cp sample.env .env
   ```

2. Install the dependencies

   ```bash
   go mod download
   ```

3. Start the backend service

   ```bash
   go run ./cmd/main.go
   ```
