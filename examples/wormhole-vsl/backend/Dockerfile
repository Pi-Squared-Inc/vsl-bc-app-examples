FROM golang:1.23.1-alpine AS builder

COPY . /opt/app/backend

WORKDIR /opt/app/backend

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev
RUN go mod download
RUN go build -o backend ./cmd/main.go

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/backend/backend .

ENTRYPOINT [ "./backend" ]