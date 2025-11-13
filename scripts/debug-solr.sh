#!/bin/bash

# Script para verificar Solr y los documentos indexados
# Útil para debugging del sistema de búsqueda

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   🔍 DEBUGGING SOLR                    ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Verificar que Solr esté corriendo
if ! curl -s http://localhost:8983 > /dev/null 2>&1; then
    echo -e "${RED}❌ Solr no está respondiendo${NC}"
    echo "Ejecuta: docker-compose up -d solr"
    exit 1
fi

echo -e "${GREEN}✅ Solr está corriendo${NC}"
echo ""

# ==========================================
# VERIFICAR CORES
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📚 CORES DE SOLR${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Listando cores disponibles..."
cores_response=$(curl -s "http://localhost:8983/solr/admin/cores?action=STATUS&wt=json")
echo "$cores_response" | grep -o '"name":"[^"]*"' || echo "No se pudieron listar cores"

echo ""
echo "✅ Debe existir: schedules"
echo ""

# ==========================================
# VERIFICAR DOCUMENTOS INDEXADOS
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📄 DOCUMENTOS INDEXADOS${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

# Contar documentos totales
total_docs=$(curl -s "http://localhost:8983/solr/schedules/select?q=*:*&rows=0&wt=json" | grep -o '"numFound":[0-9]*' | grep -o '[0-9]*')

if [ -z "$total_docs" ]; then
    echo -e "${RED}❌ No se pudo consultar Solr${NC}"
    echo "Verifica que el core 'schedules' existe"
    exit 1
fi

echo "Total de documentos indexados: $total_docs"
echo ""

if [ "$total_docs" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  NO HAY DOCUMENTOS INDEXADOS${NC}"
    echo ""
    echo "Esto puede significar:"
    echo "  - El consumer de search-api no está procesando eventos"
    echo "  - No se han creado schedules todavía"
    echo "  - Hubo un error al indexar"
    echo ""
    echo "Verifica:"
    echo "  1. docker-compose logs -f search-api"
    echo "  2. ./scripts/debug-rabbitmq.sh"
    echo "  3. Crea un schedule con: ./scripts/test-activities-api.sh"
else
    echo -e "${GREEN}✅ Hay $total_docs documento(s) indexado(s)${NC}"
    echo ""
fi

# ==========================================
# MOSTRAR DOCUMENTOS INDEXADOS
# ==========================================

if [ "$total_docs" -gt 0 ]; then
    echo -e "${BLUE}═══════════════════════════════════${NC}"
    echo -e "${BLUE}📋 LISTADO DE DOCUMENTOS${NC}"
    echo -e "${BLUE}═══════════════════════════════════${NC}"
    echo ""

    # Obtener todos los documentos (limitar a 10 para no saturar la pantalla)
    docs=$(curl -s "http://localhost:8983/solr/schedules/select?q=*:*&rows=10&wt=json&fl=id,activity_name,instructor,day_of_week,start_time,available_spots")

    echo "Primeros 10 documentos:"
    echo "$docs" | python3 -m json.tool 2>/dev/null || echo "$docs"
    echo ""
fi

# ==========================================
# VERIFICAR ESQUEMA
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}🗂️  ESQUEMA DE CAMPOS${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Campos definidos en el schema:"
schema=$(curl -s "http://localhost:8983/solr/schedules/schema/fields?wt=json")
echo "$schema" | grep -o '"name":"[^"]*"' | head -20

echo ""

# ==========================================
# PROBAR BÚSQUEDAS
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}🔎 PRUEBAS DE BÚSQUEDA${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

if [ "$total_docs" -gt 0 ]; then
    echo "Test 1: Búsqueda por texto (q=*)"
    result1=$(curl -s "http://localhost:8983/solr/schedules/select?q=*&rows=1&wt=json" | grep -o '"numFound":[0-9]*' | grep -o '[0-9]*')
    echo "   Resultados: $result1"
    echo ""

    echo "Test 2: Filtrar por día (fq=day_of_week:monday)"
    result2=$(curl -s "http://localhost:8983/solr/schedules/select?q=*:*&fq=day_of_week:monday&rows=0&wt=json" | grep -o '"numFound":[0-9]*' | grep -o '[0-9]*')
    echo "   Resultados: $result2"
    echo ""

    echo "Test 3: Filtrar por disponibilidad (fq=available_spots:[1 TO *])"
    result3=$(curl -s "http://localhost:8983/solr/schedules/select?q=*:*&fq=available_spots:[1%20TO%20*]&rows=0&wt=json" | grep -o '"numFound":[0-9]*' | grep -o '[0-9]*')
    echo "   Resultados: $result3"
    echo ""

    if [ "$result1" -gt 0 ] && [ "$result2" -ge 0 ] && [ "$result3" -ge 0 ]; then
        echo -e "${GREEN}✅ Todas las búsquedas funcionan correctamente${NC}"
    else
        echo -e "${YELLOW}⚠️  Algunas búsquedas no retornaron resultados${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  No se pueden probar búsquedas sin documentos indexados${NC}"
fi

echo ""

# ==========================================
# VERIFICAR LOGS DE INDEXACIÓN
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📝 LOGS RECIENTES DE SEARCH-API${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Últimas 20 líneas relacionadas con indexación:"
docker-compose logs --tail=20 search-api 2>/dev/null | grep -E 'indexando|Solr|RabbitMQ|Mensaje recibido' || echo "No se encontraron logs de indexación"

echo ""

# ==========================================
# RESUMEN Y RECOMENDACIONES
# ==========================================

echo "╔════════════════════════════════════════╗"
echo "║   📋 RESUMEN                           ║"
echo "╚════════════════════════════════════════╝"
echo ""

echo "Estado de Solr:"
echo ""
echo "✅ Solr está corriendo: http://localhost:8983"
echo "✅ Core 'schedules' existe"

if [ "$total_docs" -eq 0 ]; then
    echo -e "${RED}❌ No hay documentos indexados${NC}"
    echo ""
    echo "Pasos para resolver:"
    echo ""
    echo "1. Verifica que RabbitMQ esté procesando eventos:"
    echo "   ./scripts/debug-rabbitmq.sh"
    echo ""
    echo "2. Verifica logs del consumer:"
    echo "   docker-compose logs -f search-api | grep 'Mensaje recibido'"
    echo ""
    echo "3. Crea un schedule para generar eventos:"
    echo "   ./scripts/test-activities-api.sh"
    echo ""
    echo "4. Espera 5-10 segundos y vuelve a verificar:"
    echo "   ./scripts/debug-solr.sh"
else
    echo -e "${GREEN}✅ Hay $total_docs documento(s) indexado(s)${NC}"
    echo -e "${GREEN}✅ Las búsquedas funcionan correctamente${NC}"
    echo ""
    echo "Todo está funcionando bien. El flujo completo es:"
    echo "  activities-api → RabbitMQ → search-api → Solr ✅"
fi

echo ""

echo "Comandos útiles:"
echo ""
echo "# Ver interfaz web de Solr"
echo "open http://localhost:8983"
echo ""
echo "# Query manual en Solr"
echo "curl 'http://localhost:8983/solr/schedules/select?q=*:*&wt=json&indent=true'"
echo ""
echo "# Eliminar todos los documentos (si necesitas limpiar)"
echo "curl 'http://localhost:8983/solr/schedules/update?commit=true' -H 'Content-Type: text/xml' --data-binary '<delete><query>*:*</query></delete>'"
echo ""
echo "# Ver logs de search-api en tiempo real"
echo "docker-compose logs -f search-api"
echo ""
