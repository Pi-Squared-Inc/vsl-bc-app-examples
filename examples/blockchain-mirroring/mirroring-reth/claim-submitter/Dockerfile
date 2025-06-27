FROM rust:1.86.0-slim-bullseye AS build

ARG GITHUB_ACTOR
ARG GITHUB_TOKEN

ENV CARGO_NET_GIT_FETCH_WITH_CLI=true

# Install dependencies
RUN apt-get update && apt-get install -y pkg-config libssl-dev libclang-dev clang git

RUN git config --global url."https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

WORKDIR /opt/app/vsl-blockchain-app
# Copy dependencies
COPY --from=root ./base/rs ./base/rs
COPY --from=root ./generation/block-processing/evm/rs ./generation/block-processing/evm/rs

WORKDIR /opt/app/vsl-blockchain-app/examples/blockchain-mirroring/mirroring-reth/claim-submitter
RUN --mount=type=bind,source=src,target=src \
  --mount=type=bind,source=Cargo.toml,target=Cargo.toml \
  --mount=type=cache,target=/app/target/ \
  <<EOF
set -e
cargo build --release
cp ./target/release/mirroring-reth-claim-submitter /bin/server
EOF

FROM debian:bullseye-slim AS final

ARG UID=10001
RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid "${UID}" \
  appuser
USER appuser

COPY --from=build /bin/server /bin/

CMD ["/bin/server"]