#!/bin/bash

# Simple test script to verify Reality key generation endpoint

echo "Testing Reality Key Generation API..."
echo "======================================"

# Test the reality key generation endpoint
echo "Generating X25519 key pair..."
curl -s -X GET "http://localhost:8080/api/v1/protocols/reality/keys" \
  -H "Accept: application/json" | jq '.'

echo ""
echo "Test completed!"
