#!/bin/bash
# Test script for enhanced agent API with all configuration fields

set -e

API_KEY="${API_KEY:-ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7}"
AGENT_URL="${AGENT_URL:-http://localhost:9090}"

echo "=== Testing Enhanced Agent API ==="
echo "Agent URL: $AGENT_URL"
echo ""

# Test 1: Minimal request (only required fields)
echo "Test 1: Minimal request (only required fields)"
curl -X POST "$AGENT_URL/instances" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-minimal",
    "api_endpoint": "10.33.7.80",
    "username": "monitor_user",
    "password": "secret123"
  }' | jq .
echo ""
echo "---"
echo ""

# Test 2: Request with documentation fields
echo "Test 2: Request with documentation fields"
curl -X POST "$AGENT_URL/instances" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-documented",
    "api_endpoint": "10.33.7.81",
    "username": "superuser",
    "password": "secret456",
    "storage_system": "svc_cluster01",
    "storage_description": "IBM FlashSystem 9200 - Production Site A",
    "storage_location": "Datacenter A, Rack 12",
    "ipquorum_name": "prodquorum01"
  }' | jq .
echo ""
echo "---"
echo ""

# Test 3: Full request with mkquorumapp enabled
echo "Test 3: Full request with mkquorumapp enabled"
curl -X POST "$AGENT_URL/instances" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-full",
    "api_endpoint": "10.33.7.82",
    "username": "admin",
    "password": "secret789",
    "storage_system": "svc_cluster01",
    "storage_description": "IBM FlashSystem 9200 - Production Site A",
    "storage_location": "Datacenter A, Rack 12",
    "download_enabled": true,
    "mkquorumapp_enabled": true,
    "partnersystem": "svc_cluster02",
    "ip6": false,
    "partnerip6": false,
    "nometadata": false
  }' | jq .
echo ""
echo "---"
echo ""

# Test 4: Invalid request - mkquorumapp enabled but no partnersystem (should fail)
echo "Test 4: Invalid request - mkquorumapp enabled but no partnersystem (should fail)"
curl -X POST "$AGENT_URL/instances" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-invalid",
    "api_endpoint": "10.33.7.83",
    "username": "admin",
    "password": "secret000",
    "mkquorumapp_enabled": true
  }' | jq .
echo ""
echo "---"
echo ""

# Test 5: List all instances
echo "Test 5: List all instances"
curl -X GET "$AGENT_URL/instances" \
  -H "X-API-Key: $API_KEY" | jq .
echo ""
echo "---"
echo ""

echo "=== Test Complete ==="
echo ""
echo "To verify config files, check:"
echo "  sudo cat /etc/ipquorum/instances/test-minimal.conf"
echo "  sudo cat /etc/ipquorum/instances/test-documented.conf"
echo "  sudo cat /etc/ipquorum/instances/test-full.conf"
echo ""
echo "To clean up test instances:"
echo "  curl -X DELETE '$AGENT_URL/instances/test-minimal' -H 'X-API-Key: $API_KEY'"
echo "  curl -X DELETE '$AGENT_URL/instances/test-documented' -H 'X-API-Key: $API_KEY'"
echo "  curl -X DELETE '$AGENT_URL/instances/test-full' -H 'X-API-Key: $API_KEY'"





curl -X POST http://localhost:9090/instances   -H "X-API-Key: ea17f6738a9db12f157265dc833154615f05c8c294269b4ced57d20684e6d4d7"   -H "Content-Type: application/json"   -d '{
    "name": "test01",
    "api_endpoint": "10.33.7.80",
    "username": "monitor",
    "password": "secret"
    "mkquorumapp_enabled": true,
    "partnersystem": "svc_cluster02"
  }'