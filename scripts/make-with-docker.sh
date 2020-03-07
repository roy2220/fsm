#!/usr/bin/env bash

set -euxo pipefail

tar -C scripts/builder -czh . | docker build -t fsm-builder:latest --cache-from=fsm-builder:latest -

tee<<EOF | docker run -i --rm -v"${PWD}:/fsm" fsm-builder:latest /usr/bin/env bash -euxo pipefail /dev/stdin "${@}"
cd /fsm
CGO_ENABLED=0 GOCACHE=\${PWD}/build/cache/go make "\${@}"
EOF
