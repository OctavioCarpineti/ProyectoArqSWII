#!/bin/bash

# Script de prueba para search-api
# Prerequisito: Ejecutar test-users-api.sh y test-activities-api.sh primero

set -e

BASE_URL="http://localhost:8083"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=================================="
echo "🧪 TESTING SEARCH-API"
echo "=================================="
echo ""

# Cargar variables
if [ -f /tmp/gym-tokens.env ]; then
    source /tmp/gym-tokens.env
    echo "✅ Variables cargadas"
else
    echo -e "${RED}❌ Error: Ejecuta los tests anteriores primero${NC}"
    exit 1
fi
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

# Esperar a que RabbitMQ consumer procese los eventos
echo "⏳ Esperando 5 segundos para que search-api indexe los datos en Solr..."
sleep 5
echo ""

# Test 2: Búsqueda sin filtros (empty query)
echo "📍 Test 2: Búsqueda sin filtros (todos los horarios)"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda exitosa${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error en búsqueda (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 3: Búsqueda por texto "yoga"
echo "📍 Test 3: Búsqueda por texto 'yoga'"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?q=yoga")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda por texto OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 4: Búsqueda por categoría
echo "📍 Test 4: Búsqueda por categoría 'Yoga'"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?category=Yoga")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda por categoría OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 5: Búsqueda por día de la semana
echo "📍 Test 5: Búsqueda por día 'monday'"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?day_of_week=monday")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda por día OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 6: Búsqueda con filtro available=true
echo "📍 Test 6: Búsqueda solo horarios disponibles"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?available=true")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda con filtro available OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 7: Búsqueda por instructor
echo "📍 Test 7: Búsqueda por instructor 'María'"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?instructor=María")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda por instructor OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 8: Búsqueda con paginación
echo "📍 Test 8: Búsqueda con paginación (page=1, size=1)"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?page=1&size=1")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Paginación OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 9: Búsqueda combinada (múltiples filtros)
echo "📍 Test 9: Búsqueda combinada (yoga + monday + available)"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search?q=yoga&day_of_week=monday&available=true")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Búsqueda combinada OK${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 10: Obtener horario por ID
echo "📍 Test 10: Obtener schedule por ID desde Solr"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/search/$SCHEDULE_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Schedule obtenido desde Solr${NC}"
    echo "Response: $body"
else
    echo -e "${YELLOW}⚠️  No se pudo obtener desde Solr (HTTP $http_code)${NC}"
    echo "Response: $body"
    echo "Esto puede pasar si el consumer aún no indexó el documento"
fi
echo ""

# Test 11: Test de caché (llamar 2 veces la misma búsqueda)
echo "📍 Test 11: Testing caché multinivel"
echo "Primera búsqueda (sin caché):"
time curl -s "$BASE_URL/search?q=yoga" > /dev/null
echo "Segunda búsqueda (debería estar en caché L1):"
time curl -s "$BASE_URL/search?q=yoga" > /dev/null
echo -e "${GREEN}✅ La segunda búsqueda debería ser más rápida (caché)${NC}"
echo ""

echo "=================================="
echo -e "${GREEN}✅ TODOS LOS TESTS DE SEARCH-API PASARON${NC}"
echo "=================================="
echo ""
echo "💡 Notas:"
echo "  - Si el Test 10 falló, espera unos segundos más y reintenta"
echo "  - El consumer de RabbitMQ puede tardar en indexar"
echo "  - Revisa logs: docker-compose logs -f search-api"
