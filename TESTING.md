# 🧪 Guía de Testing - Gym Booking System

## 📋 Índice
1. [Prerequisitos](#prerequisitos)
2. [Opción 1: Scripts Automáticos](#opción-1-scripts-automáticos)
3. [Opción 2: Postman](#opción-2-postman)
4. [Opción 3: Testing Manual con curl](#opción-3-testing-manual-con-curl)
5. [Verificación de RabbitMQ](#verificación-de-rabbitmq)
6. [Verificación de Solr](#verificación-de-solr)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisitos

### 1. Levantar todos los servicios

```bash
# Desde la raíz del proyecto
docker-compose up -d

# Verificar que todos estén corriendo
docker-compose ps

# Ver logs en tiempo real
docker-compose logs -f
```

### 2. Verificar health checks

```bash
# Verificar todos los servicios
make health

# O manualmente:
curl http://localhost:8080/health  # users-api
curl http://localhost:8081/health  # activities-api
curl http://localhost:8082/health  # bookings-api
curl http://localhost:8083/health  # search-api
```

---

## Opción 1: Scripts Automáticos

### ⭐ Limpia y testea todo (RECOMENDADO)

```bash
cd scripts

# Ejecutar limpieza completa + tests
./clean-and-test.sh
```

Este script hace TODO automáticamente:
1. 🧹 Limpia contenedores, volúmenes y datos
2. 🚀 Levanta todos los servicios
3. ⏳ Espera a que estén listos
4. 🧪 Ejecuta todos los tests E2E

**Duración:** ~2-3 minutos
**Uso:** Ejecutar antes de presentar o cuando quieras verificar que todo funciona

---

### Solo limpiar (sin testear)

```bash
./scripts/clean-all.sh
```

Esto elimina:
- Todos los contenedores
- Todos los volúmenes (MySQL, MongoDB, RabbitMQ, Solr)
- Cache y datos temporales

---

### Ejecutar tests (sin limpiar)

```bash
cd scripts

# Prerequisito: servicios deben estar corriendo
docker-compose up -d
sleep 15

# Ejecutar test completo
./test-all.sh
```

Esto ejecutará:
1. ✅ Test de users-api (creación de usuarios, login, JWT)
2. ✅ Test de activities-api (actividades, horarios, concurrencia, RabbitMQ)
3. ✅ Test de search-api (búsquedas, filtros, caché)
4. ✅ Test de bookings-api (reservas, concurrencia, validaciones)

### Ejecutar tests individuales

```bash
# Solo users-api
./test-users-api.sh

# Solo activities-api (requiere haber ejecutado test-users-api.sh primero)
./test-activities-api.sh

# Solo search-api (requiere test-users-api.sh y test-activities-api.sh)
./test-search-api.sh

# Solo bookings-api (requiere los 3 anteriores)
./test-bookings-api.sh
```

### Ver tokens y variables guardadas

```bash
cat /tmp/gym-tokens.env
```

---

## Opción 2: Postman

### Importar colección

1. Abrir Postman
2. Click en **Import**
3. Seleccionar el archivo `Gym_Booking_System.postman_collection.json`
4. La colección se importará con todas las requests organizadas

### Ejecutar requests en orden

La colección está organizada en 4 carpetas:

#### 1. users-api
- **Health Check** - Verificar servicio
- **Create User** - Crea usuario y guarda USER_ID
- **Login User** - Login y guarda USER_TOKEN
- **Login Admin** - Login admin y guarda ADMIN_TOKEN
- **Get User By ID** - Obtiene usuario

#### 2. activities-api
- **Health Check**
- **Create Activity** - Crea actividad (requiere ADMIN_TOKEN) y guarda ACTIVITY_ID
- **List All Activities**
- **Get Activity By ID**
- **Create Schedule** - Crea horario y guarda SCHEDULE_ID (⚡ ejecuta concurrencia)
- **List All Schedules**
- **Get Schedule By ID**

#### 3. search-api
- **Health Check**
- **Search - All** - Búsqueda sin filtros
- **Search - By Text** - Búsqueda por "yoga"
- **Search - By Category** - Filtro por categoría
- **Search - By Day** - Filtro por día de semana
- **Search - Available Only** - Solo horarios disponibles
- **Search - Combined Filters** - Múltiples filtros

#### 4. bookings-api
- **Health Check**
- **Create Booking** - Crear reserva (⚡ ejecuta concurrencia) y guarda BOOKING_ID
- **Get Booking By ID**
- **List My Bookings**
- **Cancel Booking** - Soft delete

### Variables de colección

La colección usa variables que se auto-completan:
- `USER_TOKEN` - Token JWT del usuario
- `ADMIN_TOKEN` - Token JWT del admin
- `USER_ID` - ID del usuario creado
- `ACTIVITY_ID` - ID de la actividad creada
- `SCHEDULE_ID` - ID del horario creado
- `BOOKING_ID` - ID de la reserva creada

---

## Opción 3: Testing Manual con curl

### FASE 1: users-api

#### 1.1 Crear usuario normal

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "Password123!",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Respuesta esperada** (HTTP 201):
```json
{
  "id": 3,
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "normal"
}
```

Guardar el `id` para uso posterior.

#### 1.2 Login con usuario normal

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "johndoe",
    "password": "Password123!"
  }'
```

**Respuesta esperada** (HTTP 200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 3,
    "username": "johndoe",
    ...
  }
}
```

**IMPORTANTE**: Copiar el `token` completo.

#### 1.3 Login con admin (pre-creado)

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username_or_email": "admin",
    "password": "admin123"
  }'
```

Copiar el `token` del admin.

---

### FASE 2: activities-api

#### 2.1 Crear actividad (como admin)

```bash
curl -X POST http://localhost:8081/activities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TU_ADMIN_TOKEN_AQUI" \
  -d '{
    "owner_id": 1,
    "name": "Yoga Intermedio",
    "description": "Clase de yoga nivel intermedio",
    "category": "Yoga",
    "duration": 60,
    "price": 1500.00,
    "image_url": "https://example.com/yoga.jpg"
  }'
```

**Respuesta esperada** (HTTP 201):
```json
{
  "id": "673abc123def456...",
  "owner_id": 1,
  "name": "Yoga Intermedio",
  ...
}
```

Guardar el `id` de la actividad.

#### 2.2 Crear horario (⚡ ejecuta concurrencia)

```bash
curl -X POST http://localhost:8081/activities/ACTIVITY_ID_AQUI/schedules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TU_ADMIN_TOKEN_AQUI" \
  -d '{
    "instructor": "María González",
    "day_of_week": "monday",
    "start_time": "18:00",
    "end_time": "19:00",
    "location": "Sala 1 - Sede Centro",
    "max_capacity": 20
  }'
```

**Este request ejecuta 3 goroutines en paralelo**:
1. Valida que la actividad existe
2. Verifica que no hay conflicto de horario/sala
3. Valida formato de tiempo

**Respuesta esperada** (HTTP 201):
```json
{
  "id": "673def789abc123...",
  "activity_id": "673abc123def456...",
  "instructor": "María González",
  "day_of_week": "monday",
  ...
  "current_bookings": 0,
  "available_spots": 20
}
```

Guardar el `id` del schedule.

**⏳ IMPORTANTE**: Este request también publica un evento a RabbitMQ. Espera 3-5 segundos para que search-api lo consuma e indexe en Solr.

#### 2.3 Listar todos los horarios

```bash
curl http://localhost:8081/schedules
```

---

### FASE 3: search-api

**⏳ Espera 5 segundos** después de crear horarios para que RabbitMQ consumer procese.

#### 3.1 Búsqueda sin filtros

```bash
curl http://localhost:8083/search
```

#### 3.2 Búsqueda por texto

```bash
curl "http://localhost:8083/search?q=yoga"
```

#### 3.3 Búsqueda por categoría

```bash
curl "http://localhost:8083/search?category=Yoga"
```

#### 3.4 Búsqueda por día

```bash
curl "http://localhost:8083/search?day_of_week=monday"
```

#### 3.5 Solo horarios disponibles

```bash
curl "http://localhost:8083/search?available=true"
```

#### 3.6 Búsqueda combinada

```bash
curl "http://localhost:8083/search?q=yoga&day_of_week=monday&available=true&page=1&size=10"
```

---

### FASE 4: bookings-api

#### 4.1 Crear reserva (⚡ ejecuta concurrencia)

```bash
curl -X POST http://localhost:8082/bookings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TU_USER_TOKEN_AQUI" \
  -d '{
    "user_id": 3,
    "schedule_id": "SCHEDULE_ID_AQUI"
  }'
```

**Este request ejecuta 3 goroutines en paralelo**:
1. Valida que el usuario existe (HTTP GET a users-api)
2. Valida que el schedule existe y tiene cupos (HTTP GET a activities-api)
3. Verifica que no hay reserva duplicada (MongoDB query)

**Respuesta esperada** (HTTP 201):
```json
{
  "id": "673xyz456abc789...",
  "user_id": 3,
  "schedule_id": "673def789abc123...",
  "activity_name": "Yoga Intermedio",
  "instructor": "María González",
  ...
  "status": "confirmed"
}
```

Guardar el `id` de la reserva.

#### 4.2 Verificar que current_bookings se incrementó

```bash
curl http://localhost:8081/schedules/SCHEDULE_ID_AQUI
```

Verifica que `current_bookings` ahora es 1.

#### 4.3 Listar mis reservas

```bash
curl -H "Authorization: Bearer TU_USER_TOKEN_AQUI" \
  "http://localhost:8082/bookings?user_id=3"
```

#### 4.4 Cancelar reserva

```bash
curl -X DELETE http://localhost:8082/bookings/BOOKING_ID_AQUI \
  -H "Authorization: Bearer TU_USER_TOKEN_AQUI"
```

#### 4.5 Verificar que current_bookings se decrementó

```bash
curl http://localhost:8081/schedules/SCHEDULE_ID_AQUI
```

Verifica que `current_bookings` volvió a 0.

---

## Verificación de RabbitMQ

### ⭐ Opción 1: Script de debugging (RECOMENDADO)

```bash
./scripts/debug-rabbitmq.sh
```

Este script automáticamente verifica:
- ✅ RabbitMQ está corriendo
- 📡 Exchange `schedules_exchange` existe (tipo: topic)
- 📬 Queue `schedules_queue` existe
- 🔗 Bindings correctos (routing_key: `schedule.*`)
- 🔌 Consumers activos (search-api conectado)
- 📨 Estado de mensajes (cuántos están pendientes, listos, no confirmados)

**Uso:** Ejecuta este script cuando veas errores 404 en Solr o cuando search-api no retorne resultados.

---

### Opción 2: Interfaz web manual

1. Abrir: http://localhost:15672
2. Usuario: `guest`
3. Contraseña: `guest`

### Verificar exchange y queue

1. Ir a tab **Exchanges**
2. Buscar `schedules_exchange` (tipo: topic)
3. Ir a tab **Queues**
4. Buscar `schedules_queue`
5. Verificar que hay bindings y mensajes procesados

### Ver mensajes (si hay en cola)

1. Click en `schedules_queue`
2. Ir a **Get messages**
3. Click en **Get Message(s)**

---

## Verificación de Solr

### ⭐ Opción 1: Script de debugging (RECOMENDADO)

```bash
./scripts/debug-solr.sh
```

Este script automáticamente verifica:
- ✅ Solr está corriendo (http://localhost:8983)
- 📚 Core `schedules` existe
- 📄 Cantidad de documentos indexados
- 📋 Listado de primeros 10 documentos
- 🗂️ Esquema de campos
- 🔎 Pruebas de búsqueda (texto, día, disponibilidad)
- 📝 Logs recientes de indexación

**Uso:** Ejecuta este script cuando search-api no retorne resultados o para verificar que los schedules se indexaron correctamente.

---

### Opción 2: Interfaz web manual

1. Abrir: http://localhost:8983
2. Click en **Core Admin** o **Core Selector**
3. Seleccionar core `schedules`

### Verificar documentos indexados

1. En el menú izquierdo, seleccionar **Query**
2. Dejar `q` como `*:*`
3. Click en **Execute Query**
4. Deberías ver todos los schedules indexados

### Probar búsquedas

- Buscar por texto: `q=yoga`
- Filtrar por día: `fq=day_of_week:monday`
- Filtrar disponibles: `fq=available_spots:[1 TO *]`

---

## Troubleshooting

### Problema: "Connection refused" en alguna API

**Solución**:
```bash
# Verificar que los servicios estén corriendo
docker-compose ps

# Si alguno no está, levantarlo
docker-compose up -d

# Ver logs del servicio con problemas
docker-compose logs -f users-api
```

### Problema: search-api no encuentra resultados (Test 10 falla con 404)

**Causa**: El consumer de RabbitMQ aún no procesó los eventos y no indexó en Solr.

**Este es el flujo asíncrono**:
1. activities-api crea schedule → publica evento a RabbitMQ
2. RabbitMQ encola el mensaje en `schedules_queue`
3. search-api consumer recibe el mensaje
4. search-api indexa en Solr
5. Ahora las búsquedas funcionan

**Solución - Opción 1 (RECOMENDADA): Usar scripts de debugging**
```bash
# Ver estado completo de RabbitMQ
./scripts/debug-rabbitmq.sh

# Ver estado completo de Solr
./scripts/debug-solr.sh
```

**Solución - Opción 2: Manual**
```bash
# Ver logs del consumer
docker-compose logs -f search-api | grep -E "Mensaje recibido|indexando|Solr"

# Esperar 5-10 segundos después de crear schedules
sleep 10

# Reintentar búsqueda
curl http://localhost:8083/search

# Verificar manualmente en Solr
curl "http://localhost:8983/solr/schedules/select?q=*:*&wt=json&indent=true"
```

**Si el consumer NO está procesando:**
- Verificar que search-api está corriendo: `docker-compose ps search-api`
- Ver logs de errores: `docker-compose logs search-api | grep -i error`
- Reiniciar el servicio: `docker-compose restart search-api`

### Problema: Error 401 en requests autenticados

**Causa**: Token JWT inválido o expirado.

**Solución**:
1. Hacer login nuevamente
2. Copiar el token completo (incluido prefijo "Bearer")
3. Verificar que no haya espacios extras

### Problema: MongoDB connection timeout

**Solución**:
```bash
# Verificar que MongoDB está corriendo
docker-compose ps mongodb

# Ver logs
docker-compose logs mongodb

# Si no está healthy, reiniciar
docker-compose restart mongodb
```

### Problema: Solr no indexa documentos

**Solución**:
```bash
# Verificar logs de search-api
docker-compose logs -f search-api

# Verificar que RabbitMQ está corriendo
docker-compose ps rabbitmq

# Verificar que el consumer está consumiendo
# Debería ver logs tipo: "📨 Mensaje recibido..."

# Reiniciar search-api si es necesario
docker-compose restart search-api
```

### Problema: Error "schedule conflict" al crear horario

**Causa**: Ya existe un horario en esa sala, día y rango horario.

**Solución**:
- Cambiar `location` (usar otra sala)
- Cambiar `day_of_week`
- Cambiar rango de `start_time` y `end_time`

---

## Comandos Útiles

```bash
# Ver logs de todos los servicios
docker-compose logs -f

# Ver logs de un servicio específico
docker-compose logs -f users-api

# Reiniciar un servicio
docker-compose restart activities-api

# Reiniciar todo
docker-compose restart

# Limpiar y empezar de cero
docker-compose down -v
docker-compose up -d

# Ver estado de servicios
docker-compose ps

# Conectar a MySQL
docker-compose exec mysql mysql -uroot -proot123 gym_users

# Conectar a MongoDB
docker-compose exec mongodb mongosh -u admin -p admin123 gym_db

# Ver exchanges de RabbitMQ
docker-compose exec rabbitmq rabbitmqctl list_exchanges

# Ver queues de RabbitMQ
docker-compose exec rabbitmq rabbitmqctl list_queues
```

---

## Checklist de Verificación Final

Antes de pasar al frontend, verifica que TODO funcione:

- [ ] ✅ users-api health check OK
- [ ] ✅ Crear usuario normal funciona
- [ ] ✅ Login retorna JWT válido
- [ ] ✅ Login admin funciona
- [ ] ✅ activities-api health check OK
- [ ] ✅ Crear actividad funciona (requiere admin token)
- [ ] ✅ Crear horario funciona (ejecuta concurrencia)
- [ ] ✅ RabbitMQ recibe eventos (ver web UI)
- [ ] ✅ search-api health check OK
- [ ] ✅ search-api consume eventos de RabbitMQ
- [ ] ✅ Solr indexa documentos correctamente
- [ ] ✅ Búsquedas funcionan con filtros
- [ ] ✅ Caché funciona (segunda búsqueda más rápida)
- [ ] ✅ bookings-api health check OK
- [ ] ✅ Crear reserva funciona (ejecuta concurrencia)
- [ ] ✅ current_bookings se incrementa correctamente
- [ ] ✅ Listar reservas funciona
- [ ] ✅ Cancelar reserva funciona (soft delete)
- [ ] ✅ current_bookings se decrementa correctamente

**Si TODO está ✅, el backend está PERFECTO y listo para el frontend** 🎉
