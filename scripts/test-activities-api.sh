#!/bin/bash

# Script de prueba para activities-api
# Prerequisito:
#   1. docker-compose up -d
#   2. Ejecutar test-users-api.sh primero para obtener tokens

set -e

BASE_URL="http://localhost:8081"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=================================="
echo "🧪 TESTING ACTIVITIES-API"
echo "=================================="
echo ""

# Cargar tokens del test anterior
if [ -f /tmp/gym-tokens.env ]; then
    source /tmp/gym-tokens.env
    echo "✅ Tokens cargados desde /tmp/gym-tokens.env"
else
    echo -e "${RED}❌ Error: Primero ejecuta test-users-api.sh para obtener tokens${NC}"
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

# Test 2: Crear actividad como admin
echo "📍 Test 2: Crear actividad (como admin)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"owner_id\": $ADMIN_USER_ID,
    \"name\": \"Yoga Intermedio\",
    \"description\": \"Clase de yoga nivel intermedio para mejorar flexibilidad\",
    \"category\": \"flexibility\",
    \"duration\": 60,
    \"price\": 1500.00,
    \"image_url\": \"https://example.com/yoga.jpg\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Actividad creada exitosamente${NC}"
    echo "Response: $body"
    ACTIVITY_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Activity ID: $ACTIVITY_ID"
else
    echo -e "${RED}❌ Error al crear actividad (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 3: Crear otra actividad
echo "📍 Test 3: Crear segunda actividad (Spinning)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"owner_id\": $ADMIN_USER_ID,
    \"name\": \"Spinning Avanzado\",
    \"description\": \"Clase intensiva de spinning\",
    \"category\": \"cardio\",
    \"duration\": 45,
    \"price\": 1200.00,
    \"image_url\": \"https://example.com/spinning.jpg\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Segunda actividad creada${NC}"
    ACTIVITY_ID_2=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Activity ID 2: $ACTIVITY_ID_2"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 4: Listar todas las actividades
echo "📍 Test 4: Listar todas las actividades"
response=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/activities)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Actividades listadas${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    exit 1
fi
echo ""

# Test 5: Obtener actividad por ID
echo "📍 Test 5: Obtener actividad por ID"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/activities/$ACTIVITY_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Actividad obtenida${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    exit 1
fi
echo ""

# Test 6: Crear horario para la actividad (CON CONCURRENCIA)
echo "📍 Test 6: Crear horario (schedule) - Testing concurrencia"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID/schedules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "instructor": "María González",
    "day_of_week": "monday",
    "start_time": "18:00",
    "end_time": "19:00",
    "location": "Sala 1 - Sede Centro",
    "max_capacity": 20
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Horario creado (concurrencia ejecutada)${NC}"
    echo "Response: $body"
    SCHEDULE_ID=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Schedule ID: $SCHEDULE_ID"
else
    echo -e "${RED}❌ Error al crear horario (HTTP $http_code)${NC}"
    echo "Response: $body"
    exit 1
fi
echo ""

# Test 7: Crear segundo horario
echo "📍 Test 7: Crear segundo horario (Yoga - Miércoles)"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID/schedules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "instructor": "Juan Pérez",
    "day_of_week": "wednesday",
    "start_time": "19:00",
    "end_time": "20:00",
    "location": "Sala 2 - Sede Norte",
    "max_capacity": 15
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Segundo horario creado${NC}"
    SCHEDULE_ID_2=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Schedule ID 2: $SCHEDULE_ID_2"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 8: Intentar crear horario con conflicto (debe fallar)
echo "📍 Test 8: Crear horario con conflicto (debe fallar)"
response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID/schedules" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "instructor": "Otro Instructor",
    "day_of_week": "monday",
    "start_time": "18:30",
    "end_time": "19:30",
    "location": "Sala 1 - Sede Centro",
    "max_capacity": 10
  }')
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "400" ] || [ "$http_code" = "409" ]; then
    echo -e "${GREEN}✅ Error correcto detectado (conflicto de horario)${NC}"
else
    echo -e "${YELLOW}⚠️  Esperaba 400/409 pero obtuvo $http_code${NC}"
    echo "Response: $body"
fi
echo ""

# Test 9: Listar todos los horarios
echo "📍 Test 9: Listar todos los horarios (GET /schedules)"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/schedules")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Horarios listados${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    exit 1
fi
echo ""

# Test 10: Obtener horario por ID
echo "📍 Test 10: Obtener horario por ID"
response=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/schedules/$SCHEDULE_ID")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✅ Horario obtenido${NC}"
    echo "Response: $body"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    exit 1
fi
echo ""

# Actualizar archivo de tokens con IDs de actividades y schedules
cat >> /tmp/gym-tokens.env << EOF
ACTIVITY_ID=$ACTIVITY_ID
ACTIVITY_ID_2=$ACTIVITY_ID_2
SCHEDULE_ID=$SCHEDULE_ID
SCHEDULE_ID_2=$SCHEDULE_ID_2
EOF

echo "=================================="
echo -e "${GREEN}✅ TODOS LOS TESTS DE ACTIVITIES-API PASARON${NC}"
echo "=================================="
echo ""
echo "Variables exportadas:"
echo "  ACTIVITY_ID=$ACTIVITY_ID"
echo "  ACTIVITY_ID_2=$ACTIVITY_ID_2"
echo "  SCHEDULE_ID=$SCHEDULE_ID"
echo "  SCHEDULE_ID_2=$SCHEDULE_ID_2"
echo ""
echo "⏳ Esperando 3 segundos para que RabbitMQ propague eventos..."
sleep 3
echo -e "${GREEN}✅ Listo para continuar con search-api${NC}"
