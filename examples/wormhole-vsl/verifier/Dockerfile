FROM golang:1.23-alpine AS builder

COPY --from=root ./base/go /opt/app/base/go
COPY --from=root ./generation/view-fn/evm/go /opt/app/generation/view-fn/evm/go
COPY --from=root ./verification/view-fn/evm/go /opt/app/verification/view-fn/evm/go
COPY . /opt/app/examples/wormhole/verifier

WORKDIR /opt/app/examples/wormhole/verifier

ENV CGO_ENABLED=1

RUN apk add --no-cache gcc musl-dev
RUN go mod download
RUN go build -o verifier ./cmd/main.go

FROM alpine:latest

WORKDIR /opt/app

COPY --from=builder /opt/app/examples/wormhole/verifier/verifier .

ENTRYPOINT [ "./verifier" ]