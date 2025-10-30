# 🏗️ Arquitectura del Sistema

## Visión General

Sistema de reservas de gimnasio basado en **microservicios** con arquitectura orientada a eventos.

```
┌─────────────┐
│   Frontend  │
│   (React)   │
└──────┬──────┘
       │ HTTP/JSON
       ├──────────────┬──────────────┬──────────────┐
       │              │              │              │
   ┌───▼────┐    ┌───▼────┐    ┌───▼────┐    ┌───▼────┐
   │ Users  │    │Activities│   │Bookings│   │ Search │
   │  API   │    │   API   │    │  API   │    │  API   │
   └───┬────┘    └────┬─────┘   └───┬────┘    └───┬────┘
       │              │              │              │
   ┌───▼────┐    ┌───▼────┐    ┌───▼────┐    ┌───▼────┐
   │ MySQL  │    │ MongoDB│    │ MongoDB│    │  Solr  │
   └────────┘    └────┬───┘    └────────┘    │+Cache  │
                      │                       └────────┘
                 ┌────▼─────┐
                 │ RabbitMQ │
                 └──────────┘
```

---

## 📦 Microservicios

### 1. users-api (Puerto 8080)

**Responsabilidad**: Gestión de usuarios y autenticación

**Tecnologías**:
- Go + Gin
- MySQL + GORM
- JWT (autenticación)
- Bcrypt (hashing passwords)

**Endpoints**:
```
POST   /users      → Crear usuario
GET    /users/:id  → Obtener usuario
POST   /login      → Autenticar y obtener JWT
```

**Base de datos**: `gym_users` (MySQL)
- Tabla `users`: id, username, email, password, first_name, last_name, role

**Roles**:
- `normal`: Usuario estándar que puede reservar clases
- `admin`: Administrador con permisos de gestión

---

### 2. activities-api (Puerto 8081) ⭐ ENTIDAD PRINCIPAL

**Responsabilidad**: Gestión del catálogo de actividades y horarios

**Tecnologías**:
- Go + Gin
- MongoDB
- RabbitMQ (publicador)
- HTTP Client (comunicación con users-api)

**Endpoints de Activities**:
```
POST   /activities          → Crear actividad (ADMIN)
GET    /activities          → Listar todas
GET    /activities/:id      → Obtener una actividad
PUT    /activities/:id      → Actualizar (ADMIN/owner)
DELETE /activities/:id      → Eliminar (ADMIN/owner)
```

**Endpoints de Schedules** (nested bajo activities):
```
POST   /activities/:activity_id/schedules              → Crear horario
GET    /activities/:activity_id/schedules              → Listar horarios de una actividad
GET    /activities/:activity_id/schedules/:schedule_id → Obtener horario específico
PUT    /activities/:activity_id/schedules/:schedule_id → Actualizar horario
DELETE /activities/:activity_id/schedules/:schedule_id → Eliminar horario

GET    /schedules           → Listar TODOS los horarios (útil para búsqueda)
GET    /schedules/:id       → Obtener horario por ID (para bookings-api)
```

**Base de datos**: `gym_db` (MongoDB)
- Colección `activities`
- Colección `schedules`

**Eventos publicados** (RabbitMQ):
```json
{
  "operation": "CREATE|UPDATE|DELETE",
  "schedule_id": "...",
  "activity_id": "...",
  "timestamp": 1234567890
}
```

**Concurrencia**: Implementada en `CreateSchedule()` con goroutines + channels + WaitGroup

---

### 3. bookings-api (Puerto 8082)

**Responsabilidad**: Gestión de reservas de usuarios

**Tecnologías**:
- Go + Gin
- MongoDB
- HTTP Clients (users-api, activities-api)

**Endpoints**:
```
POST   /bookings           → Crear reserva (JWT requerido)
GET    /bookings/:id       → Obtener reserva
GET    /bookings           → Listar reservas (filtrable por user_id)
DELETE /bookings/:id       → Cancelar reserva
```

**Base de datos**: `gym_db` (MongoDB)
- Colección `bookings`

