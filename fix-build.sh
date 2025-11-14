#!/bin/bash

# Script de solución rápida para el error de compilación

echo "🔧 Solucionando error de compilación de activities-api..."
echo ""

# Paso 1: Obtener últimos cambios
echo "📥 Paso 1: Obteniendo últimos cambios del repositorio..."
git fetch origin
git pull origin claude/analyze-repository-01RtitfadSMWodbXvat4cPzx

# Paso 2: Verificar que el método SetPublisher existe
echo ""
echo "✅ Paso 2: Verificando que el método SetPublisher existe..."
if grep -q "func (s \*ScheduleServiceImpl) SetPublisher" backend/activities-api/services/schedule_service_impl.go; then
    echo "   ✅ Método SetPublisher encontrado!"
else
    echo "   ❌ ERROR: Método SetPublisher NO encontrado"
    echo "   Agregando el método manualmente..."

    # Agregar el método al final del archivo si no existe
    cat >> backend/activities-api/services/schedule_service_impl.go << 'EOF'

// SetPublisher permite inyectar el publisher (usado para testing)
func (s *ScheduleServiceImpl) SetPublisher(publisher *messaging.RabbitMQPublisher) {
	s.publisher = publisher
}
EOF
    echo "   ✅ Método agregado!"
fi

# Paso 3: Verificar compilación
echo ""
echo "🧪 Paso 3: Verificando que compila correctamente..."
cd backend/activities-api
if go build .; then
    echo "   ✅ Compilación exitosa!"
    cd ../..
else
    echo "   ❌ ERROR: Aún hay errores de compilación"
    cd ../..
    exit 1
fi

# Paso 4: Limpiar contenedores anteriores
echo ""
echo "🧹 Paso 4: Limpiando contenedores anteriores..."
docker-compose down -v 2>/dev/null || true

echo ""
echo "✅ ¡Listo! Ahora podés ejecutar:"
echo "   docker-compose up --build -d"
echo ""
