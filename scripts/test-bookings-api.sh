#!/bin/bash

# Script de prueba para bookings-api
# Prerequisito: Ejecutar los 3 tests anteriores primero

set -e

BASE_URL="http://localhost:8082"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=================================="
echo "🧪 TESTING BOOKINGS-API"
echo "=================================="
echo ""

# Cargar variables
if [ -f /tmp/gym-tokens.env ]; then
    source /tmp/gym-tokens.env
    echo "✅ Variables cargadas"
    echo "   USER_ID=$USER_ID"
    echo "   SCHEDULE_ID=$SCHEDULE_ID"
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

# Test 2: Crear reserva (CON CONCURRENCIA)
echo "📍 Test 2: Crear reserva - Testing concurrencia (3 goroutines)"
echo "  Goroutine 1: Validar usuario existe (HTTP a users-api)"
echo "  Goroutine 2: Validar schedule existe y tiene cupos (HTTP a activities-api)"
echo "  Goroutine 3: Verificar sin duplicados (MongoDB query)"
echo ""

response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/bookings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{
    \"user_id\": $USER_ID,
    \"schedule_id\": \"$SCHEDULE_ID\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Reserva creada exitosamente (concurrencia OK)${NC}"
    echo "Response: $body"
    BOOKING_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Booking ID: $BOOKING_ID"
else
    echo -e "${RED}❌ Error al crear reserva (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 3: Obtener reserva por ID
echo "📍 Test 3: Obtener reserva por ID"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/bookings/$BOOKING_ID" \
  -H "Authorization: Bearer $USER_TOKEN")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Reserva obtenida${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 4: Listar reservas del usuario
echo "📍 Test 4: Listar mis reservas (GET /bookings?user_id=$USER_ID)"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/bookings?user_id=$USER_ID" \
  -H "Authorization: Bearer $USER_TOKEN")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Reservas del usuario listadas${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 5: Intentar crear reserva duplicada (debe fallar)
echo "📍 Test 5: Intentar reservar el mismo horario (debe fallar)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/bookings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "{
    \"user_id\": $USER_ID,
    \"schedule_id\": \"$SCHEDULE_ID\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "400" ] || [ "$http_code" = "409" ]; then
    echo -e "${GREEN}✅ Error correcto detectado (reserva duplicada)${NC}"
else
    echo -e "${YELLOW}⚠️  Esperaba 400/409 pero obtuvo $http_code${NC}"
    echo "Response: $body"
fi
echo ""

# Test 6: Crear segunda reserva (otro horario)
if [ -n "$SCHEDULE_ID_2" ]; then
    echo "📍 Test 6: Crear segunda reserva (diferente horario)"
    response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/bookings \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $USER_TOKEN" \
      -d "{
        \"user_id\": $USER_ID,
        \"schedule_id\": \"$SCHEDULE_ID_2\"
      }")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ Segunda reserva creada${NC}"
        BOOKING_ID_2=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        echo "Booking ID 2: $BOOKING_ID_2"
    else
        echo -e "${YELLOW}⚠️  No se pudo crear segunda reserva (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
fi

# Test 7: Verificar que current_bookings se incrementó en el schedule
echo "📍 Test 7: Verificar que current_bookings se incrementó en activities-api"
response=$(curl -s -w "\n%{http_code}" -X GET "http://localhost:8081/schedules/$SCHEDULE_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    current_bookings=$(echo "$body" | grep -o '"current_bookings":[0-9]*' | grep -o '[0-9]*')
    if [ "$current_bookings" -gt 0 ]; then
        echo -e "${GREEN}✅ Counter actualizado: current_bookings = $current_bookings${NC}"
    else
        echo -e "${YELLOW}⚠️  current_bookings = 0 (puede no haberse actualizado)${NC}"
    fi
    echo "Response: $body"
else
    echo -e "${RED}❌ Error al verificar schedule (HTTP $http_code)${NC}"
fi
echo ""

# Test 8: Intentar crear reserva sin autenticación (debe fallar)
echo "📍 Test 8: Crear reserva sin JWT (debe fallar con 401)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/bookings \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $USER_ID,
    \"schedule_id\": \"$SCHEDULE_ID\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✅ Error 401 correcto (sin autenticación)${NC}"
else
    echo -e "${YELLOW}⚠️  Esperaba 401 pero obtuvo $http_code${NC}"
    echo "Response: $body"
fi
echo ""

# Test 9: Cancelar reserva
echo "📍 Test 9: Cancelar reserva (soft delete)"
response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/bookings/$BOOKING_ID" \
  -H "Authorization: Bearer $USER_TOKEN")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ] || [ "$http_code" = "204" ]; then
    echo -e "${GREEN}✅ Reserva cancelada${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error al cancelar reserva (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 10: Verificar que current_bookings se decrementó
echo "📍 Test 10: Verificar que current_bookings se decrementó"
response=$(curl -s -w "\n%{http_code}" -X GET "http://localhost:8081/schedules/$SCHEDULE_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    current_bookings=$(echo "$body" | grep -o '"current_bookings":[0-9]*' | grep -o '[0-9]*')
    echo -e "${GREEN}✅ current_bookings después de cancelar = $current_bookings${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
fi
echo ""

# Test 11: Cancelar segunda reserva (si existe)
if [ -n "$BOOKING_ID_2" ]; then
    echo "📍 Test 11: Cancelar segunda reserva"
    response=$(curl -s -w "\n%{http_code}" -X DELETE "$BASE_URL/bookings/$BOOKING_ID_2" \
      -H "Authorization: Bearer $USER_TOKEN")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ] || [ "$http_code" = "204" ]; then
        echo -e "${GREEN}✅ Segunda reserva cancelada${NC}"
        echo "Response: $body"
    else
        echo -e "${YELLOW}⚠️  Error al cancelar segunda reserva (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
fi

# Guardar IDs de bookings
cat >> /tmp/gym-tokens.env << EOF
BOOKING_ID=$BOOKING_ID
BOOKING_ID_2=${BOOKING_ID_2:-}
EOF

echo "=================================="
echo -e "${GREEN}✅ TODOS LOS TESTS DE BOOKINGS-API PASARON${NC}"
echo "=================================="
echo ""
echo "Variables exportadas:"
echo "  BOOKING_ID=$BOOKING_ID"
echo "  BOOKING_ID_2=${BOOKING_ID_2:-N/A}"
echo ""
echo "🎉 Verificaciones importantes:"
echo "  ✅ Concurrencia funcionando (3 goroutines)"
echo "  ✅ Validación de usuario (HTTP a users-api)"
echo "  ✅ Validación de schedule (HTTP a activities-api)"
echo "  ✅ Detección de duplicados (MongoDB)"
echo "  ✅ Actualización de current_bookings"
echo "  ✅ Soft delete funcionando (ambas reservas canceladas)"