**Validaciones concurrentes** en `CreateBooking()`:
1. Usuario existe (HTTP GET a users-api)
2. Schedule existe y tiene cupos (HTTP GET a activities-api)
3. Usuario no tiene reserva duplicada

**Actualizaciones**:
- Al crear: incrementa `schedule.current_bookings`
- Al cancelar: decrementa `schedule.current_bookings`

---

### 4. search-api (Puerto 8083)

**Responsabilidad**: Búsqueda avanzada de horarios disponibles

**Tecnologías**:
- Go + Gin
- Apache Solr (motor de búsqueda)
- CCache (caché local en memoria)
- Memcached (caché distribuida)
- RabbitMQ (consumidor)
- HTTP Client (activities-api)

**Endpoint**:
```
GET /search?q=yoga&category=Yoga&day_of_week=monday&available=true&page=1&size=10
```

**Parámetros de búsqueda**:
- `q`: Texto libre (busca en nombre, instructor, categoría)
- `activity_id`: Filtrar por actividad específica
- `category`: Filtrar por categoría
- `day_of_week`: monday, tuesday, etc.
- `instructor`: Nombre del instructor
- `location`: Ubicación de la clase
- `time_from` / `time_to`: Rango horario
- `available`: Solo clases con cupos
- `sort`: Campo para ordenar (start_time, price, etc.)
- `page`, `size`: Paginación

**Estrategia de caché multinivel**:
```
1. Genera cache key basado en query params
2. Busca en CCache (L1 - local) → si encuentra, retorna
3. Busca en Memcached (L2 - distribuida) → si encuentra, retorna + guarda en L1
4. Busca en Solr → retorna + guarda en L2 y L1
```

**Consumidor RabbitMQ**:
- Escucha eventos de schedules (CREATE, UPDATE, DELETE)
- Obtiene schedule completo desde activities-api
- Indexa/actualiza/elimina en Solr
- Invalida caché relacionada

**Índice Solr**: `schedules`
- Campos indexados: id, activity_name, instructor, day_of_week, location, etc.

---

## 🔄 Flujos de Comunicación

### Flujo 1: Admin crea actividad con horarios

```
1. Frontend → POST /activities (activities-api)
   Headers: { Authorization: "Bearer JWT_ADMIN" }
   Body: { name, description, category, duration, price }

2. activities-api:
   - Valida JWT (middleware)
   - Valida owner_id existe → HTTP GET users-api/users/:id
   - Crea Activity en MongoDB
   - Devuelve activity_id

3. Frontend → POST /activities/{activity_id}/schedules
   Body: { instructor, day_of_week, start_time, end_time, location, max_capacity }

4. activities-api:
   - Valida concurrentemente (goroutines):
     * Activity existe
     * No hay conflicto horario/sala
     * Instructor disponible
   - Crea Schedule en MongoDB
   - Publica evento → RabbitMQ: { operation: "CREATE", schedule_id, activity_id }

5. search-api (consumidor en background):
   - Consume evento de RabbitMQ
   - HTTP GET activities-api/schedules/{schedule_id} → obtiene datos completos
   - HTTP GET activities-api/activities/{activity_id} → obtiene datos de activity
   - Combina info y crea documento ScheduleSearch
   - Indexa en Solr
```

### Flujo 2: Usuario busca clases

```
1. Frontend → GET /search?q=yoga&day_of_week=monday (search-api)

2. search-api:
   - Genera cache key: "search:yoga:monday:..."
   - Busca en CCache (L1) → ❌ miss
   - Busca en Memcached (L2) → ❌ miss
   - Consulta Solr con filtros
   - Obtiene resultados paginados
   - Guarda en Memcached (TTL 10 min)
   - Guarda en CCache (TTL 5 min)
   - Devuelve resultados al Frontend

3. Frontend muestra lista de clases:
   - Yoga Intermedio - Lunes 18:00-19:00 - Instructor: María - Sala 1 (15/20 cupos)
   - Yoga Avanzado - Lunes 19:00-20:00 - Instructor: Juan - Sala 2 (8/15 cupos)
```

### Flujo 3: Usuario reserva un horario

