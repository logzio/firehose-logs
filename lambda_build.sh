#!/bin/bash

# Ensure the script exits on first error
set -e

# Define the function directories
FUNCTION_DIRS=("cfn-lambda" "eventbridge-lambda")

# Loop through each function directory to build and zip
for dir in "${FUNCTION_DIRS[@]}"; do
    echo "Building and zipping $dir..."

    # Copy the common directory to the current function directory
    cp -R common "$dir/"

    # Move into the function directory
    cd "$dir"

    # Build the Go binary named bootstrap
    GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

    # Set the file permissions as needed
    chmod 644 $(find . -type f)
    chmod 755 $(find . -type d) bootstrap

    # Zip the binary into a package named after the directory
    zip "../${dir}.zip" bootstrap

    # Clean up: Remove the binary and copied common directory
    rm bootstrap
    rm -rf common

    # Move back to the project root directory
    cd ..

    echo "$dir build and packaging complete."
done

echo "All Lambda functions have been built and packaged."
