#!/bin/bash

# Script maestro: Limpia todo y ejecuta tests completos
# Este script garantiza un entorno limpio antes de cada ejecución

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   🧪 CLEAN & TEST - GYM BOOKING SYSTEM ║"
echo "╚════════════════════════════════════════╝"
echo ""
echo -e "${BLUE}Este script realizará:${NC}"
echo "  1. 🧹 Limpieza completa del sistema"
echo "  2. 🚀 Levantar todos los servicios"
echo "  3. ⏳ Esperar a que estén listos"
echo "  4. 🧪 Ejecutar tests E2E completos"
echo ""

# Ir a la raíz del proyecto
cd "$SCRIPT_DIR/.."

# ==========================================
# FASE 1: LIMPIEZA
# ==========================================

echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 1/4: Limpiando sistema       ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

echo "🧹 Deteniendo contenedores..."
docker-compose down -v > /dev/null 2>&1 || true

echo "🗑️  Eliminando volúmenes..."
docker volume rm proyectoarqswii_mysql_data > /dev/null 2>&1 || true
docker volume rm proyectoarqswii_mongodb_data > /dev/null 2>&1 || true
docker volume rm proyectoarqswii_rabbitmq_data > /dev/null 2>&1 || true
docker volume rm proyectoarqswii_solr_data > /dev/null 2>&1 || true

echo "🧹 Limpiando archivos temporales..."
rm -f /tmp/gym-tokens.env

echo -e "${GREEN}✅ Limpieza completada${NC}"
echo ""
sleep 2

# ==========================================
# FASE 2: LEVANTAR SERVICIOS
# ==========================================

echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 2/4: Levantando servicios    ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

echo "🚀 Levantando servicios con docker-compose..."
docker-compose up -d

echo -e "${GREEN}✅ Servicios iniciados${NC}"
echo ""

# ==========================================
# FASE 3: ESPERAR A QUE ESTÉN LISTOS
# ==========================================

echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 3/4: Esperando servicios     ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

echo "⏳ Esperando a que los servicios estén listos..."
echo ""

# Función para verificar health de un servicio
check_health() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $name listo${NC}"
            return 0
        fi
        echo "   Intento $attempt/$max_attempts: Esperando $name..."
        sleep 2
        attempt=$((attempt + 1))
    done

    echo -e "${RED}❌ Timeout esperando $name${NC}"
    return 1
}

# Verificar cada servicio
check_health "http://localhost:8080/health" "users-api"
check_health "http://localhost:8081/health" "activities-api"
check_health "http://localhost:8082/health" "bookings-api"
check_health "http://localhost:8083/health" "search-api"

echo ""
echo -e "${GREEN}✅ Todos los servicios están listos${NC}"
echo ""

# Espera adicional para que RabbitMQ y Solr estén completamente inicializados
echo "⏳ Esperando 5 segundos adicionales para inicialización completa..."
sleep 5
echo ""

# ==========================================
# FASE 4: EJECUTAR TESTS
# ==========================================

echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   FASE 4/4: Ejecutando tests        ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
echo ""

# Ejecutar el script de tests
bash "$SCRIPT_DIR/test-all.sh"

exit_code=$?

# ==========================================
# RESUMEN FINAL
# ==========================================

echo ""
echo "╔════════════════════════════════════════╗"
if [ $exit_code -eq 0 ]; then
    echo "║     🎉 TODOS LOS TESTS PASARON ✅      ║"
else
    echo "║        ❌ ALGUNOS TESTS FALLARON       ║"
fi
echo "╚════════════════════════════════════════╝"
echo ""

if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}✅ El sistema está funcionando perfectamente${NC}"
    echo ""
    echo "📋 Próximos pasos sugeridos:"
    echo "   1. Implementar frontend React"
    echo "   2. Crear tests unitarios en Go"
    echo "   3. Eliminar código duplicado (opcional)"
    echo ""
else
    echo -e "${RED}⚠️  Algunos tests fallaron. Revisa los logs arriba.${NC}"
    echo ""
    echo "Para ver logs de un servicio específico:"
    echo "   docker-compose logs -f <servicio>"
    echo ""
fi

exit $exit_code