```
1. Usuario hace click en "Reservar" en una clase específica
   Frontend → POST /bookings (bookings-api)
   Headers: { Authorization: "Bearer JWT_USER" }
   Body: { user_id: 5, schedule_id: "507fabc..." }

2. bookings-api valida JWT y extrae user_id

3. bookings-api ejecuta validaciones CONCURRENTES (3 goroutines):
   
   Goroutine 1:
   → HTTP GET users-api/users/5
   → Verifica que usuario existe
   → Envía resultado a channel
   
   Goroutine 2:
   → HTTP GET activities-api/schedules/507fabc
   → Verifica que schedule existe
   → Verifica que current_bookings < max_capacity
   → Envía resultado a channel
   
   Goroutine 3:
   → Query MongoDB: bookings where user_id=5 AND schedule_id="507fabc"
   → Verifica que usuario no tenga reserva duplicada
   → Envía resultado a channel
   
   WaitGroup espera que terminen las 3 goroutines
   Lee errores del channel

4. Si todas las validaciones pasan:
   - Crea Booking en MongoDB con datos desnormalizados
   - HTTP PUT activities-api/schedules/507fabc
     Body: { current_bookings: current_bookings + 1 }
   
5. activities-api:
   - Actualiza Schedule en MongoDB
   - Publica evento → RabbitMQ: { operation: "UPDATE", schedule_id, activity_id }

6. search-api (consumidor):
   - Consume evento UPDATE
   - Actualiza documento en Solr (nuevo current_bookings)
   - Invalida caché para ese schedule

7. bookings-api devuelve respuesta exitosa
   Frontend muestra: "¡Reserva confirmada! Te esperamos el Lunes a las 18:00"
```

### Flujo 4: Usuario ve sus reservas

```
1. Frontend → GET /bookings?user_id=5 (bookings-api)
   Headers: { Authorization: "Bearer JWT_USER" }

2. bookings-api:
   - Valida JWT
   - Verifica que user_id del JWT coincide con user_id del query (o es admin)
   - Query MongoDB: db.bookings.find({ user_id: 5, status: "confirmed" })
   - Devuelve lista de reservas con toda la info desnormalizada

3. Frontend muestra "Mis Reservas":
   - Yoga Intermedio - Lunes 18:00 - Instructor: María - Sala 1
   - Spinning - Miércoles 19:00 - Instructor: Carlos - Sala 2
```

---

## 🔐 Seguridad

### JWT (JSON Web Tokens)

**Estructura del token**:
```json
{
  "user_id": 5,
  "username": "john_doe",
  "email": "john@example.com",
  "role": "normal",
  "exp": 1699999999
}
```

**Flujo de autenticación**:
1. Usuario envía credentials → `POST /login`
2. users-api valida password (bcrypt.CompareHashAndPassword)
3. Si válido, genera JWT firmado con secret
4. Frontend guarda token (localStorage o cookie)
5. Cada request incluye: `Authorization: Bearer <token>`
6. Cada API valida token con middleware

