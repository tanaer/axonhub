#!/bin/bash
set -e
SRC="/root/axonhub/frontend/dist"
DST="/root/axonhub/internal/server/static/dist"

# Copy all files from frontend/dist to server static/dist
for item in "$SRC"/*; do
    cp -r "$item" "$DST/"
done

echo "Sync completed"
ls -la "$DST"
