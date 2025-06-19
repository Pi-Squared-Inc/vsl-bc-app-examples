#!/usr/bin/env bash
set -euo pipefail

# Configuration
JDK_VERSION=21
LLVM_VERSION=16
KFRAMEWORK_REPO="https://github.com/runtimeverification/k.git"
KFRAMEWORK_VERSION="7.1.267"
RUST_VERSION="1.86.0"
KEVM_BRANCH="master"
KEVM_COMMIT="706c08858bc3068cbbd8a1503515bd04173c4780"
KEVM_REPO="https://github.com/Pi-Squared-Inc/evm-semantics.git"

export DEBIAN_FRONTEND=noninteractive
export TZ=America/Chicago

# non-interactive mode
NON_INTERACTIVE=0
for arg in "$@"; do
  if [[ "$arg" == "-y" ]]; then
    NON_INTERACTIVE=1
  fi
done

# Function to ask for confirmation
ask_confirm() {
  local prompt="$1"
  if [[ $NON_INTERACTIVE -eq 1 ]]; then
    echo "y"
  else
    read -p "$prompt" confirm
    echo "$confirm"
  fi
}

# Function to ask for input with a prompt and a default value
ask_input() {
  local prompt="$1"
  if [[ $NON_INTERACTIVE -eq 1 ]]; then
    echo ""
  else
    read -p "$prompt" input
    echo "$input"
  fi
}

# Save the current working directory
CURRENT_DIR=$(pwd)

echo "Setting up KReth..."
echo "============================="
confirm=$(ask_confirm ">>> Checking and installing LLVM version ${LLVM_VERSION} (which will be set to default) if needed. Do you want to continue? [y/N]: ")
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Exiting setup."
  exit 0
fi

# Install LLVM
## Check if clang and clang++ version 16 are already installed and are the default versions
CLANG_VERSION_OK=false
if command -v clang >/dev/null 2>&1 && command -v clang++ >/dev/null 2>&1; then
  CLANG_VER=$(clang --version | head -n1 | grep -oE '[0-9]+\.[0-9]+' | head -n1 | cut -d. -f1)
  CLANGXX_VER=$(clang++ --version | head -n1 | grep -oE '[0-9]+\.[0-9]+' | head -n1 | cut -d. -f1)
  if [[ "$CLANG_VER" == "16" && "$CLANGXX_VER" == "16" ]]; then
    CLANG_VERSION_OK=true
  fi
fi
## If clang and clang++ version 16 are already installed, skip LLVM installation
## Otherwise, install LLVM version 16
if $CLANG_VERSION_OK; then
  echo "✅ clang and clang++ version 16 are already installed. Skipping LLVM installation."
else
  echo "Installing LLVM ${LLVM_VERSION}..."
  echo "===========START============="
  if ! sudo apt-get install -y llvm-${LLVM_VERSION}-tools clang-${LLVM_VERSION} lldb-${LLVM_VERSION} lld-${LLVM_VERSION}; then
    wget https://apt.llvm.org/llvm.sh -O llvm.sh
    chmod +x llvm.sh
    sudo ./llvm.sh ${LLVM_VERSION} all
    sudo apt-get install -y --no-install-recommends clang-${LLVM_VERSION} lldb-${LLVM_VERSION} lld-${LLVM_VERSION}
    rm -f llvm.sh
  fi
  ## Set clang aliases
  sudo ln -sf /usr/bin/clang-${LLVM_VERSION} /usr/bin/clang
  sudo ln -sf /usr/bin/clang++-${LLVM_VERSION} /usr/bin/clang++
  echo "============END=============="
fi


# Check if the user wants to continue with Rust installation
confirm=$(ask_confirm ">>> Checking and installing Rust version ${RUST_VERSION} (which will be set to default) if needed. Do you want to continue? [y/N]: ")
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Exiting setup."
  exit 0
fi

# Install Rust
## Check if rustc and cargo are already installed and have the correct version
RUSTC_OK=false
if command -v rustc >/dev/null 2>&1 && command -v cargo >/dev/null 2>&1; then
  INSTALLED_RUSTC_VERSION=$(rustc --version | awk '{print $2}')
  INSTALLED_CARGO_VERSION=$(cargo --version | awk '{print $2}')
  if [[ "$INSTALLED_RUSTC_VERSION" == "$RUST_VERSION" && "$INSTALLED_CARGO_VERSION" == "$RUST_VERSION" ]]; then
    RUSTC_OK=true
  fi