**Middleware de autenticación** (en cada API):
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extraer token del header Authorization
        // 2. Validar firma y expiración
        // 3. Extraer claims (user_id, role)
        // 4. Guardar en contexto
        // 5. c.Next() o c.AbortWithStatus(401)
    }
}
```

**Middleware de autorización Admin**:
```go
func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role := c.GetString("role")
        if role != "admin" {
            c.AbortWithStatusJSON(403, gin.H{"error": "Admin access required"})
            return
        }
        c.Next()
    }
}
```

---

## 📊 Bases de Datos

### MySQL (users-api)

**Conexión**:
```
Host: mysql:3306
User: gym_user
Password: gym_pass
Database: gym_users
```

**Schema**:
```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,  -- bcrypt hash
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    role ENUM('normal', 'admin') DEFAULT 'normal',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
);
```

### MongoDB (activities-api, bookings-api)

**Conexión**:
```
URI: mongodb://admin:admin123@mongodb:27017
Database: gym_db
```

**Colecciones**:

1. **activities**
```javascript
{
  _id: ObjectId("..."),
  owner_id: 1,
  name: "Yoga Intermedio",
  description: "Clase de yoga nivel intermedio...",
  category: "Yoga",
  duration: 60,
  price: 1500.00,
  image_url: "https://...",
  status: "active",
  created_at: ISODate("..."),
  updated_at: ISODate("...")
}
```

2. **schedules**
```javascript
{
  _id: ObjectId("..."),
  activity_id: "507f1f77bcf86cd799439011",
  instructor: "María González",
  day_of_week: "monday",
  start_time: "18:00",
  end_time: "19:00",
  location: "Sala 1 - Sede Centro",
  max_capacity: 20,
  current_bookings: 15,
  status: "active",
  created_at: ISODate("..."),
  updated_at: ISODate("...")
}
```

3. **bookings**
```javascript
{
  _id: ObjectId("..."),
  user_id: 5,
  schedule_id: "507fabc...",
  // Datos desnormalizados (snapshot)
  activity_name: "Yoga Intermedio",
  activity_category: "Yoga",
  instructor: "María González",
  day_of_week: "monday",
  start_time: "18:00",
  end_time: "19:00",
  location: "Sala 1",
  price: 1500.00,
  status: "confirmed",
  created_at: ISODate("..."),
  updated_at: ISODate("...")
}
```

**Índices importantes**:
```javascript
// schedules
db.schedules.createIndex({ activity_id: 1 })
db.schedules.createIndex({ day_of_week: 1, start_time: 1 })
db.schedules.createIndex({ location: 1, day_of_week: 1, start_time: 1 }, { unique: true })

// bookings
db.bookings.createIndex({ user_id: 1 })
db.bookings.createIndex({ schedule_id: 1 })
db.bookings.createIndex({ user_id: 1, schedule_id: 1 }, { unique: true })
```

---

## 🔍 Apache Solr

**Core**: `schedules`

**Schema** (campos indexados):
```xml
<field name="id" type="string" indexed="true" stored="true" required="true"/>
<field name="schedule_id" type="string" indexed="true" stored="true"/>
<field name="activity_id" type="string" indexed="true" stored="true"/>
<field name="activity_name" type="text_general" indexed="true" stored="true"/>
<field name="activity_category" type="string" indexed="true" stored="true"/>
<field name="instructor" type="text_general" indexed="true" stored="true"/>
<field name="day_of_week" type="string" indexed="true" stored="true"/>
<field name="start_time" type="string" indexed="true" stored="true"/>
<field name="location" type="string" indexed="true" stored="true"/>
<field name="available_spots" type="int" indexed="true" stored="true"/>
<field name="price" type="double" indexed="true" stored="true"/>
<field name="status" type="string" indexed="true" stored="true"/>
```

**Ejemplo de consulta**:
```
GET http://localhost:8983/solr/schedules/select?
  q=yoga
  &fq=day_of_week:monday
  &fq=available_spots:[1 TO *]
  &fq=status:active
  &sort=start_time asc
  &start=0
  &rows=10
```

---

## 📨 RabbitMQ

**Exchange**: `schedules_exchange` (tipo: topic)
**Queue**: `schedules_queue`
**Routing Key**: `schedule.*`

**Bindings**:
- `schedule.create` → schedules_queue
- `schedule.update` → schedules_queue
- `schedule.delete` → schedules_queue

**Formato de mensaje**:
```json
{
  "operation": "CREATE|UPDATE|DELETE",
  "schedule_id": "507f1f77bcf86cd799439011",
  "activity_id": "507f1f77bcf86cd799439022",
  "timestamp": 1699123456789
}
```

**Publicador**: activities-api (cuando se crea/modifica/elimina un schedule)
**Consumidor**: search-api (indexa cambios en Solr)

---

## 💾 Estrategia de Caché (search-api)

### Nivel 1: CCache (Local en memoria)

**Características**:
- Caché en memoria del proceso
- Muy rápido (nanosegundos)
- No compartida entre instancias
- TTL: 5 minutos
- Tamaño: LRU con límite de items

**Uso**:
```go
import "github.com/karlseguin/ccache"

cache := ccache.New(ccache.Configure().MaxSize(1000))
cache.Set(key, value, 5*time.Minute)
item := cache.Get(key)
```

### Nivel 2: Memcached (Distribuida)

**Características**:
- Compartida entre todas las instancias
- Rápido (milisegundos)
- TTL: 10 minutos
- Host: memcached:11211

**Uso**:
```go
import "github.com/bradfitz/gomemcache/memcache"

