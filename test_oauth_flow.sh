#!/bin/bash

echo "🧪 Testing OAuth2 Server API"
echo "================================="

BASE_URL="http://localhost:8080"

echo ""
echo "1️⃣ Health Check..."
curl -s $BASE_URL/health | jq .

echo ""
echo "2️⃣ Testing Registration..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testapi",
    "email": "testapi@example.com",
    "password": "password123",
    "full_name": "Test API User"
  }')
echo $REGISTER_RESPONSE | jq .

echo ""
echo "3️⃣ Testing Login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')
echo $LOGIN_RESPONSE | jq .

# Extract JWT token
JWT_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token // .token // empty')

if [ ! -z "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
  echo ""
  echo "🔑 JWT Token received: ${JWT_TOKEN:0:50}..."
  
  echo ""
  echo "4️⃣ Testing Profile Access..."
  curl -s $BASE_URL/api/auth/profile \
    -H "Authorization: Bearer $JWT_TOKEN" | jq .
  
  echo ""
  echo "5️⃣ Testing OAuth Client Creation..."
  CLIENT_RESPONSE=$(curl -s -X POST $BASE_URL/api/auth/clients \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -d '{
      "name": "Test OAuth Client",
      "description": "OAuth client for testing",
      "redirect_uris": ["http://localhost:3000/callback"],
      "scopes": ["read", "write"],
      "grant_types": ["authorization_code", "refresh_token"]
    }')
  echo $CLIENT_RESPONSE | jq .
  
  CLIENT_ID=$(echo $CLIENT_RESPONSE | jq -r '.client_id // empty')
  CLIENT_SECRET=$(echo $CLIENT_RESPONSE | jq -r '.client_secret // empty')
  
  if [ ! -z "$CLIENT_ID" ] && [ "$CLIENT_ID" != "null" ]; then
    echo ""
    echo "🔐 OAuth Client created:"
    echo "   Client ID: $CLIENT_ID"
    echo "   Client Secret: ${CLIENT_SECRET:0:20}..."
    
    echo ""
    echo "6️⃣ Testing OAuth Discovery..."
    curl -s $BASE_URL/api/auth/oauth/.well-known/openid-configuration | jq .
    
    echo ""
    echo "7️⃣ OAuth Authorization URL:"
    echo "   $BASE_URL/api/auth/oauth/authorize?response_type=code&client_id=$CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=read&state=test123"
  else
    echo "❌ Failed to create OAuth client"
  fi
else
  echo "❌ Failed to get JWT token from login"
fi

echo ""
echo "✅ API Testing Complete!"