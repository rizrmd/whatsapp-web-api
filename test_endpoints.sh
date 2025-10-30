#!/bin/bash

# Test script for WhatsApp Web API endpoints
# Usage: ./test_endpoints.sh

BASE_URL="http://localhost:8080"

echo "🧪 Testing WhatsApp Web API Endpoints"
echo "====================================="

# Test health endpoint
echo ""
echo "1. Testing Health Endpoint:"
curl -s "$BASE_URL/health" | jq '.' 2>/dev/null || curl -s "$BASE_URL/health"

# Test devices endpoint
echo ""
echo "2. Testing Devices Endpoint:"
curl -s "$BASE_URL/devices" | jq '.' 2>/dev/null || curl -s "$BASE_URL/devices"

# Test disconnect endpoint
echo ""
echo "3. Testing Disconnect Endpoint:"
curl -s -X POST "$BASE_URL/disconnect" | jq '.' 2>/dev/null || curl -s -X POST "$BASE_URL/disconnect"

echo ""
echo "====================================="
echo "✅ Endpoint testing complete!"
echo ""
echo "💡 To test QR pairing, run:"
echo "   curl $BASE_URL/pair"
echo ""
echo "💡 To send a message after pairing:"
echo "   curl -X POST $BASE_URL/send -H 'Content-Type: application/json' -d '{\"number\":\"1234567890\",\"message\":\"Hello from API!\"}'"