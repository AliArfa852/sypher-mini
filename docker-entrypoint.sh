#!/bin/sh
set -e
CONFIG_DIR="${HOME}/.sypher-mini"
CONFIG_FILE="${CONFIG_DIR}/config.json"

if [ ! -f "$CONFIG_FILE" ]; then
  echo "Initializing config..."
  sypher onboard
fi

exec sypher "$@"
