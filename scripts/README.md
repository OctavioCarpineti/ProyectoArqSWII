# 🧪 Scripts de Testing y Limpieza

## 📋 Scripts Disponibles

### 🧹 Scripts de Limpieza

#### `clean-all.sh` - Limpieza completa
```bash
./scripts/clean-all.sh
```

**¿Qué hace?**
- 🛑 Detiene todos los contenedores
- 🗑️ Elimina todos los volúmenes (datos de MySQL, MongoDB, RabbitMQ, Solr)
- 🧹 Limpia archivos temporales
- 🔧 Limpia sistema Docker

**¿Cuándo usarlo?**
- Cuando quieres empezar completamente de cero
- Si tienes datos corruptos o conflictos
- Antes de ejecutar tests para garantizar ambiente limpio

---

#### `clean-and-test.sh` - Limpia y testea (⭐ RECOMENDADO)
```bash
./scripts/clean-and-test.sh
```

**¿Qué hace?**
1. 🧹 Limpia todo el sistema
2. 🚀 Levanta todos los servicios
3. ⏳ Espera a que estén listos (con verificación de health)
4. 🧪 Ejecuta **todos** los tests E2E

**¿Cuándo usarlo?**
- ⭐ **Antes de presentar el proyecto**
- Cada vez que quieras verificar que TODO funciona
- Después de hacer cambios importantes
- Para garantizar ambiente limpio en cada test

**Duración:** ~2-3 minutos

---

### 🧪 Scripts de Testing Individual

#### `test-all.sh` - Ejecuta todos los tests
```bash
./scripts/test-all.sh
```

**Prerequisito:** Servicios deben estar corriendo (`docker-compose up -d`)

**¿Qué testea?**
1. users-api (JWT, login, CRUD)
2. activities-api (actividades, horarios, concurrencia)
3. search-api (Solr, caché, RabbitMQ consumer)
4. bookings-api (reservas, concurrencia, validaciones)

---

#### Tests individuales por API

**Solo si servicios YA están corriendo:**

```bash
# Test de users-api
./scripts/test-users-api.sh

# Test de activities-api (requiere haber ejecutado test-users-api.sh primero)
./scripts/test-activities-api.sh

# Test de search-api (requiere test-users-api.sh y test-activities-api.sh)
./scripts/test-search-api.sh

# Test de bookings-api (requiere los 3 anteriores)
./scripts/test-bookings-api.sh
```

**Nota:** Los tests individuales dependen entre sí porque comparten tokens y IDs guardados en `/tmp/gym-tokens.env`

---

## 🚀 Uso Rápido

### Caso 1: Quiero verificar que todo funciona (PRIMERA VEZ)

```bash
# Opción A: Todo en un solo comando (recomendado)
./scripts/clean-and-test.sh

# Opción B: Paso a paso
./scripts/clean-all.sh
docker-compose up -d
sleep 15
./scripts/test-all.sh
```

### Caso 2: Ya tengo servicios corriendo, solo quiero testear

```bash
./scripts/test-all.sh
```

### Caso 3: Quiero empezar de cero (hay datos viejos o errores)

```bash
./scripts/clean-all.sh
docker-compose up -d
sleep 15
```

### Caso 4: Tengo un error y quiero limpiar + verificar

```bash
./scripts/clean-and-test.sh
```

---

## 📊 ¿Qué verifica cada test?

### ✅ test-users-api.sh
- Health check
- Crear usuario normal
- Login con JWT
- Login admin
- Obtener usuario por ID
- Validación de credenciales incorrectas
- Validación de duplicados

### ✅ test-activities-api.sh
- Health check
- Crear actividades (requiere admin token)
- Listar actividades
- **Crear horarios CON CONCURRENCIA** (3 goroutines)
  - Validar actividad existe
  - Detectar conflictos de horario/sala
  - Validar formato de tiempo
- **Publicar eventos a RabbitMQ**

### ✅ test-search-api.sh
- Health check
- **Verificar consumo de RabbitMQ**
- **Verificar indexación en Solr**
- Búsqueda sin filtros
- Búsqueda por texto
- Búsqueda por categoría
- Búsqueda por día de semana
- Búsqueda con múltiples filtros
- **Verificar caché multinivel (L1 + L2)**

### ✅ test-bookings-api.sh
- Health check
- **Crear reserva CON CONCURRENCIA** (3 goroutines)
  - Validar usuario existe (HTTP a users-api)
  - Validar schedule existe y tiene cupos (HTTP a activities-api)
  - Detectar duplicados (MongoDB)
- Listar mis reservas
- **Verificar incremento de current_bookings**
- Cancelar reserva (soft delete)
- **Verificar decremento de current_bookings**
- Validación de duplicados
- Validación de autenticación

