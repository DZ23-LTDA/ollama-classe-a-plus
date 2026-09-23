#!/bin/sh

set -eu

. $(dirname $0)/env.sh

# Set PUSH to a non-empty string to trigger push instead of load
PUSH=${PUSH:-""}

if [ -z "${PUSH}" ] ; then
    echo "Building ${FINAL_IMAGE_REPO}:$VERSION locally.  set PUSH=1 to push"
    LOAD_OR_PUSH="--load"
else
    if [ "${FINAL_IMAGE_REPO}" = "local/ollama-classe-a-plus" ]; then
        echo "ERROR: set FINAL_IMAGE_REPO to an operator-owned registry before pushing" >&2
        exit 1
    fi
    echo "Will be pushing ${FINAL_IMAGE_REPO}:$VERSION"
    LOAD_OR_PUSH="--push"
fi

docker buildx build \
    ${LOAD_OR_PUSH} \
    --platform=${PLATFORM} \
    ${OLLAMA_COMMON_BUILD_ARGS} \
    -f Dockerfile \
    -t ${FINAL_IMAGE_REPO}:$VERSION \
    .

if echo $PLATFORM | grep "amd64" > /dev/null; then
    docker buildx build \
        ${LOAD_OR_PUSH} \
        --platform=linux/amd64 \
        ${OLLAMA_COMMON_BUILD_ARGS} \
        --build-arg FLAVOR=rocm \
        -f Dockerfile \
        -t ${FINAL_IMAGE_REPO}:$VERSION-rocm \
        .
fi
