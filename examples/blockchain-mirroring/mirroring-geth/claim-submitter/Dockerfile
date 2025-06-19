FROM golang:1.24.1-alpine AS builder

COPY --from=root ./base/go ../../../base/go
COPY --from=root ./generation/block-processing/evm/go ../../../generation/block-processing/evm/go
COPY --from=root ./verification/block-processing/evm/go ../../../verification/block-processing/evm/go
COPY ./ /opt/app/mirroring-geth

WORKDIR /opt/app/mirroring-geth

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl musl-dev
RUN go build ./cmd/main.go

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/mirroring-geth/main .

ENTRYPOINT [ "./main" ]