#!/bin/bash

set -e

API_URL="${API_URL:-http://localhost:8080}"
EMAIL="test@example.com"
USERNAME="testuser"
PASSWORD="testpassword123"

echo "================================"
echo "Testing Habit Tracker API"
echo "================================"
echo ""

# Health Check
echo "1. Health Check..."
curl -s "$API_URL/health" | jq '.' || echo "OK"
echo ""
echo ""

# Register
echo "2. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"$USERNAME\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"Test\",
    \"timezone\": \"Europe/Moscow\"
  }")

echo "$REGISTER_RESPONSE" | jq '.'
echo ""
echo ""

# Login
echo "3. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email_or_username\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "$LOGIN_RESPONSE" | jq '.'

# Extract access token
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "Failed to get access token"
  exit 1
fi

echo ""
echo "Access Token: $ACCESS_TOKEN"
echo ""
echo ""

# Get Profile
echo "4. Getting user profile..."
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/api/v1/users/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PROFILE_RESPONSE" | jq '.'
echo ""
echo ""

# Logout
echo "5. Logging out..."
LOGOUT_RESPONSE=$(curl -s -X POST "$API_URL/api/v1/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$LOGOUT_RESPONSE" | jq '.'
echo ""
echo ""

echo "================================"
echo "All tests completed successfully!"
echo "================================"
