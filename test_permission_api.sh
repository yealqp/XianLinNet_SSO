#!/bin/bash

# OAuth Server Permission API Test Script
# This script tests the permission management APIs

# Configuration
BASE_URL="http://localhost:8080"
CLIENT_ID="your-client-id"
CLIENT_SECRET="your-client-secret"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$data" ]; then
        curl -s -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json"
    else
        curl -s -X $method "$BASE_URL$endpoint" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

echo "========================================="
echo "OAuth Server Permission API Test"
echo "========================================="
echo ""

# Step 1: Get admin token
print_info "Step 1: Getting admin token..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/oauth/token" \
    -d "grant_type=password" \
    -d "username=admin" \
    -d "password=admin" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" \
    -d "scope=openid profile email")

TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
    print_success "Got admin token"
else
    print_error "Failed to get admin token"
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

echo ""

# Step 2: List all roles
print_info "Step 2: Listing all roles..."
ROLES=$(api_call GET "/api/roles")
echo "$ROLES" | jq '.'
print_success "Listed roles"
echo ""

# Step 3: Get specific role
print_info "Step 3: Getting admin role details..."
ROLE=$(api_call GET "/api/roles/admin/admin")
echo "$ROLE" | jq '.'
print_success "Got role details"
echo ""

# Step 4: List all permissions
print_info "Step 4: Listing all permissions..."
PERMS=$(api_call GET "/api/permissions")
echo "$PERMS" | jq '.'
print_success "Listed permissions"
echo ""

# Step 5: Get role permissions
print_info "Step 5: Getting permissions for admin role..."
ROLE_PERMS=$(api_call GET "/api/roles/admin/admin/permissions")
echo "$ROLE_PERMS" | jq '.'
print_success "Got role permissions"
echo ""

# Step 6: Get user roles
print_info "Step 6: Getting roles for admin user..."
USER_ROLES=$(api_call GET "/api/users/built-in/admin/roles")
echo "$USER_ROLES" | jq '.'
print_success "Got user roles"
echo ""

# Step 7: Get user effective permissions
print_info "Step 7: Getting effective permissions for admin user..."
USER_PERMS=$(api_call GET "/api/users/built-in/admin/permissions")
echo "$USER_PERMS" | jq '.'
print_success "Got user permissions"
echo ""

# Step 8: Create custom role
print_info "Step 8: Creating custom role..."
CREATE_ROLE=$(api_call POST "/api/roles" '{
    "owner": "admin",
    "name": "test-role",
    "displayName": "Test Role",
    "description": "A test role for API testing",
    "type": "custom",
    "organization": "built-in",
    "isEnabled": true
}')
echo "$CREATE_ROLE" | jq '.'

if echo "$CREATE_ROLE" | jq -e '.status == "ok"' > /dev/null; then
    print_success "Created custom role"
else
    print_error "Failed to create custom role (may already exist)"
fi
echo ""

# Step 9: Create custom permission
print_info "Step 9: Creating custom permission..."
CREATE_PERM=$(api_call POST "/api/permissions" '{
    "owner": "admin",
    "name": "test-permission",
    "displayName": "Test Permission",
    "description": "A test permission",
    "resource": "test",
    "action": "read",
    "effect": "allow",
    "isEnabled": true
}')
echo "$CREATE_PERM" | jq '.'

if echo "$CREATE_PERM" | jq -e '.status == "ok"' > /dev/null; then
    print_success "Created custom permission"
else
    print_error "Failed to create custom permission (may already exist)"
fi
echo ""

# Step 10: Assign permission to role
print_info "Step 10: Assigning permission to role..."
ASSIGN_PERM=$(api_call POST "/api/roles/admin/test-role/permissions" '{
    "permOwner": "admin",
    "permName": "test-permission"
}')
echo "$ASSIGN_PERM" | jq '.'

if echo "$ASSIGN_PERM" | jq -e '.status == "ok"' > /dev/null; then
    print_success "Assigned permission to role"
else
    print_error "Failed to assign permission (may already exist)"
fi
echo ""

# Step 11: Verify role permissions
print_info "Step 11: Verifying role permissions..."
VERIFY_PERMS=$(api_call GET "/api/roles/admin/test-role/permissions")
echo "$VERIFY_PERMS" | jq '.'
print_success "Verified role permissions"
echo ""

# Step 12: Clean up - Remove permission from role
print_info "Step 12: Cleaning up - Removing permission from role..."
REMOVE_PERM=$(api_call DELETE "/api/roles/admin/test-role/permissions/admin/test-permission")
echo "$REMOVE_PERM" | jq '.'
print_success "Removed permission from role"
echo ""

# Step 13: Clean up - Delete custom permission
print_info "Step 13: Cleaning up - Deleting custom permission..."
DELETE_PERM=$(api_call DELETE "/api/permissions/admin/test-permission")
echo "$DELETE_PERM" | jq '.'
print_success "Deleted custom permission"
echo ""

# Step 14: Clean up - Delete custom role
print_info "Step 14: Cleaning up - Deleting custom role..."
DELETE_ROLE=$(api_call DELETE "/api/roles/admin/test-role")
echo "$DELETE_ROLE" | jq '.'
print_success "Deleted custom role"
echo ""

echo "========================================="
echo "All tests completed!"
echo "========================================="
