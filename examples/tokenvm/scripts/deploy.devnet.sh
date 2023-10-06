#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Ensure we return back to current directory
pw=$(pwd)
function cleanup() {
  cd $pw
}
trap cleanup EXIT

# Setup cache
BUST_CACHE=${BUST_CACHE:-false}
if ${BUST_CACHE}; then
  rm -rf /tmp/avalanche-ops-cache
fi
mkdir -p /tmp/avalanche-ops-cache

# Create deployment directory (avalanche-ops creates metadata in cwd)
DATE=$(date '+%m_%d_%Y-%H:%M:%S')
DEPLOY_PREFIX=~/avalanche-ops/deploys/${DATE}
mkdir -p ${DEPLOY_PREFIX}
DEPLOY_ARTIFACT_PREFIX=${DEPLOY_PREFIX}/artifacts
mkdir -p ${DEPLOY_ARTIFACT_PREFIX}
echo create deployment folder: ${DEPLOY_PREFIX}
cd ${DEPLOY_PREFIX}

# Set constants
export ARCH_TYPE=$(uname -m)
[ $ARCH_TYPE = x86_64 ] && ARCH_TYPE=amd64
echo ARCH_TYPE: ${ARCH_TYPE}
export OS_TYPE=$(uname | tr '[:upper:]' '[:lower:]')
echo OS_TYPE: ${OS_TYPE}
export AVALANCHEGO_VERSION="1.10.11"
echo AVALANCHEGO_VERSION: ${AVALANCHEGO_VERSION}
export HYPERSDK_VERSION="0.0.14"
echo HYPERSDK_VERSION: ${HYPERSDK_VERSION}
# TODO: set deploy os/arch

# Check valid setup
if [ ${OS_TYPE} != 'darwin' ]; then
  echo 'os is not supported' >&2
  exit 1
fi
if [ ${ARCH_TYPE} != 'arm64' ]; then
  echo 'arch is not supported' >&2
  exit 1
fi
if ! [ -x "$(command -v aws)" ]; then
  echo 'aws-cli is not installed' >&2
  exit 1
fi
if ! [ -x "$(command -v yq)" ]; then
  echo 'yq is not installed' >&2
  exit 1
fi

