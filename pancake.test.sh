#!/bin/bash

check_status() {
    if [ $? -ne 0 ]; then
        echo "$1 failed!"
        exit 1
    fi
}

run_test() {
    echo "🧪 Pancake Test Suite: $2..."
    echo "Running command: $1" # Add a comment to show which command is running
    eval $1
    check_status "$2"
}

echo "🧪 Pancake Test Suite: Starting..."

run_test "go build" "Building"
run_test "go install" "Installing"
run_test "pancake version" "Checking version"

if [ -e ~/pancake.yml ]; then
    echo "pancake.yml exists. Deleting the file..."
    rm ~/pancake.yml
    check_status "Delete pancake.yml"
    echo "pancake.yml deleted successfully."
else
    echo "pancake.yml does not exist. No action needed."
fi

run_test "pancake edit-config" "Opening config file for the first time"
run_test "pancake edit-config" "Opening config file for the second time"
run_test "pancake project list" "Listing projects"

echo "🧪 Pancake Test Suite: End."


export PATH="/Users/unicorn/go/bin:$PATH"
