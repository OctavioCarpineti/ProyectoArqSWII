#!/bin/bash

# Script de configuración inicial del proyecto

set -e

echo "🏋️  Gym Booking System - Setup Inicial"
echo "======================================"
echo ""

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Función para imprimir mensajes
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 1. Verificar prerequisitos
echo "1. Verificando prerequisitos..."

if ! command -v docker &> /dev/null; then
    print_error "Docker no está instalado"
    exit 1
fi
print_success "Docker instalado"

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose no está instalado"
    exit 1
fi
print_success "Docker Compose instalado"

echo ""

# 2. Crear estructura de directorios
echo "2. Creando estructura de directorios..."

mkdir -p backend/{users-api,activities-api,bookings-api,search-api}
mkdir -p frontend/src
mkdir -p database/{mysql,mongo}
mkdir -p config/rabbitmq
mkdir -p config/solr/schedules/conf
mkdir -p docs/api
mkdir -p scripts

print_success "Directorios creados"
echo ""

# 3. Verificar archivos de configuración
echo "3. Verificando archivos de configuración..."

files=(
    "docker-compose.yml"
    "database/mysql/init.sql"
    "database/mongo/init.js"
    "config/rabbitmq/definitions.json"
    "config/rabbitmq/rabbitmq.conf"
    "config/solr/schedules/conf/schema.xml"
    "config/solr/schedules/conf/solrconfig.xml"
    "config/solr/schedules/conf/stopwords.txt"
    "config/solr/schedules/core.properties"
)

missing_files=0
for file in "${files[@]}"; do
    if [ ! -f "$file" ]; then
        print_warning "Falta archivo: $file"
        missing_files=$((missing_files + 1))
    fi
done

if [ $missing_files -eq 0 ]; then
    print_success "Todos los archivos de configuración presentes"
else
    print_error "Faltan $missing_files archivo(s) de configuración"
    echo "Por favor, asegúrate de tener todos los archivos necesarios"
fi
echo ""

# 4. Crear .gitignore si no existe
if [ ! -f ".gitignore" ]; then
    echo "4. Creando .gitignore..."
    cat > .gitignore << 'EOF'
# Binarios de Go
*.exe
*.test
*.out
go.work

# IDEs
.idea/
.vscode/
*.swp
.DS_Store

# Dependencias
vendor/

# Variables de entorno
.env
.env.local

# Logs
*.log

# Node modules
/frontend/node_modules/
/frontend/build/

# Datos persistentes
/data/
EOF
    print_success ".gitignore creado"
else
    print_success ".gitignore ya existe"
fi
echo ""

# 5. Verificar puertos disponibles
echo "5. Verificando puertos disponibles..."

ports=(3000 8080 8081 8082 8083 3306 27017 5672 15672 8983 11211)
ports_in_use=0

for port in "${ports[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        print_warning "Puerto $port ya está en uso"
        ports_in_use=$((ports_in_use + 1))
    fi
done

if [ $ports_in_use -eq 0 ]; then
    print_success "Todos los puertos están disponibles"
else
    print_warning "$ports_in_use puerto(s) en uso. Puede haber conflictos al levantar servicios"
fi
echo ""

# 6. Resumen y próximos pasos
echo "======================================"
echo "✅ Setup completado"
echo ""
echo "📋 Próximos pasos:"
echo ""
echo "1. Implementar cada microservicio en backend/"
echo "   - users-api (puerto 8080)"
echo "   - activities-api (puerto 8081)"
echo "   - bookings-api (puerto 8082)"
echo "   - search-api (puerto 8083)"
echo ""
echo "2. Levantar infraestructura:"
echo "   $ make infra-up"
echo "   o"
echo "   $ docker-compose up -d mysql mongodb rabbitmq solr memcached"
echo ""
echo "3. Construir servicios:"
echo "   $ make build"
echo ""
echo "4. Levantar todos los servicios:"
echo "   $ make up"
echo ""
echo "5. Ver logs:"
echo "   $ make logs"
echo ""
echo "📚 Documentación completa en README.md"
echo ""