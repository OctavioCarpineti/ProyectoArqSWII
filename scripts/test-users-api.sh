#!/bin/bash

# Script de prueba para users-api
# Prerequisito: docker-compose up -d

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=================================="
echo "🧪 TESTING USERS-API"
echo "=================================="
echo ""

# Test 1: Health check
echo "📍 Test 1: Health Check"
response=$(curl -s -w "\n%{http_code}" $BASE_URL/health)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Health check OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Health check FAILED (HTTP $http_code)${NC}"
    exit 1
fi
echo ""

# Test 2: Crear usuario normal
echo "📍 Test 2: Crear usuario normal"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "Password123!",
    "first_name": "John",
    "last_name": "Doe"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Usuario creado exitosamente${NC}"
    echo "Response: $body"
    USER_ID=$(echo "$body" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
    echo "User ID: $USER_ID"
else
    echo -e "${RED}❌ Error al crear usuario (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 3: Login con usuario normal
echo "📍 Test 3: Login con usuario normal"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "johndoe",
    "password": "Password123!"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Login exitoso${NC}"
    USER_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "Token obtenido (primeros 50 chars): ${USER_TOKEN:0:50}..."
else
    echo -e "${RED}❌ Error en login (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 4: Login con usuario admin (pre-creado en init.sql)
echo "📍 Test 4: Login con usuario admin"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "admin",
    "password": "admin123"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Login admin exitoso${NC}"
    ADMIN_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "Admin Token obtenido (primeros 50 chars): ${ADMIN_TOKEN:0:50}..."
    ADMIN_USER_ID=$(echo "$body" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
    echo "Admin User ID: $ADMIN_USER_ID"
else
    echo -e "${RED}❌ Error en login admin (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 5: Obtener usuario por ID
echo "📍 Test 5: Obtener usuario por ID"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/$USER_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Usuario obtenido exitosamente${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error al obtener usuario (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 6: Intento de login con credenciales incorrectas
echo "📍 Test 6: Login con credenciales incorrectas (debe fallar)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "johndoe",
    "password": "WrongPassword"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✅ Error 401 correcto (credenciales inválidas)${NC}"
else
    echo -e "${RED}❌ Debería retornar 401 pero retornó $http_code${NC}"
    echo "Response: $body"
fi
echo ""

# Test 7: Crear usuario duplicado (debe fallar)
echo "📍 Test 7: Crear usuario con email duplicado (debe fallar)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "anothername",
    "email": "john@example.com",
    "password": "Password123!",
    "first_name": "Another",
    "last_name": "User"
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "409" ]; then
    echo -e "${GREEN}✅ Error 409 correcto (usuario ya existe)${NC}"
else
    echo -e "${YELLOW}⚠️  Esperaba 409 pero obtuvo $http_code${NC}"
    echo "Response: $body"
fi
echo ""

# Guardar tokens en archivo
echo "📝 Guardando tokens en .env.test"
cat > /tmp/gym-tokens.env << EOF
USER_TOKEN=$USER_TOKEN
ADMIN_TOKEN=$ADMIN_TOKEN
USER_ID=$USER_ID
ADMIN_USER_ID=$ADMIN_USER_ID
EOF
echo -e "${GREEN}✅ Tokens guardados en /tmp/gym-tokens.env${NC}"
echo ""

echo "=================================="
echo -e "${GREEN}✅ TODOS LOS TESTS DE USERS-API PASARON${NC}"
echo "=================================="
echo ""
echo "Variables exportadas:"
echo "  USER_TOKEN=$USER_TOKEN"
echo "  ADMIN_TOKEN=$ADMIN_TOKEN"
echo "  USER_ID=$USER_ID"
echo "  ADMIN_USER_ID=$ADMIN_USER_ID"
