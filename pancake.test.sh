#!/bin/bash
echo "🧪 Pancake Test Suite: Starting..."

echo "🧪 Pancake Test Suite: Building and Installing..."
go build
go install

echo "🧪 Pancake Test Suite: Checking version..."
pancake version

# if [ -e ~/pancake.yml ]; then
#     echo "pancake.yml exists. Deleting the file..."
#     rm ~/pancake.yml
#     echo "pancake.yml deleted successfully."
# else
#     echo "pancake.yml does not exist. No action needed."
# fi

# echo "🧪 Pancake Test Suite: Opening files for the first time"
# pancake edit-config

echo "🧪 Pancake Test Suite: Opening files for the second time"
pancake edit-config

echo "🧪 Pancake Test Suite: List projects..."
pancake project list

echo "🧪 Pancake Test Suite: End."
