FROM golang:1.24.1-alpine AS builder

COPY --from=root ./base/go ../../../../base/go
COPY ./ /opt/app/mock-rpc

WORKDIR /opt/app/mock-rpc

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl musl-dev
RUN go build -o main ./cmd/

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/mock-rpc/mock_block.json .
COPY --from=builder /opt/app/mock-rpc/mock_chain_config.json .
COPY --from=builder /opt/app/mock-rpc/mock_geth_witness.json .
COPY --from=builder /opt/app/mock-rpc/mock_raw_block.txt .
COPY --from=builder /opt/app/mock-rpc/mock_reth_witness.json .
COPY --from=builder /opt/app/mock-rpc/main .

ENTRYPOINT [ "./main" ]