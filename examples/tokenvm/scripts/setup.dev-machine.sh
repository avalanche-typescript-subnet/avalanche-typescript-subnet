#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Set constants
export HYPERSDK_VERSION="0.0.14"
echo HYPERSDK_VERSION: ${HYPERSDK_VERSION}

# Download token-cli
wget "https://github.com/ava-labs/hypersdk/releases/download/v${HYPERSDK_VERSION}/tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz"
mkdir tmp-hypersdk
tar -xvf tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz -C tmp-hypersdk
rm -rf tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz
mv tmp-hypersdk/token-cli /tmp/token-cli
rm -rf tmp-hypersdk

# Download prometheus
rm -f /tmp/prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.43.0/prometheus-2.43.0.linux-amd64.tar.gz
tar -xvf prometheus-2.43.0.linux-amd64.tar.gz
rm prometheus-2.43.0.linux-amd64.tar.gz
mv prometheus-2.43.0.linux-amd64/prometheus /tmp/prometheus
rm -rf prometheus-2.43.0.linux-amd64

# Import demo.pk and avalanche-ops spec

# Start prometheus server