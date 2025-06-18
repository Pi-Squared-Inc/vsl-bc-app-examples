FROM rust:1.85-bullseye AS builder


COPY --from=root ./base/rs /opt/app/base/rs
COPY . /opt/app/examples/wormhole/relayer

WORKDIR /opt/app/examples/wormhole/relayer

RUN apt update
RUN apt install -y git build-essential clang

RUN cargo build --release

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /opt/app/examples/wormhole/relayer/target/release/relayer ./relayer
COPY ./sample.env ./.env

RUN apt update
RUN apt install -y ca-certificates

RUN chmod +x relayer

ENTRYPOINT ["./relayer"]