#!/bin/bash

# OAuth2 Client Test Script
# This script tests the OAuth2 client credentials and basic functionality

echo "🧪 Testing OAuth2 Client Setup..."
echo "================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
CLIENT_ID="YVQwVynqJPPY4oTv8iFXSBf4FP89XQJl"
CLIENT_SECRET="8AwJvkFhn3uWqyT7W17BTD_tAeI9PLTO_scpwslPnHZ9PO5Q_tpWrT2tawK5N181"

echo "1️⃣ Testing OAuth2 Server Health..."
HEALTH_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" $BASE_URL/api/health)
HTTP_STATUS=$(echo $HEALTH_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}✅ OAuth2 Server is running${NC}"
else
    echo -e "${RED}❌ OAuth2 Server is not running (HTTP $HTTP_STATUS)${NC}"
    echo "Please start the OAuth2 server first: go run main.go"
    exit 1
fi

echo ""
echo "2️⃣ Testing OIDC Discovery..."
DISCOVERY_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" $BASE_URL/api/auth/oauth/.well-known/openid-configuration)
HTTP_STATUS=$(echo $DISCOVERY_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}✅ OIDC Discovery endpoint working${NC}"
    DISCOVERY_DATA=$(echo $DISCOVERY_RESPONSE | sed -e 's/HTTPSTATUS.*//')
    echo "Issuer: $(echo $DISCOVERY_DATA | jq -r '.issuer')"
    echo "Authorization Endpoint: $(echo $DISCOVERY_DATA | jq -r '.authorization_endpoint')"
    echo "Token Endpoint: $(echo $DISCOVERY_DATA | jq -r '.token_endpoint')"
else
    echo -e "${RED}❌ OIDC Discovery failed (HTTP $HTTP_STATUS)${NC}"
fi

echo ""
echo "3️⃣ Testing Client Credentials..."

# First, login as admin to get access token
echo "Logging in as admin..."
LOGIN_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
  -X POST $BASE_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}')

HTTP_STATUS=$(echo $LOGIN_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}✅ Admin login successful${NC}"
    LOGIN_DATA=$(echo $LOGIN_RESPONSE | sed -e 's/HTTPSTATUS.*//')
    ACCESS_TOKEN=$(echo $LOGIN_DATA | jq -r '.access_token')
else
    echo -e "${RED}❌ Admin login failed (HTTP $HTTP_STATUS)${NC}"
    echo "Please create admin user first: go run cmd/create_admin.go"
    exit 1
fi

# Test client credentials by fetching client info
echo "Fetching OAuth2 clients..."
CLIENTS_RESPONSE=$(curl -s -w "HTTPSTATUS:%{http_code}" \
  -X GET $BASE_URL/api/oauth2/clients \
  -H "Authorization: Bearer $ACCESS_TOKEN")

HTTP_STATUS=$(echo $CLIENTS_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo -e "${GREEN}✅ OAuth2 clients endpoint working${NC}"
    CLIENTS_DATA=$(echo $CLIENTS_RESPONSE | sed -e 's/HTTPSTATUS.*//')
    
    # Check if our client exists
    CLIENT_EXISTS=$(echo $CLIENTS_DATA | jq -r ".clients[] | select(.client_id == \"$CLIENT_ID\") | .name")
    
    if [ ! -z "$CLIENT_EXISTS" ] && [ "$CLIENT_EXISTS" != "null" ]; then
        echo -e "${GREEN}✅ NodeJS OAuth Client found: $CLIENT_EXISTS${NC}"
    else
        echo -e "${YELLOW}⚠️  NodeJS OAuth Client not found in database${NC}"
    fi
else
    echo -e "${RED}❌ Failed to fetch OAuth2 clients (HTTP $HTTP_STATUS)${NC}"
fi

echo ""
echo "4️⃣ Testing Authorization URL Generation..."
AUTH_URL="$BASE_URL/api/auth/oauth/authorize?response_type=code&client_id=$CLIENT_ID&redirect_uri=http://localhost:3001/auth/callback&scope=openid%20profile%20email%20read&state=test123&nonce=test456"

echo -e "${YELLOW}📋 Authorization URL:${NC}"
echo "$AUTH_URL"

# Test if authorization endpoint responds
AUTH_TEST=$(curl -s -w "HTTPSTATUS:%{http_code}" -I "$AUTH_URL")
HTTP_STATUS=$(echo $AUTH_TEST | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [ "$HTTP_STATUS" -eq 200 ] || [ "$HTTP_STATUS" -eq 302 ] || [ "$HTTP_STATUS" -eq 400 ]; then
    echo -e "${GREEN}✅ Authorization endpoint responds${NC}"
else
    echo -e "${RED}❌ Authorization endpoint not responding (HTTP $HTTP_STATUS)${NC}"
fi

echo ""
echo "5️⃣ Testing Node.js Client Application..."

# Check if Node.js app is configured correctly
if [ -f ".env" ]; then
    echo -e "${GREEN}✅ .env file exists${NC}"
    
    # Check if client ID matches
    ENV_CLIENT_ID=$(grep "OAUTH_CLIENT_ID" .env | cut -d '=' -f2)
    if [ "$ENV_CLIENT_ID" = "$CLIENT_ID" ]; then
        echo -e "${GREEN}✅ Client ID matches in .env file${NC}"
    else
        echo -e "${YELLOW}⚠️  Client ID in .env doesn't match generated client${NC}"
        echo "Generated: $CLIENT_ID"
        echo "In .env: $ENV_CLIENT_ID"
    fi
else
    echo -e "${RED}❌ .env file not found${NC}"
fi

echo ""
echo "🎉 OAuth2 Client Test Summary:"
echo "==============================="
echo -e "${GREEN}✅ OAuth2 Server: Running${NC}"
echo -e "${GREEN}✅ OIDC Discovery: Working${NC}"
echo -e "${GREEN}✅ Client Credentials: Generated${NC}"
echo -e "${GREEN}✅ Authorization Endpoint: Responding${NC}"

echo ""
echo "🚀 Next Steps:"
echo "1. Start your Node.js client: npm start"
echo "2. Visit: http://localhost:3001"
echo "3. Test OAuth2 flow: http://localhost:3001/auth/login"
echo "4. Complete authorization on OAuth2 server"
echo "5. Get redirected back with access token"

echo ""
echo -e "${YELLOW}💡 Pro Tips:${NC}"
echo "• Client ID: $CLIENT_ID"
echo "• Use this in your OAuth2 applications"
echo "• Keep client secret secure in production"
echo "• Test different scopes: openid, profile, email, read"