---

## 🔍 Verificaciones Adicionales

### Ver logs de servicios
```bash
# Todos
docker-compose logs -f

# Solo uno
docker-compose logs -f activities-api
docker-compose logs -f search-api
```

### Verificar RabbitMQ
```bash
# Web UI
open http://localhost:15672
# Usuario: guest / Contraseña: guest

# O por terminal
docker-compose exec rabbitmq rabbitmqctl list_exchanges
docker-compose exec rabbitmq rabbitmqctl list_queues
```

### Verificar Solr
```bash
# Web UI
open http://localhost:8983

# Directamente con curl
curl "http://localhost:8983/solr/schedules/select?q=*:*&wt=json"
```

### Ver tokens y variables generadas
```bash
cat /tmp/gym-tokens.env
```

---

## 🔍 Scripts de Debugging

### `debug-rabbitmq.sh` - Verificar RabbitMQ en detalle
```bash
./scripts/debug-rabbitmq.sh
```

**¿Qué verifica?**
- ✅ RabbitMQ está corriendo
- 📡 Exchanges existentes (debe existir `schedules_exchange`)
- 📬 Queues con mensajes (debe existir `schedules_queue`)
- 🔗 Bindings entre exchange y queue (routing_key: `schedule.*`)
- 🔌 Consumers activos (search-api debe estar conectado)
- 📨 Estado de mensajes (listos, no confirmados)

**¿Cuándo usarlo?**
- Cuando `test-search-api.sh` falla con Solr 404
- Para verificar que el consumer está procesando eventos
- Para debugging del flujo asíncrono: activities-api → RabbitMQ → search-api

---

### `debug-solr.sh` - Verificar Solr e indexación
```bash
./scripts/debug-solr.sh
```

**¿Qué verifica?**
- ✅ Solr está corriendo (http://localhost:8983)
- 📚 Core `schedules` existe
- 📄 Cantidad de documentos indexados
- 📋 Listado de documentos (primeros 10)
- 🗂️ Esquema de campos definidos
- 🔎 Pruebas de búsqueda (por texto, día, disponibilidad)
- 📝 Logs recientes de search-api

**¿Cuándo usarlo?**
- Cuando search-api no retorna resultados
- Para verificar que los schedules se indexaron correctamente
- Para debugging del flujo: RabbitMQ → search-api → Solr

---

## 🐛 Troubleshooting

### Error: "Connection refused"
```bash
# Verificar que servicios estén corriendo
docker-compose ps

# Si no están, levantarlos
docker-compose up -d

# Esperar 10-15 segundos
sleep 15
```

### Error: "schedule conflict"
- Esto significa que ya hay un horario en esa sala/día/hora
- Es comportamiento correcto (detección de conflictos funciona)
- Solución: Ejecutar `clean-and-test.sh` para empezar limpio

### Error: "Solr no encuentra resultados"
- El consumer de RabbitMQ necesita tiempo para indexar
- Esperar 5-10 segundos después de crear horarios
- Verificar logs: `docker-compose logs -f search-api`

### Error: "401 Unauthorized"
- Token JWT inválido o expirado
- Ejecutar `test-users-api.sh` primero para generar nuevos tokens

---

## 📁 Archivos Generados

### `/tmp/gym-tokens.env`
Contiene todas las variables generadas por los tests:
- `USER_TOKEN` - Token JWT del usuario
- `ADMIN_TOKEN` - Token JWT del admin
- `USER_ID` - ID del usuario creado
- `ADMIN_USER_ID` - ID del admin (siempre 1)
- `ACTIVITY_ID` - ID de la actividad creada
- `SCHEDULE_ID` - ID del horario creado
- `BOOKING_ID` - ID de la reserva creada

---

## ✅ Checklist antes de Entregar

```bash
# 1. Limpiar y testear todo
./scripts/clean-and-test.sh

# 2. Verificar que TODO pase ✅
# Si todos los tests pasan → Backend está perfecto

# 3. Verificar manualmente servicios
open http://localhost:15672  # RabbitMQ
open http://localhost:8983   # Solr

# 4. Ver logs para asegurar que no hay errores
docker-compose logs | grep -i error
```

Si `clean-and-test.sh` pasa completamente ✅ → Tu backend está **PERFECTO** y listo para el frontend.

---

## 💡 Tips

- ⭐ Ejecuta `clean-and-test.sh` antes de presentar
- ⚡ Usa `test-all.sh` para tests rápidos (si servicios ya corren)
- 🧹 Usa `clean-all.sh` si tienes problemas con datos
- 📝 Los tests guardan tokens en `/tmp/gym-tokens.env` para reutilizar
- ⏱️ Espera 10-15 segundos después de `docker-compose up -d`
