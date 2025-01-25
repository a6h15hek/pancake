#!/bin/bash

check_status() {
    if [ $? -ne 0 ]; then
        echo "$1 failed!"
        exit 1
    fi
}

run_test() {
    echo "-----------------------------------------------------------------------"
    echo "🧪 Pancake Test Suite: $2..."
    echo "Running command: $1" # Add a comment to show which command is running
    eval $1
    check_status "$2"
    echo "-----------------------------------------------------------------------"
}

echo "🧪 Pancake Test Suite: Starting..."

run_test "go build" "Building"
run_test "go install" "Installing"
run_test "pancake version" "Checking version"

# Test to check default file create works.
if [ -e ~/pancake.yml ]; then
    echo "pancake.yml exists. Deleting the file..."
    rm ~/pancake.yml
    check_status "Delete pancake.yml"
    echo "pancake.yml deleted successfully."
else
    echo "pancake.yml does not exist. No action needed."
fi

run_test "pancake editconfig" "Opening config file for the first time"
run_test "pancake editconfig" "Opening config file for the second time"
run_test "pancake project list" "Listing projects"

run_test "pancake project sync spring-helloworld" "Sync 1 projects"
run_test "pancake project sync" "Sync all projects"

run_test "pancake project open spring-helloworld" "Open 1 projects"
run_test "pancake project open" "Open all projects"

run_test "pancake project build spring-helloworld" "build 1 projects"
run_test "pancake project build" "build all projects"

#run_test "pancake project start spring-helloworld" "Start 1 projects"
#run_test "pancake project start" "Start all projects"

echo "🧪 Pancake Test Suite: End."


export PATH="/Users/unicorn/go/bin:$PATH"
