#!/bin/bash

# Script maestro: ejecuta todos los tests en orden
# Prerequisito: docker-compose up -d

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   🧪 GYM BOOKING SYSTEM - E2E TESTS   ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Verificar que docker-compose esté corriendo
echo "🔍 Verificando que los servicios estén corriendo..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: users-api no está respondiendo${NC}"
    echo "Por favor ejecuta: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8081/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: activities-api no está respondiendo${NC}"
    echo "Por favor ejecuta: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: bookings-api no está respondiendo${NC}"
    echo "Por favor ejecuta: docker-compose up -d"
    exit 1
fi

if ! curl -s http://localhost:8083/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Error: search-api no está respondiendo${NC}"
    echo "Por favor ejecuta: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}✅ Todos los servicios están corriendo${NC}"
echo ""

# Limpiar archivo de tokens anterior
rm -f /tmp/gym-tokens.env

# Test 1: users-api
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 1/4: Testing users-api       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
bash "$SCRIPT_DIR/test-users-api.sh"
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error en users-api${NC}"
    exit 1
fi
echo ""
sleep 2

# Test 2: activities-api
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 2/4: Testing activities-api   ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
bash "$SCRIPT_DIR/test-activities-api.sh"
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error en activities-api${NC}"
    exit 1
fi
echo ""
sleep 2

# Test 3: search-api
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 3/4: Testing search-api       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
bash "$SCRIPT_DIR/test-search-api.sh"
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error en search-api${NC}"
    exit 1
fi
echo ""
sleep 2

# Test 4: bookings-api
echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 4/4: Testing bookings-api     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
bash "$SCRIPT_DIR/test-bookings-api.sh"
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error en bookings-api${NC}"
    exit 1
fi
echo ""

# Resumen final
echo ""
echo "╔════════════════════════════════════════╗"
echo "║         🎉 TODOS LOS TESTS PASARON     ║"
echo "╚════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}✅ Flujo completo verificado:${NC}"
echo "   1. ✅ Creación de usuarios y login (JWT)"
echo "   2. ✅ Creación de actividades y horarios (con concurrencia)"
echo "   3. ✅ Publicación de eventos a RabbitMQ"
echo "   4. ✅ Consumo de eventos e indexación en Solr"
echo "   5. ✅ Búsquedas con filtros y caché multinivel"
echo "   6. ✅ Creación de reservas (con concurrencia y validaciones)"
echo "   7. ✅ Actualización de current_bookings"
echo "   8. ✅ Soft delete de reservas"
echo ""
echo -e "${GREEN}🚀 El backend está funcionando PERFECTAMENTE${NC}"
echo ""
echo "📋 Próximos pasos:"
echo "   1. Implementar frontend React"
echo "   2. Crear tests unitarios en Go"
echo ""
echo "📁 Tokens y variables guardados en: /tmp/gym-tokens.env"
echo ""
