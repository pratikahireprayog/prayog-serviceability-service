#!/bin/bash

# Test script to verify the ISO 3166 validation fix
# Run this after restarting your server

SERVER_URL="http://localhost:8080/serviceability/v1"

echo "🧪 Testing validation fix for ISO 3166 validation functions"
echo "========================================================="

# Test 1: Original failing case - should now work (region-type still uses snake_case)
echo ""
echo "Test 1: Creating region-type with code 'state' (original failing case)"
echo "----------------------------------------------------------------------"
response1=$(curl -s -X POST "$SERVER_URL/region-types" \
  -H "Content-Type: application/json" \
  -d '{"code": "state", "name": "State", "description": "State level administrative division"}')

if [[ $response1 == *"Undefined validation function"* ]]; then
  echo "❌ FAIL: Still getting 'Undefined validation function' error"
  echo "Response: $response1"
else
  echo "✅ PASS: No 'Undefined validation function' error"
  echo "Response: $response1"
fi

# Test 2: Snake case format for region-type - should work
echo ""
echo "Test 2: Creating region-type with code 'state_province' (snake_case format)"
echo "--------------------------------------------------------------------------"
response2=$(curl -s -X POST "$SERVER_URL/region-types" \
  -H "Content-Type: application/json" \
  -d '{"code": "state_province", "name": "State Province", "description": "State province level"}')

if [[ $response2 == *"Undefined validation function"* ]]; then
  echo "❌ FAIL: Still getting 'Undefined validation function' error"
  echo "Response: $response2"
else
  echo "✅ PASS: No 'Undefined validation function' error"
  echo "Response: $response2"
fi

# Test 3: ISO 3166-2 format for regions - should work
echo ""
echo "Test 3: Creating region with code 'US-CA' (ISO 3166-2 format)"
echo "-------------------------------------------------------------"
response3=$(curl -s -X POST "$SERVER_URL/regions" \
  -H "Content-Type: application/json" \
  -d '{"code": "US-CA", "name": "California", "country_code": "US", "region_type_code": "state"}')

if [[ $response3 == *"Undefined validation function"* ]]; then
  echo "❌ FAIL: Still getting 'Undefined validation function' error"
  echo "Response: $response3"
elif [[ $response3 == *"error"* ]] && [[ $response3 == *"region_code_iso"* ]]; then
  echo "✅ PASS: Getting proper validation error (not undefined function error)"
  echo "Response: $response3"
else
  echo "✅ PASS: No validation errors"
  echo "Response: $response3"
fi

# Test 4: Invalid format for regions - should fail with proper validation message
echo ""
echo "Test 4: Creating region with code 'state' (should fail ISO 3166-2 validation)"
echo "-----------------------------------------------------------------------------"
response4=$(curl -s -X POST "$SERVER_URL/regions" \
  -H "Content-Type: application/json" \
  -d '{"code": "state", "name": "State", "country_code": "US", "region_type_code": "state"}')

if [[ $response4 == *"Undefined validation function"* ]]; then
  echo "❌ FAIL: Still getting 'Undefined validation function' error instead of proper validation"
  echo "Response: $response4"
elif [[ $response4 == *"region_code_iso"* ]] || [[ $response4 == *"ISO 3166-2"* ]]; then
  echo "✅ PASS: Getting proper validation error (not undefined function error)"
  echo "Response: $response4"
else
  echo "⚠️  UNKNOWN: Unexpected response"
  echo "Response: $response4"
fi

echo ""
echo "========================================================="
echo "📋 **Validation Standards Summary**"
echo ""
echo "🏛️  **Region Codes**: ISO 3166-2 format (CC-XXX)"
echo "   ✅ Valid: US-CA, IN-MH, GB-ENG, AU-NSW"
echo "   ❌ Invalid: state, US_CA, USCA, us-ca"
echo ""
echo "🌍 **Country Codes**: ISO 3166-1 A-2 format (CC)"
echo "   ✅ Valid: US, IN, GB, AU"
echo "   ❌ Invalid: us, USA, 1A"
echo ""
echo "🏪 **Region Types**: snake_case format"
echo "   ✅ Valid: state, state_province, county"
echo "   ❌ Invalid: STATE, State, state-province"
echo ""
echo "🏢 **Districts/Cities/Areas**: snake_case format"
echo "   ✅ Valid: downtown, metro_area, city_center"
echo "   ❌ Invalid: DOWNTOWN, Metro_Area, city-center"
echo ""
echo "🔄 Remember to restart your server before running this test!"
echo "   Command: make restart-server  (or your restart command)"
echo "=========================================================" 