ARG RUST_VERSION=1.86.0
ARG APP_NAME=claim-submitter-btc-header
FROM rust:${RUST_VERSION}-slim-bullseye AS build
ARG APP_NAME
WORKDIR /app
# Install dependencies
RUN apt-get update && apt-get install -y \
    pkg-config \
    libssl-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=root ./base/rs ../../../base/rs
COPY --from=root ./generation/ ../../../generation/
COPY --from=root ./verification/ ../../../verification/
COPY . .
# Build the application
RUN --mount=type=bind,source=src,target=src \
    --mount=type=bind,source=Cargo.toml,target=Cargo.toml \
    --mount=type=bind,source=Cargo.lock,target=Cargo.lock \
    --mount=type=cache,target=/app/target/ \
    <<EOF
set -e
cargo build --release
cp ./target/release/$APP_NAME /bin/claim_submitter_btc_header
EOF
FROM debian:bullseye-slim AS final
ARG UID=10001
# Add certificates to final image
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
# Add non-root user
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser
COPY --from=build /bin/claim_submitter_btc_header /bin/
ENV RUST_LOG=info
CMD ["/bin/claim_submitter_btc_header"]
