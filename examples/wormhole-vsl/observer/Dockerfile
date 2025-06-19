FROM golang:1.23.1-alpine AS builder

COPY --from=root ./base/go /opt/app/base/go
COPY --from=root ./generation/view-fn/evm/go /opt/app/generation/view-fn/evm/go
COPY . /opt/app/examples/wormhole/observer

WORKDIR /opt/app/examples/wormhole/observer

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev
RUN go mod download
RUN go build -o observer ./cmd/main.go

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/examples/wormhole/observer/observer .

ENTRYPOINT [ "./observer" ]