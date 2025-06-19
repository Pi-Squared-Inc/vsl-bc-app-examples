FROM golang:1.23-alpine AS builder

COPY --from=root ./base/go /opt/app/base/go
COPY --from=root ./generation/block-processing/evm/go /opt/app/generation/block-processing/evm/go
COPY --from=root ./verification/block-processing/evm/go /opt/app/verification/block-processing/evm/go
COPY . /opt/app/examples/blockchain-mirroring/mirroring-geth/claim-verifier

WORKDIR /opt/app/examples/blockchain-mirroring/mirroring-geth/claim-verifier

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl musl-dev
RUN go build ./cmd/main.go

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/examples/blockchain-mirroring/mirroring-geth/claim-verifier/main .

ENTRYPOINT [ "./main" ]