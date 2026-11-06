#!/bin/sh

# Exit immediately if a command exits with a non-zero status
set -e

echo "Initializing Upload Directories..."
mkdir -p uploads/attachments uploads/faces uploads/selfies uploads/files

if [ "$SKIP_SEED" != "true" ]; then
    echo "Running Database Seeding..."
    ./seeder
else
    echo "Skipping Database Seeding..."
fi

echo "Starting API Server..."
exec ./api
