#!/bin/bash

# Script para limpiar completamente el proyecto
# Elimina todos los contenedores, volúmenes y datos

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   🧹 LIMPIEZA COMPLETA DEL SISTEMA     ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Ir a la raíz del proyecto
cd "$(dirname "$0")/.."

echo -e "${YELLOW}⚠️  ADVERTENCIA: Esto eliminará:${NC}"
echo "   - Todos los contenedores"
echo "   - Todos los volúmenes (datos de MySQL, MongoDB, etc.)"
echo "   - Cache de RabbitMQ, Solr, Memcached"
echo ""
read -p "¿Continuar? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}Operación cancelada${NC}"
    exit 0
fi

echo ""
echo -e "${BLUE}📦 Paso 1/4: Deteniendo contenedores...${NC}"
docker-compose down
echo -e "${GREEN}✅ Contenedores detenidos${NC}"
echo ""

echo -e "${BLUE}🗑️  Paso 2/4: Eliminando volúmenes...${NC}"
docker-compose down -v
echo -e "${GREEN}✅ Volúmenes eliminados${NC}"
echo ""

echo -e "${BLUE}🧹 Paso 3/4: Limpiando archivos temporales...${NC}"
rm -f /tmp/gym-tokens.env
echo -e "${GREEN}✅ Archivos temporales eliminados${NC}"
echo ""

echo -e "${BLUE}🔨 Paso 4/4: Limpiando sistema Docker...${NC}"
docker system prune -f
echo -e "${GREEN}✅ Sistema Docker limpio${NC}"
echo ""

echo "╔════════════════════════════════════════╗"
echo "║        ✅ LIMPIEZA COMPLETADA          ║"
echo "╚════════════════════════════════════════╝"
echo ""
echo -e "${GREEN}El sistema está completamente limpio.${NC}"
echo ""
echo "Próximos pasos:"
echo "  1. Levantar servicios: docker-compose up -d"
echo "  2. Esperar 10-15 segundos"
echo "  3. Ejecutar tests: ./scripts/test-all.sh"
echo ""
echo "O ejecutar todo de una vez:"
echo "  ./scripts/clean-and-test.sh"
echo ""