mc := memcache.New("memcached:11211")
mc.Set(&memcache.Item{Key: key, Value: value, Expiration: 600})
item, _ := mc.Get(key)
```

### Nivel 3: Solr (Fuente de verdad)

**Cuando se consulta**:
- Cache miss en L1 y L2
- Primera búsqueda de una query específica

**Invalidación de caché**:
- Cuando se recibe evento UPDATE/DELETE de RabbitMQ
- Se eliminan keys relacionadas con ese schedule

---

## 🧪 Testing

### Unit Tests

**Ubicación**: `*_test.go` en cada paquete

**Ejemplo** (activities-api):
```go
// services/activity_service_test.go
func TestCreateActivity_Success(t *testing.T) {
    // Arrange
    mockRepo := new(MockActivityRepository)
    mockUserClient := new(MockUserClient)
    service := NewActivityService(mockRepo, mockUserClient)
    
    // Act
    result, err := service.CreateActivity(activity)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

**Ejecutar tests**:
```bash
cd backend/activities-api
go test ./... -v
```

### Integration Tests

**Testing manual con curl**:
```bash
# Login
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Crear actividad
curl -X POST http://localhost:8081/activities \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Yoga","category":"Yoga","duration":60,"price":1500}'
```

---

## 📈 Escalabilidad

### Horizontal Scaling

**Servicios stateless** (pueden replicarse):
- users-api
- activities-api
- bookings-api
- search-api (con caché distribuida)

**Ejemplo con Docker Compose**:
```yaml
activities-api:
  deploy:
    replicas: 3
```

### Load Balancing

**NGINX como reverse proxy**:
```nginx
upstream activities_backend {
    server activities-api-1:8081;
    server activities-api-2:8081;
    server activities-api-3:8081;
}
```

### Optimizaciones

1. **Índices en MongoDB**: Queries rápidas
2. **Caché multinivel**: Reduce carga en Solr
3. **Conexión pool**: Reutilizar conexiones DB
4. **Paginación**: Limitar resultados
5. **Desnormalización**: Evitar joins en tiempo real

---

## 🛠️ Patrón MVC en Go

### Estructura de cada API

```
api/
├── domain/
│   ├── entity.go       # Structs de entidades
│   ├── dto.go          # Data Transfer Objects
│   └── errors.go       # Errores custom
├── repositories/
│   ├── interface.go    # Repository interface
│   └── impl.go         # Implementación (MySQL/MongoDB)
├── services/
│   ├── interface.go    # Service interface
│   ├── impl.go         # Lógica de negocio
│   └── impl_test.go    # Tests
├── controllers/
│   └── controller.go   # Handlers HTTP
├── middlewares/
│   └── auth.go         # Middlewares
├── clients/
│   └── http_client.go  # Clientes HTTP para otras APIs
├── config/
│   └── database.go     # Configuración
└── main.go             # Entry point
```

### Flujo de request

```
HTTP Request
    ↓
Controller (validación, parsing)
    ↓
Service (lógica de negocio)
    ↓
Repository (acceso a datos)
    ↓
Database/External API
    ↓
Repository devuelve entidad
    ↓
Service procesa
    ↓
Controller serializa response
    ↓
HTTP Response (JSON)
```

---

## 🚀 Deployment

### Docker Compose (Desarrollo)

```bash
# Levantar todo
docker-compose up -d

# Ver logs
docker-compose logs -f

# Reiniciar un servicio
docker-compose restart activities-api
```

### Producción (Recomendaciones)

1. **Kubernetes** para orquestación
2. **ConfigMaps** para configuración
3. **Secrets** para credenciales
4. **Ingress** para routing
5. **HPA** (Horizontal Pod Autoscaler)
6. **Prometheus + Grafana** para monitoreo
7. **ELK Stack** para logging centralizado

---

## 📚 Referencias

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [Apache Solr](https://solr.apache.org/)
- [RabbitMQ](https://www.rabbitmq.com/)
- [JWT](https://jwt.io/)