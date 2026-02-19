#!/bin/bash

# OAuth Server Test Script

BASE_URL="http://localhost:8080"

echo "=== OAuth Server Test Script ==="
echo ""

# 1. Test Health Check
echo "1. Testing health check..."
curl -s "$BASE_URL/health" | jq .
echo ""

# 2. Test OIDC Discovery
echo "2. Testing OIDC Discovery..."
curl -s "$BASE_URL/.well-known/openid-configuration" | jq .
echo ""

# 3. Test Authorization Code Flow
echo "3. Testing Authorization Code Flow..."
echo "Step 1: Get authorization code (auto-approved for demo)"
CODE_RESPONSE=$(curl -s -L "$BASE_URL/oauth/authorize?client_id=YOUR_CLIENT_ID&response_type=code&redirect_uri=http://localhost:3000/callback&scope=openid%20profile%20email&state=random_state")
echo "Authorization response: $CODE_RESPONSE"
echo ""

# Extract code from redirect (this is simplified, in real scenario you'd parse the redirect)
# CODE=$(echo $CODE_RESPONSE | grep -oP 'code=\K[^&]+')

# For demo, use a placeholder
CODE="demo_code"

echo "Step 2: Exchange code for token"
curl -s -X POST "$BASE_URL/api/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "code=$CODE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "redirect_uri=http://localhost:3000/callback" | jq .
echo ""

# 4. Test Client Credentials Flow
echo "4. Testing Client Credentials Flow..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=read write")
echo "$TOKEN_RESPONSE" | jq .
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo ""

# 5. Test Token Introspection
if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
  echo "5. Testing Token Introspection..."
  curl -s -X POST "$BASE_URL/api/oauth/introspect" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "token=$ACCESS_TOKEN" | jq .
  echo ""

  # 6. Test UserInfo Endpoint
  echo "6. Testing UserInfo Endpoint..."
  curl -s "$BASE_URL/api/userinfo" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
  echo ""
fi

# 7. Test Password Flow
echo "7. Testing Password Flow..."
curl -s -X POST "$BASE_URL/api/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=admin" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=openid profile email" | jq .
echo ""

# 8. Test Dynamic Client Registration
echo "8. Testing Dynamic Client Registration..."
curl -s -X POST "$BASE_URL/api/oauth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Test Application",
    "redirect_uris": ["http://localhost:3000/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"],
    "scope": "openid profile email"
  }' | jq .
echo ""

echo "=== Test Complete ==="
