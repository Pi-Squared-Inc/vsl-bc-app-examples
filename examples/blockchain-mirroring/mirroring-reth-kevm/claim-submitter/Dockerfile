FROM ubuntu:jammy

ENV TZ=America/Chicago
ENV DEBIAN_FRONTEND=noninteractive

## Install dependecies
ARG JDK_VERSION=21
RUN apt-get update        \
  && apt-get upgrade -y           \
  && apt-get install -y           \
  bison                      \
  build-essential            \
  cmake                      \
  curl                       \
  debhelper                  \
  flex                       \
  gcc                        \
  git                        \
  gnupg                      \
  libboost-test-dev          \
  libbz2-dev                 \
  libcrypto++-dev            \
  libffi-dev                 \
  libfmt-dev                 \
  libgdbm-dev                \
  libgmp-dev                 \
  libjemalloc-dev            \
  libmpfr-dev                \
  libncurses5-dev            \
  libnss3-dev                \
  libreadline-dev            \
  libsecp256k1-dev           \
  libsqlite3-dev             \
  libssl-dev                 \
  libunwind-dev              \
  libyaml-dev                \
  libz3-dev                  \
  locales                    \
  lsb-release                \
  lsof                       \
  maven                      \
  openjdk-${JDK_VERSION}-jdk \
  parallel                   \
  pkg-config                 \
  python3                    \
  python3-dev                \
  python3-pip                \
  software-properties-common \
  wget                       \
  xxd                        \
  z3                         \
  zlib1g-dev                 \
  && add-apt-repository -y        \
  ppa:ethereum/ethereum      \
  && apt-get update               \
  && apt-get upgrade -y           \
  && apt-get install -y ethereum  \
  && apt-get clean                \
  && rm -rf /var/lib/apt/lists/*


# Configure Github authentication
ARG GITHUB_TOKEN
ENV GITHUB_TOKEN=${GITHUB_TOKEN}

ARG GITHUB_ACTOR
ENV GITHUB_ACTOR=${GITHUB_ACTOR}

RUN git config --global url."https://${GITHUB_ACTOR}:${GITHUB_TOKEN}@github.com/".insteadOf https://github.com/
ENV CARGO_NET_GIT_FETCH_WITH_CLI=true

## Install LLVM:
## Note: we need to install version 16 as later versions have a known bug
## affecting the code generator.
## Note: we try 2 different methods here because one or the other may not
## work depending on the distro.
ARG LLVM_VERSION=16
RUN (                                                 \
  apt-get update                               \
  && apt-get install -y                           \
  llvm-${LLVM_VERSION}-tools                 \
  clang-${LLVM_VERSION}                      \
  lldb-${LLVM_VERSION}                       \
  lld-${LLVM_VERSION}                        \
  ) || (                                           \
  wget https://apt.llvm.org/llvm.sh -O llvm.sh \
  && chmod +x llvm.sh                             \
  && ./llvm.sh ${LLVM_VERSION} all                \
  && apt-get install -y --no-install-recommends   \
  clang-${LLVM_VERSION}                      \
  lldb-${LLVM_VERSION}                       \
  lld-${LLVM_VERSION}                        \
  && rm -f llvm.sh                                \
  )

# After installing the particular clang version make the `clang` alias for it.
# It's needed for some sub-modules of `evm-semantics`,
# e.g. `kevm-pyk/src/kevm_pyk/kproj/plugin/deps/c-kzg-4844/src/Makefile`
RUN ln -sf /usr/bin/clang-${LLVM_VERSION} /usr/bin/clang \
  && ln -sf /usr/bin/clang++-${LLVM_VERSION} /usr/bin/clang++

# Install K
ARG KFRAMEWORK_REPO=github.com/runtimeverification/k.git
ARG KFRAMEWORK_VERSION=7.1.267
WORKDIR /app
RUN git clone --depth=1 --branch v${KFRAMEWORK_VERSION} https://${KFRAMEWORK_REPO} k
RUN cd k && git submodule update --init --recursive

RUN cd k && mvn package -Dhaskell.backend.skip -DskipTests
ENV PATH=/app/k/k-distribution/target/release/k/bin:$PATH

# Install Rust
ARG RUST_VERSION=1.86.0
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup install ${RUST_VERSION}
RUN rustup default ${RUST_VERSION}
RUN cargo install cargo-nextest --locked

WORKDIR /opt/pkg
# Retrieve KEVM
ARG KEVM_BRANCH=master
ARG KEVM_COMMIT=706c08858bc3068cbbd8a1503515bd04173c4780
ARG KEVM_REPO=github.com/Pi-Squared-Inc/evm-semantics.git
RUN git clone --branch ${KEVM_BRANCH} https://${KEVM_REPO} evm-semantics
RUN cd evm-semantics && git checkout ${KEVM_COMMIT} && git submodule update --init --recursive
ENV KEVM_DIR=/opt/pkg/evm-semantics

WORKDIR /opt/app/vsl-blockchain-app
# Copy dependencies
COPY --from=root ./base/rs ./base/rs
COPY --from=root ./generation/block-processing-kevm/rs ./generation/block-processing-kevm/rs

WORKDIR /opt/app/vsl-blockchain-app/examples/blockchain-mirroring/mirroring-reth-kevm/claim-submitter
RUN --mount=type=bind,source=src,target=src \
  --mount=type=bind,source=Cargo.toml,target=Cargo.toml \
  --mount=type=bind,source=Cargo.lock,target=Cargo.lock \
  --mount=type=cache,target=/app/target/ \
  --mount=type=cache,target=/usr/local/cargo/registry/ \
  <<EOF
set -e
RETH_CPPFLAGS=-DRETH_BUILD EXTRA_CPPFLAGS=-DEVM_ONLY cargo build --release
cp ./target/release/mirroring-reth-kevm-claim-submitter /opt/app/mirroring-reth-kevm-claim-submitter
EOF

RUN find /opt/app -name "libulmkllvm.so" -exec mv {} /usr/lib \;
RUN find /opt/app -name "libkevm.so" -exec mv {} /usr/lib \;

CMD ["/opt/app/mirroring-reth-kevm-claim-submitter"]