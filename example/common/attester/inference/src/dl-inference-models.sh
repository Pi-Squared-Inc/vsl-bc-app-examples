#!/bin/bash

set -e

# Check if gdown is installed, if not, install it
if ! command -v gdown &> /dev/null; then
    echo "gdown not found. Installing with pip..."
    pip install --user gdown
    export PATH="$HOME/.local/bin:$PATH"
fi

echo "Downloading inference models..."
echo "==========================================================="
# Find the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Set MODEL_DIR relative to the script location
MODEL_DIR="$SCRIPT_DIR/models"
mkdir -p "$MODEL_DIR"

# LLM model
LLM_ID="10pvHz51JmtpktQDY0P6OYFq8oAtooPSk"
LLM_OUT="$MODEL_DIR/pi2prunedLaMa.gguf"

# ResNet model
RESNET_ID="1cSbZ4_A8LYoTBHaNGHgj2BY8jGCHwajL"
RESNET_OUT="$MODEL_DIR/pi2resnetmodel.pt"

echo "Downloading LLM model..."
gdown "$LLM_ID" -O "$LLM_OUT"

echo "Downloading ResNet model..."
gdown "$RESNET_ID" -O "$RESNET_OUT"

echo "Download complete. Models saved in $MODEL_DIR"
echo "==========================================================="