fi
## If they are, skip installation
## Otherwise, install Rust using rustup and set the default version
if $RUSTC_OK; then
  echo "✅ rustc and cargo version $RUST_VERSION are already installed. Skipping Rust installation."
else
  echo "Installing Rust ${RUST_VERSION}..."
  echo "===========START============="
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
  export PATH="$HOME/.cargo/bin:${PATH}"
  rustup install ${RUST_VERSION}
  rustup default ${RUST_VERSION}
  source $HOME/.bashrc
  echo "============END=============="
fi
## Install cargo-nextest
## Check if cargo-nextest is already installed
## If it is, skip installation
## Otherwise, install cargo-nextest using cargo
echo "Checking for cargo-nextest..."
if command -v cargo-nextest >/dev/null 2>&1; then
  echo "✅ cargo-nextest is already installed. Skipping installation."
else
  echo "Installing cargo-nextest..."
  cargo install cargo-nextest --locked
fi
export CARGO_NET_GIT_FETCH_WITH_CLI=true

# Check if the user wants to continue with K and KEVM installation
confirm=$(ask_confirm ">>> Checking and installing K and KEVM if needed. Do you want to continue? [y/N]: ")
if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
  echo "Exiting setup."
  exit 0
fi

# Update and install prerequisites for K and KEVM
echo "Installing prerequisites..."
echo "===========START============="
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y \
  bison build-essential cmake curl debhelper flex gcc git gnupg \
  libboost-test-dev libbz2-dev libcrypto++-dev libffi-dev libfmt-dev \
  libgdbm-dev libgmp-dev libjemalloc-dev libmpfr-dev libncurses5-dev \
  libnss3-dev libreadline-dev libsecp256k1-dev libsqlite3-dev libssl-dev \
  libunwind-dev libyaml-dev libz3-dev locales lsb-release lsof maven \
  openjdk-${JDK_VERSION}-jdk parallel pkg-config python3 python3-dev \
  python3-pip software-properties-common wget xxd z3 zlib1g-dev
