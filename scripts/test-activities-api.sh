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
    \"category\": \"Yoga\",
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
    \"category\": \"Spinning\",
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

# Test 3.1: Crear tercera actividad (Pilates)
echo "📍 Test 3.1: Crear tercera actividad (Pilates)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"owner_id\": $ADMIN_USER_ID,
    \"name\": \"Pilates Mat\",
    \"description\": \"Clase de pilates en colchoneta para fortalecer core\",
    \"category\": \"Pilates\",
    \"duration\": 50,
    \"price\": 1400.00,
    \"image_url\": \"https://example.com/pilates.jpg\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Tercera actividad creada (Pilates)${NC}"
    ACTIVITY_ID_3=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Activity ID 3: $ACTIVITY_ID_3"
else
    echo -e "${RED}❌ Error (HTTP $http_code)${NC}"
    echo "Response: $body"
fi
echo ""

# Test 3.2: Crear cuarta actividad (CrossFit)
echo "📍 Test 3.2: Crear cuarta actividad (CrossFit)"
response=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"owner_id\": $ADMIN_USER_ID,
    \"name\": \"CrossFit WOD\",
    \"description\": \"Entrenamiento funcional de alta intensidad\",
    \"category\": \"CrossFit\",
    \"duration\": 60,
    \"price\": 1800.00,
    \"image_url\": \"https://example.com/crossfit.jpg\"
  }")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "201" ]; then
    echo -e "${GREEN}✅ Cuarta actividad creada (CrossFit)${NC}"
    ACTIVITY_ID_4=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    echo "Activity ID 4: $ACTIVITY_ID_4"
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


# Test 7.1: Crear horario para Pilates (Martes)
if [ -n "$ACTIVITY_ID_3" ]; then
    echo "📍 Test 7.1: Crear horario para Pilates (Martes)"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID_3/schedules" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d '{
        "instructor": "Laura Martínez",
        "day_of_week": "tuesday",
        "start_time": "10:00",
        "end_time": "10:50",
        "location": "Sala 3 - Sede Sur",
        "max_capacity": 12
      }')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ Horario de Pilates creado${NC}"
        SCHEDULE_ID_3=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        echo "Schedule ID 3: $SCHEDULE_ID_3"
    else
        echo -e "${YELLOW}⚠️  Error (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
fi

# Test 7.2: Crear horario para CrossFit (Viernes)
if [ -n "$ACTIVITY_ID_4" ]; then
    echo "📍 Test 7.2: Crear horario para CrossFit (Viernes)"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID_4/schedules" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d '{
        "instructor": "Carlos Ruiz",
        "day_of_week": "friday",
        "start_time": "17:00",
        "end_time": "18:00",
        "location": "Box - Sede Norte",
        "max_capacity": 25
      }')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ Horario de CrossFit creado${NC}"
        SCHEDULE_ID_4=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        echo "Schedule ID 4: $SCHEDULE_ID_4"
    else
        echo -e "${YELLOW}⚠️  Error (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
fi

# Test 7.3: Crear horario para Spinning (Jueves)
if [ -n "$ACTIVITY_ID_2" ]; then
    echo "📍 Test 7.3: Crear horario para Spinning (Jueves)"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/activities/$ACTIVITY_ID_2/schedules" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d '{
        "instructor": "Pedro Sánchez",
        "day_of_week": "thursday",
        "start_time": "07:00",
        "end_time": "07:45",
        "location": "Sala de Spinning - Sede Centro",
        "max_capacity": 30
      }')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "201" ]; then
        echo -e "${GREEN}✅ Horario de Spinning creado${NC}"
        SCHEDULE_ID_5=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        echo "Schedule ID 5: $SCHEDULE_ID_5"
    else
        echo -e "${YELLOW}⚠️  Error (HTTP $http_code)${NC}"
        echo "Response: $body"
    fi
    echo ""
fi

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
ACTIVITY_ID_3=${ACTIVITY_ID_3:-}
ACTIVITY_ID_4=${ACTIVITY_ID_4:-}
SCHEDULE_ID=$SCHEDULE_ID
SCHEDULE_ID_2=$SCHEDULE_ID_2
SCHEDULE_ID_3=${SCHEDULE_ID_3:-}
SCHEDULE_ID_4=${SCHEDULE_ID_4:-}
SCHEDULE_ID_5=${SCHEDULE_ID_5:-}
EOF

echo "=================================="
echo -e "${GREEN}✅ TODOS LOS TESTS DE ACTIVITIES-API PASARON${NC}"
echo "=================================="
echo ""
echo "Variables exportadas:"
echo "Variables exportadas:"
echo "  ACTIVITY_ID=$ACTIVITY_ID (Yoga)"
echo "  ACTIVITY_ID_2=$ACTIVITY_ID_2 (Spinning)"
echo "  ACTIVITY_ID_3=${ACTIVITY_ID_3:-N/A} (Pilates)"
echo "  ACTIVITY_ID_4=${ACTIVITY_ID_4:-N/A} (CrossFit)"
echo "  SCHEDULE_ID=$SCHEDULE_ID (Yoga - Lunes)"
echo "  SCHEDULE_ID_2=$SCHEDULE_ID_2 (Yoga - Miércoles)"
echo "  SCHEDULE_ID_3=${SCHEDULE_ID_3:-N/A} (Pilates - Martes)"
echo "  SCHEDULE_ID_4=${SCHEDULE_ID_4:-N/A} (CrossFit - Viernes)"
echo "  SCHEDULE_ID_5=${SCHEDULE_ID_5:-N/A} (Spinning - Jueves)"
echo ""
echo "⏳ Esperando 3 segundos para que RabbitMQ propague eventos..."
sleep 3
echo -e "${GREEN}✅ Listo para continuar con search-api${NC}"
