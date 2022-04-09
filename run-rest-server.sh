#!/bin/bash

# remove old build
rm -rf pda-processor

if [ ! $# -eq 1 ]; then
  # build project
  go build
else
  if [[ $1 =~ ^[0-9]{4}[:.,-]?$ ]]; then
    # build with using custom port
    go build -ldflags "-X main.port=$1"
  else
    echo "invalid port number: $1"
    exit 1
  fi
fi

# run project
./pda-processor
