.PHONY: help build up down restart logs clean test

# Variables
DOCKER_COMPOSE = docker-compose
BACKEND_SERVICES = users-api activities-api bookings-api search-api

help: ## Mostrar ayuda
	@echo "Comandos disponibles:"
	@echo "  make build         - Construir todas las imágenes"
	@echo "  make up            - Levantar todos los servicios"
	@echo "  make down          - Detener todos los servicios"
	@echo "  make restart       - Reiniciar todos los servicios"
	@echo "  make logs          - Ver logs de todos los servicios"
	@echo "  make clean         - Limpiar volúmenes y containers"
	@echo "  make test          - Ejecutar tests"
	@echo "  make infra-up      - Levantar solo infraestructura (DB, RabbitMQ, etc)"
	@echo "  make services-up   - Levantar solo servicios backend"

build: ## Construir todas las imágenes Docker
	$(DOCKER_COMPOSE) build

up: ## Levantar todos los servicios
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Servicios levantados. Accesos:"
	@echo "   - Frontend: http://localhost:3000"
	@echo "   - Users API: http://localhost:8080"
	@echo "   - Activities API: http://localhost:8081"
	@echo "   - Bookings API: http://localhost:8082"
	@echo "   - Search API: http://localhost:8083"
	@echo "   - RabbitMQ UI: http://localhost:15672 (guest/guest)"
	@echo "   - Solr UI: http://localhost:8983"

down: ## Detener todos los servicios
	$(DOCKER_COMPOSE) down

restart: down up ## Reiniciar todos los servicios

logs: ## Ver logs de todos los servicios
	$(DOCKER_COMPOSE) logs -f

logs-users: ## Ver logs de users-api
	$(DOCKER_COMPOSE) logs -f users-api

logs-activities: ## Ver logs de activities-api
	$(DOCKER_COMPOSE) logs -f activities-api

logs-bookings: ## Ver logs de bookings-api
	$(DOCKER_COMPOSE) logs -f bookings-api

logs-search: ## Ver logs de search-api
	$(DOCKER_COMPOSE) logs -f search-api

clean: ## Limpiar containers, volúmenes e imágenes
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f

infra-up: ## Levantar solo infraestructura
	$(DOCKER_COMPOSE) up -d mysql mongodb rabbitmq solr memcached
	@echo "✅ Infraestructura levantada"

services-up: ## Levantar solo servicios backend
	$(DOCKER_COMPOSE) up -d $(BACKEND_SERVICES)

test: ## Ejecutar tests de todos los servicios
	@echo "Ejecutando tests de activities-api..."
	cd backend/activities-api && go test ./... -v
	@echo "✅ Tests completados"

ps: ## Ver estado de los servicios
	$(DOCKER_COMPOSE) ps

shell-mysql: ## Conectar a MySQL
	$(DOCKER_COMPOSE) exec mysql mysql -uroot -proot123 gym_users

shell-mongo: ## Conectar a MongoDB
	$(DOCKER_COMPOSE) exec mongodb mongosh -u admin -p admin123 gym_db

rebuild-service: ## Reconstruir un servicio específico (uso: make rebuild-service SERVICE=users-api)
	$(DOCKER_COMPOSE) build $(SERVICE)
	$(DOCKER_COMPOSE) up -d $(SERVICE)

health: ## Verificar salud de los servicios
	@echo "Verificando servicios..."
	@curl -f http://localhost:8080/health 2>/dev/null && echo "✅ Users API" || echo "❌ Users API"
	@curl -f http://localhost:8081/health 2>/dev/null && echo "✅ Activities API" || echo "❌ Activities API"
	@curl -f http://localhost:8082/health 2>/dev/null && echo "✅ Bookings API" || echo "❌ Bookings API"
	@curl -f http://localhost:8083/health 2>/dev/null && echo "✅ Search API" || echo "❌ Search API"