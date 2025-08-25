#!/bin/bash

echo "Testing login endpoint..."

# Test the login endpoint
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123"
  }' \
  -v

echo ""
echo "Testing health endpoint..."

# Test health endpoint first
curl -X GET http://localhost:8080/api/v1/healthz -v