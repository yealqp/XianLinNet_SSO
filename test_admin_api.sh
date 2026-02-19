#!/bin/bash

# OAuth Server Admin API Test Script

BASE_URL="http://localhost:8080"

echo "=== OAuth Server Admin API Test Script ==="
echo ""

# Step 1: Get admin token
echo "1. Getting admin token..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "username=admin" \
  -d "password=admin" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "scope=openid profile email")

ADMIN_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" == "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  echo "Failed to get admin token. Please update CLIENT_ID and CLIENT_SECRET"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "Admin token obtained!"
echo ""

# Step 2: Get system stats
echo "2. Getting system statistics..."
curl -s "$BASE_URL/api/admin/stats" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 3: Get system info
echo "3. Getting system information..."
curl -s "$BASE_URL/api/admin/system" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 4: List users
echo "4. Listing users..."
curl -s "$BASE_URL/api/admin/users?owner=built-in" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 5: Create a test user
echo "5. Creating test user..."
curl -s -X POST "$BASE_URL/api/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "built-in",
    "name": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "displayName": "Test User",
    "isAdmin": false
  }' | jq .
echo ""

# Step 6: Get the test user
echo "6. Getting test user..."
curl -s "$BASE_URL/api/admin/users/built-in/testuser" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 7: Update the test user
echo "7. Updating test user..."
curl -s -X PUT "$BASE_URL/api/admin/users/built-in/testuser" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "built-in",
    "name": "testuser",
    "email": "updated@example.com",
    "displayName": "Updated Test User",
    "isAdmin": false
  }' | jq .
echo ""

# Step 8: List applications
echo "8. Listing applications..."
curl -s "$BASE_URL/api/admin/applications?owner=admin" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 9: Create a test application
echo "9. Creating test application..."
curl -s -X POST "$BASE_URL/api/admin/applications" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "owner": "admin",
    "name": "test-app",
    "displayName": "Test Application",
    "organization": "built-in",
    "redirectUris": ["http://localhost:3000/callback"],
    "grantTypes": ["authorization_code", "refresh_token"],
    "enablePassword": true,
    "enableSignUp": true,
    "expireInHours": 168
  }' | jq .
echo ""

# Step 10: List tokens
echo "10. Listing tokens..."
curl -s "$BASE_URL/api/admin/tokens?owner=admin" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 11: Clear cache (if Redis is configured)
echo "11. Clearing cache..."
curl -s -X POST "$BASE_URL/api/admin/cache/clear" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 12: Delete test user
echo "12. Deleting test user..."
curl -s -X DELETE "$BASE_URL/api/admin/users/built-in/testuser" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

# Step 13: Delete test application
echo "13. Deleting test application..."
curl -s -X DELETE "$BASE_URL/api/admin/applications/admin/test-app" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo ""

echo "=== Admin API Test Complete ==="
