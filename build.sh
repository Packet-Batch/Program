#!/bin/bash
THREADS=1

# Check for core numbers.
if [ -n "$1" ]; then
    if [ "$1" -eq 0 ]; then
        THREADS=$(nproc)
    elif [ "$1" -gt 0 ]; then
        THREADS=$1
    fi
fi

echo "Building Packet Batch..."