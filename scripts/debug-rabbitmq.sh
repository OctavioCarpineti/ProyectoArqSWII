#!/bin/bash

# Script para verificar RabbitMQ en detalle
# Útil para debugging del flujo de eventos

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo "╔════════════════════════════════════════╗"
echo "║   🐰 DEBUGGING RABBITMQ                ║"
echo "╚════════════════════════════════════════╝"
echo ""

# Verificar que RabbitMQ esté corriendo
if ! curl -s http://localhost:15672 > /dev/null 2>&1; then
    echo -e "${RED}❌ RabbitMQ no está respondiendo${NC}"
    echo "Ejecuta: docker-compose up -d rabbitmq"
    exit 1
fi

echo -e "${GREEN}✅ RabbitMQ está corriendo${NC}"
echo ""

# ==========================================
# VERIFICAR EXCHANGES
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📡 EXCHANGES${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Listando exchanges..."
docker-compose exec -T rabbitmq rabbitmqctl list_exchanges name type durable auto_delete | grep -E "schedules|^name"

echo ""
echo "✅ Debe existir: schedules_exchange (type: topic, durable: true)"
echo ""

# ==========================================
# VERIFICAR QUEUES
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📬 QUEUES${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Listando queues con mensajes..."
docker-compose exec -T rabbitmq rabbitmqctl list_queues name messages messages_ready messages_unacknowledged consumers

echo ""
echo "✅ Debe existir: schedules_queue"
echo "✅ Consumers debe ser >= 1 (search-api consumiendo)"
echo ""

# ==========================================
# VERIFICAR BINDINGS
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}🔗 BINDINGS${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Listando bindings..."
docker-compose exec -T rabbitmq rabbitmqctl list_bindings source_name destination_name routing_key | grep schedules

echo ""
echo "✅ Debe existir: schedules_exchange → schedules_queue con routing_key: schedule.*"
echo ""

# ==========================================
# VERIFICAR CONEXIONES Y CONSUMERS
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}🔌 CONSUMERS ACTIVOS${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

echo "Consumers conectados:"
docker-compose exec -T rabbitmq rabbitmqctl list_consumers queue_name consumer_tag | grep schedules

echo ""

# ==========================================
# VERIFICAR MENSAJES EN COLA
# ==========================================

echo -e "${BLUE}═══════════════════════════════════${NC}"
echo -e "${BLUE}📨 ESTADO DE MENSAJES${NC}"
echo -e "${BLUE}═══════════════════════════════════${NC}"
echo ""

queue_stats=$(docker-compose exec -T rabbitmq rabbitmqctl list_queues name messages messages_ready messages_unacknowledged | grep schedules_queue)

if [ -n "$queue_stats" ]; then
    messages=$(echo "$queue_stats" | awk '{print $2}')
    ready=$(echo "$queue_stats" | awk '{print $3}')
    unacked=$(echo "$queue_stats" | awk '{print $4}')

    echo "Queue: schedules_queue"
    echo "  Total mensajes: $messages"
    echo "  Listos para consumir: $ready"
    echo "  No confirmados: $unacked"
    echo ""

    if [ "$messages" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  HAY MENSAJES SIN PROCESAR${NC}"
        echo "Esto puede significar:"
        echo "  - El consumer de search-api no está corriendo"
        echo "  - El consumer está fallando al procesar"
        echo ""
        echo "Verifica logs: docker-compose logs -f search-api"
    else
        echo -e "${GREEN}✅ No hay mensajes pendientes (todos fueron procesados)${NC}"
    fi
else
    echo -e "${RED}❌ No se encontró la queue schedules_queue${NC}"
fi

echo ""

# ==========================================
# RESUMEN Y RECOMENDACIONES
# ==========================================

echo "╔════════════════════════════════════════╗"
echo "║   📋 RESUMEN                           ║"
echo "╚════════════════════════════════════════╝"
echo ""

echo "Para verificar que todo funciona:"
echo ""
echo "1. ✅ Exchange 'schedules_exchange' existe (tipo topic)"
echo "2. ✅ Queue 'schedules_queue' existe"
echo "3. ✅ Binding entre exchange y queue con routing_key 'schedule.*'"
echo "4. ✅ Consumer activo (search-api conectado)"
echo "5. ✅ Mensajes = 0 (todos procesados)"
echo ""

echo "Comandos útiles:"
echo ""
echo "# Ver logs del consumer (search-api)"
echo "docker-compose logs -f search-api | grep -E 'Mensaje recibido|RabbitMQ|indexando'"
echo ""
echo "# Ver interfaz web de RabbitMQ"
echo "open http://localhost:15672"
echo "Usuario: guest / Contraseña: guest"
echo ""
echo "# Purgar mensajes de la queue (si hay atascos)"
echo "docker-compose exec rabbitmq rabbitmqctl purge_queue schedules_queue"
echo ""
