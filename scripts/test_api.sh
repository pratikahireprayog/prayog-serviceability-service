#!/bin/bash

# Base URL
BASE_URL="http://localhost:8080"

echo "Testing API endpoints..."

# Test ping endpoint
echo -n "Testing /ping: "
PING_RESPONSE=$(curl -s "${BASE_URL}/ping")
if [ "$PING_RESPONSE" == "pong" ]; then
  echo "✅ Success"
else
  echo "❌ Failed (Response: $PING_RESPONSE)"
fi

# Test serviceability check endpoint
POSTAL_CODE="12345"
echo -n "Testing /api/v1/serviceability/check/$POSTAL_CODE: "
CHECK_RESPONSE=$(curl -s "${BASE_URL}/api/v1/serviceability/check/$POSTAL_CODE")
if [[ "$CHECK_RESPONSE" == *"\"postal_code\": \"$POSTAL_CODE\""* && "$CHECK_RESPONSE" == *"\"is_serviceable\": true"* ]]; then
  echo "✅ Success"
else
  echo "❌ Failed (Response: $CHECK_RESPONSE)"
fi

echo "All tests completed!" 