## Add Ethereum PPA
sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install -y ethereum
## Clean apt cache
sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/*
echo "============END=============="


# Ask user for installation directory for K framework and KEVM
echo "Please provide absolute path for the installation directory."
INSTALL_DIR=$(ask_input ">>> Enter installation directory for K framework and KEVM [default: $HOME/pkg/]: ")
INSTALL_DIR="${INSTALL_DIR:-$HOME/pkg/}"
# Ensure the installation directory exists
mkdir -p "$INSTALL_DIR"

# Install K framework (manual build)
## Check if the K framework directory already exists in $INSTALL_DIR/k
## If it exists, skip installation and prompt the user to ensure the correct version is installed
## Otherwise, clone the K framework repository, build it, and add it to the PATH
if [ -d "$INSTALL_DIR/k" ]; then
  echo "⚠️ Directory ${INSTALL_DIR}/k already exists."
  echo "Please manually ensure the correct version, i.e., ${KFRAMEWORK_VERSION}, of K is installed."
  echo "Please also ensure that you have added the K binary to your PATH, i.e., add" 
  echo "export PATH=\"${INSTALL_DIR}/k/k-distribution/target/release/k/bin:\$PATH\""
  echo "to your ~/.bashrc file."
  confirm=$(ask_confirm ">>> Have you done this? [y/N]: ")
  if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Please ensure the correct version of K is installed and added to your PATH."
    echo "Exiting setup."
    exit 0
  fi
else
  echo "Installing K framework version ${KFRAMEWORK_VERSION} in ${INSTALL_DIR}..."
  echo "===========START============="
  mkdir -p "$INSTALL_DIR" && cd "$INSTALL_DIR"
  git clone --depth=1 --branch v${KFRAMEWORK_VERSION} ${KFRAMEWORK_REPO} k
  cd k && git submodule update --init --recursive
  mvn package -Dhaskell.backend.skip -DskipTests
  if ! grep -Fxq "export PATH=\"${INSTALL_DIR}/k/k-distribution/target/release/k/bin:\$PATH\"" "$HOME/.bashrc"; then
    echo "export PATH=\"${INSTALL_DIR}/k/k-distribution/target/release/k/bin:\$PATH\"" >> "$HOME/.bashrc"
    source "$HOME/.bashrc"
  fi
  echo "============END=============="
fi


# Clone KEVM
if [ -d "$INSTALL_DIR/evm-semantics" ]; then
  echo "⚠️ Directory ${INSTALL_DIR}/evm-semantics already exists."
  echo "Please manually ensure the correct branch (${KEVM_BRANCH}) and commit (${KEVM_COMMIT}) of KEVM is installed."
  echo "Please also ensure that you have added the KEVM_DIR to your environment, i.e., add"
  echo "export KEVM_DIR=\"${INSTALL_DIR}/evm-semantics\" to your ~/.bashrc file."
  confirm=$(ask_confirm ">>> Have you done this? [y/N]: ")
  if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Please ensure the correct branch and commit of KEVM is installed and added to your environment."
    echo "Exiting setup."
    exit 0
  fi
else
  echo "Installing KEVM from branch ${KEVM_BRANCH} on commit ${KEVM_COMMIT}..."
  echo "===========START============="
  mkdir -p "$INSTALL_DIR" && cd "$INSTALL_DIR"
  git clone --branch ${KEVM_BRANCH} ${KEVM_REPO} evm-semantics
  cd evm-semantics
  git checkout ${KEVM_COMMIT}
  git submodule update --init --recursive
  if ! grep -Fxq "export KEVM_DIR=\"${INSTALL_DIR}/evm-semantics\"" "$HOME/.bashrc"; then
    echo "export KEVM_DIR=\"${INSTALL_DIR}/evm-semantics\"" >> "$HOME/.bashrc"
    source "$HOME/.bashrc"
  fi
  echo "============END=============="
fi

# Build the kreth manually — assumes Cargo.toml and src are present in the cwd
echo "Building kreth..."
echo "===========START============="
cd $CURRENT_DIR
SCRIPT_DIR=$(dirname $(realpath $BASH_SOURCE))
cd $SCRIPT_DIR
## To ensure a clean build, we will remove Cargo.lock, target directory, and ~/.cargo/git/
confirm=$(ask_confirm ">>> To ensure a clean build, is it okay to remove (i) 'Cargo.lock', (ii) 'target' directory, and (iii) '~/.cargo/git/'? [y/N]: ")
if [[ "$confirm" =~ ^[Yy]$ ]]; then
  echo "Cleaning the target directory..."
  rm -rf ~/.cargo/git/
  rm -f $SCRIPT_DIR/Cargo.lock
  cargo clean
  echo "Target directory cleaned."
else
  echo "Keeping all existing files. If you encounter issues, please consider cleaning the target directory manually."
  echo "If you want to ensure a clean build, please remove 'Cargo.lock', 'target' directory, and '~/.cargo/git/'."
fi

## Cargo build with specific flags
echo "Building with EXTRA_CPPFLAGS=-DEVM_ONLY and RETH_CPPFLAGS=-DRETH_BUILD..."
EXTRA_CPPFLAGS=-DEVM_ONLY RETH_CPPFLAGS=-DRETH_BUILD KEVM_DIR=$INSTALL_DIR/evm-semantics cargo build --release
echo "Build completed successfully."
## Move the built binary
if [ -f "./block_processing_kreth" ]; then
  echo "⚠️ './block_processing_kreth' already exists."
  confirm=$(ask_confirm ">>> Do you want to overwrite it? [y/N]: ")
  if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
    echo "Keeping the existing binary './block_processing_kreth'."
    echo "If you want to update it, please manually copy the newly compiled binary from './target/release/block_processing_kreth'."
    exit 0
  fi
  echo "Overwriting './block_processing_kreth' with the newly compiled binary."
fi
cp ./target/release/block_processing_kreth ./block_processing_kreth
echo "============END=============="

# Move shared libraries (if present)
echo "Looking for shared libraries to move..."
echo "===========START============="
sudo find "$SCRIPT_DIR/target/release" -name "libulmkllvm.so" -exec cp {} /usr/lib \;
sudo find "$SCRIPT_DIR/target/release" -name "libkevm.so" -exec cp {} /usr/lib \;
echo "============END=============="

echo "============================="
echo "🎉 Setup complete. Binary at ${SCRIPT_DIR}/block_processing_kreth"
