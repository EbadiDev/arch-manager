#!/bin/bash

# Test script for Reality key generation API with proper authentication

echo "Testing Reality Key Generation API..."
echo "======================================"

# Default admin password (can be overridden by environment variable)
ADMIN_PASSWORD="${ADMIN_PASSWORD:-password}"

echo "Using admin password: $ADMIN_PASSWORD"
echo ""

# Test the reality key generation endpoint with proper authentication
echo "Generating X25519 key pair..."
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST "http://localhost:8080/v1/generate-reality-keys" \
  -H "Authorization: Bearer $ADMIN_PASSWORD" \
  -H "Content-Type: application/json")

# Extract HTTP code and response body
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE:/d')

echo "HTTP Status: $http_code"
echo "Response:"

if [ "$http_code" = "200" ]; then
    echo "$body" | jq '.'
else
    echo "$body"
fi

echo ""
echo "Test completed!"

# Test multiple key generations to verify uniqueness
if [ "$http_code" = "200" ]; then
    echo ""
    echo "Testing key uniqueness (generating 3 more pairs)..."
    echo "=================================================="
    
    for i in {1..3}; do
        echo "Key pair $i:"
        curl -s -X POST "http://localhost:8080/v1/generate-reality-keys" \
          -H "Authorization: Bearer $ADMIN_PASSWORD" \
          -H "Content-Type: application/json" | jq -r '"Private: " + .private_key + "\nPublic:  " + .public_key'
        echo ""
    done
fi