# Install avalanche-ops
echo 'installing avalanche-ops...'
if [ -f /tmp/avalanche-ops-cache/avalancheup-aws ]; then
  cp /tmp/avalanche-ops-cache/avalancheup-aws ${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws
  echo 'found avalanche-ops in cache'
else
  wget https://github.com/ava-labs/avalanche-ops/releases/download/latest/avalancheup-aws.aarch64-apple-darwin
  mv ./avalancheup-aws.aarch64-apple-darwin ${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws
  chmod +x ${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws
  cp ${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws /tmp/avalanche-ops-cache/avalancheup-aws
fi
${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws --help

# Install token-cli
echo 'installing token-cli...'
if [ -f /tmp/avalanche-ops-cache/token-cli ]; then
  cp /tmp/avalanche-ops-cache/token-cli ${DEPLOY_ARTIFACT_PREFIX}/token-cli
  echo 'found token-cli in cache'
else
  wget "https://github.com/ava-labs/hypersdk/releases/download/v${HYPERSDK_VERSION}/tokenvm_${HYPERSDK_VERSION}_${OS_TYPE}_${ARCH_TYPE}.tar.gz"
  mkdir -p /tmp/token-installs
  tar -xvf tokenvm_${HYPERSDK_VERSION}_${OS_TYPE}_${ARCH_TYPE}.tar.gz -C /tmp/token-installs
  rm -rf tokenvm_${HYPERSDK_VERSION}_${OS_TYPE}_${ARCH_TYPE}.tar.gz
  mv /tmp/token-installs/token-cli ${DEPLOY_ARTIFACT_PREFIX}/token-cli
  rm -rf /tmp/token-installs
  cp ${DEPLOY_ARTIFACT_PREFIX}/token-cli /tmp/avalanche-ops-cache/token-cli
fi

# Download tokenvm
echo 'downloading tokenvm...'
if [ -f /tmp/avalanche-ops-cache/tokenvm ]; then
  cp /tmp/avalanche-ops-cache/tokenvm ${DEPLOY_ARTIFACT_PREFIX}/tokenvm
  echo 'found tokenvm in cache'
else
  wget "https://github.com/ava-labs/hypersdk/releases/download/v${HYPERSDK_VERSION}/tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz"
  mkdir -p /tmp/token-installs
  tar -xvf tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz -C /tmp/token-installs
  rm -rf tokenvm_${HYPERSDK_VERSION}_linux_amd64.tar.gz
  mv /tmp/token-installs/tokenvm ${DEPLOY_ARTIFACT_PREFIX}/tokenvm
  rm -rf /tmp/token-installs
  cp ${DEPLOY_ARTIFACT_PREFIX}/tokenvm /tmp/avalanche-ops-cache/tokenvm
fi

# Setup genesis and configuration files
cat <<EOF > ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-subnet-config.json
{
  "proposerMinBlockDelay": 0,
  "proposerNumHistoricalBlocks": 768
}
EOF
cat ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-subnet-config.json

# TODO: make address configurable via ENV
cat <<EOF > ${DEPLOY_ARTIFACT_PREFIX}/allocations.json
[{"address":"token1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsjzf3yp", "balance":1000000000000000}]
EOF

# TODO: make fee params configurable via ENV
${DEPLOY_ARTIFACT_PREFIX}/token-cli genesis generate ${DEPLOY_ARTIFACT_PREFIX}/allocations.json \
--genesis-file ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-genesis.json \
--max-block-units 18446744073709551615,18446744073709551615,18446744073709551615,18446744073709551615,18446744073709551615 \
--window-target-units 1800000,18446744073709551615,18446744073709551615,18446744073709551615,18446744073709551615 \
--min-block-gap 250
cat ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-genesis.json

cat <<EOF > ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-chain-config.json
{
  "mempoolSize": 10000000,
  "mempoolPayerSize": 10000000,
  "mempoolExemptPayers":["token1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsjzf3yp"],
  "streamingBacklogSize": 10000000,
  "storeTransactions": false,
  "verifySignatures": true,
  "trackedPairs":["*"],
  "logLevel": "info",
}
EOF
cat ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-chain-config.json

# Plan network deploy
if [ ! -f /tmp/avalanche-ops-cache/aws-profile ]; then
  echo 'what is your AWS profile name?'
  read prof_name
  echo ${prof_name} > /tmp/avalanche-ops-cache/aws-profile
fi
AWS_PROFILE_NAME=$(cat "/tmp/avalanche-ops-cache/aws-profile")

# Create spec file
SPEC_FILE=./aops-${DATE}.yml
echo created avalanche-ops spec file: ${SPEC_FILE}

# Create key file dir
KEY_FILES_DIR=keys
mkdir -p ${KEY_FILES_DIR}

# Create dummy metrics file (can't not upload)
# TODO: fix this
cat <<EOF > "${DEPLOY_ARTIFACT_PREFIX}/metrics.yml"
filters:
  - regex: ^*$
EOF
cat ${DEPLOY_ARTIFACT_PREFIX}/metrics.yml

echo 'planning DEVNET deploy...'
# TODO: increase size once dev machine is working
${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws default-spec \
--arch-type amd64 \
--os-type ubuntu20.04 \
--anchor-nodes 2 \
--non-anchor-nodes 1 \
--regions us-west-2 \
--instance-mode=on-demand \
--instance-types='{"us-west-2":["c5.4xlarge"]}' \
--ip-mode=ephemeral \
--metrics-fetch-interval-seconds 0 \
--upload-artifacts-prometheus-metrics-rules-file-path ${DEPLOY_ARTIFACT_PREFIX}/metrics.yml \
--network-name custom \
--avalanchego-release-tag v${AVALANCHEGO_VERSION} \
--create-dev-machine \
--keys-to-generate 5 \
--subnet-config-file ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-subnet-config.json \
--vm-binary-file ${DEPLOY_ARTIFACT_PREFIX}/tokenvm \
--chain-name tokenvm \
--chain-genesis-file ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-genesis.json \
--chain-config-file ${DEPLOY_ARTIFACT_PREFIX}/tokenvm-chain-config.json \
--enable-ssh \
--spec-file-path ${SPEC_FILE} \
--key-files-dir ${KEY_FILES_DIR} \
--profile-name ${AWS_PROFILE_NAME}

# Disable rate limits in config
echo 'updating YAML with new rate limits...'
yq -i '.avalanchego_config.throttler-inbound-validator-alloc-size = 10737418240' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-at-large-alloc-size = 10737418240' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-node-max-processing-msgs = 100000' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-bandwidth-refill-rate = 1073741824' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-bandwidth-max-burst-size = 1073741824' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-cpu-validator-alloc = 100000' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-inbound-disk-validator-alloc = 10737418240000' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-outbound-validator-alloc-size = 10737418240' ${SPEC_FILE}
yq -i '.avalanchego_config.throttler-outbound-at-large-alloc-size = 10737418240' ${SPEC_FILE}
yq -i '.avalanchego_config.consensus-on-accept-gossip-validator-size = 10' ${SPEC_FILE}
yq -i '.avalanchego_config.consensus-on-accept-gossip-non-validator-size = 0' ${SPEC_FILE}
yq -i '.avalanchego_config.consensus-on-accept-gossip-peer-size = 10' ${SPEC_FILE}
yq -i '.avalanchego_config.consensus-accepted-frontier-gossip-peer-size = 10' ${SPEC_FILE}
yq -i '.avalanchego_config.consensus-app-concurrency = 8' ${SPEC_FILE}
yq -i '.avalanchego_config.network-compression-type = "zstd"'  ${SPEC_FILE}

# Deploy DEVNET
echo 'deploying DEVNET...'
${DEPLOY_ARTIFACT_PREFIX}/avalancheup-aws apply \
--spec-file-path ${SPEC_FILE}

# Prepare dev-machine and start prometheus server
# Copy avalanche-ops spec, demo.pk, sign into dev machine, download token-cli, configure token-cli, start prometheus in background

# Print final logs
cat << EOF
to view prometheus metrics, visit the following URL:

TODO: just generate URL

to run spam script on dev machine, run the following command:

TODO

to delete all resources (but keep asg/ssm), run the following command:

/tmp/avalancheup-aws delete \
--delete-cloudwatch-log-group \
--delete-s3-objects \
--delete-ebs-volumes \
--delete-elastic-ips \
--spec-file-path ${SPEC_FILE}

to delete all resources, run the following command:

/tmp/avalancheup-aws delete \
--override-keep-resources-except-asg-ssm \
--delete-cloudwatch-log-group \
--delete-s3-objects \
--delete-ebs-volumes \
--delete-elastic-ips \
--spec-file-path ${SPEC_FILE}
EOF
