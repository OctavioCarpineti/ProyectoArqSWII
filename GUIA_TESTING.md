# 🧪 Guía de Testing - Primera Entrega

Esta guía te permitirá verificar que el sistema funciona correctamente antes de la entrega.

## ⏱️ Tiempo estimado: 15-20 minutos

---

## 📋 Pre-requisitos

- Docker y Docker Compose instalados
- Puertos disponibles: 3000, 8080, 8081, 8082, 8083, 3307, 27017, 5672, 8983, 11211
- Terminal con bash (Linux/Mac) o Git Bash (Windows)

---

## 🚀 Paso 1: Levantar todos los servicios (5 min)

```bash
# Limpiar contenedores previos (si existen)
docker-compose down -v

# Levantar todos los servicios
docker-compose up --build -d

# Esperar a que todos los servicios estén healthy (~2 minutos)
docker-compose ps
```

**✅ Verificación esperada:**
Todos los servicios deben mostrar status "Up" o "healthy":
- mysql (healthy)
- mongodb (healthy)
- rabbitmq (healthy)
- solr (healthy)
- memcached (healthy)
- users-api (Up)
- activities-api (Up)
- bookings-api (Up)
- search-api (Up)
- frontend (Up)

**⚠️ Si algún servicio falla:**
```bash
# Ver logs del servicio problemático
docker-compose logs [nombre-servicio]

# Ejemplo:
docker-compose logs users-api
```

---

## 🧪 Paso 2: Ejecutar tests de integración (5 min)

Estos scripts verifican que las APIs funcionen correctamente y crean datos de prueba.

```bash
# Dar permisos de ejecución
chmod +x scripts/test-*.sh

# Ejecutar todos los tests en orden
./scripts/test-all.sh
```

**✅ Verificación esperada:**
Deberías ver mensajes verdes (✅) en:
1. **users-api**: Login, creación de usuarios, validación JWT
2. **activities-api**: Creación de actividades, creación de schedules con concurrencia
3. **search-api**: Búsqueda con Solr, filtros, cache
4. **bookings-api**: Creación de reservas con concurrencia, listado, cancelación

**Datos creados:**
- Usuario: `testuser` / Password: `user123`
- Usuario admin: `admin` / Password: `admin123`
- 2 actividades: Yoga y Spinning
- 2 horarios por actividad
- Al menos 1 reserva de ejemplo

**⚠️ Si los tests fallan:**
```bash
# Verificar que todos los servicios estén UP
docker-compose ps

# Esperar 30 segundos más para que los servicios terminen de inicializar
sleep 30

# Reintentar tests individuales
./scripts/test-users-api.sh
./scripts/test-activities-api.sh
./scripts/test-search-api.sh
./scripts/test-bookings-api.sh
```

---

## 🌐 Paso 3: Probar el Frontend (5-10 min)

### 3.1. Acceder a la aplicación

Abrir en el navegador: **http://localhost:3000**

**✅ Verificación esperada:**
- Debe cargar la página de Login
- Debe verse un diseño moderno con fondo morado/gradiente
- Debe mostrar las credenciales de prueba

---

### 3.2. Flujo completo (CRÍTICO para la entrega)

Este es el flujo que evaluará el profesor:

#### **1️⃣ LOGIN**
- **Acción:** Ingresar `testuser` / `user123`
- **Resultado esperado:** Redirige a `/search`

#### **2️⃣ BÚSQUEDA**
- **Acción:** Deberías ver tarjetas con clases disponibles (Yoga, Spinning, etc.)
- **Filtros disponibles:**
  - Búsqueda general (texto)
  - Categoría (cardio, flexibility, etc.)
  - Día de la semana
  - Instructor
- **Resultado esperado:** Lista de schedules con:
  - Nombre de actividad
  - Instructor
  - Horario
  - Cupos disponibles

#### **3️⃣ DETALLE**
- **Acción:** Click en "Ver Detalles" de cualquier clase
- **Resultado esperado:** Página con:
  - Toda la información del horario
  - Badge de disponibilidad
  - Botón "Reservar esta clase"

#### **4️⃣ ACCIÓN (Reserva)**
- **Acción:** Click en "Reservar esta clase"
- **Resultado esperado:**
  - Si hay cupos: redirige a `/congrats`
  - Si ya reservaste: mensaje de error

#### **5️⃣ CONGRATS**
- **Acción:** Deberías ver la página de confirmación
- **Resultado esperado:**
  - Emoji de celebración 🎉
  - Mensaje de confirmación
  - Resumen de la reserva
  - Botón "Buscar más clases"

---

### 3.3. Verificaciones técnicas

#### **A) Verificar que el token se guarda:**
```
1. Abrir DevTools (F12)
2. Ir a: Application → Local Storage → http://localhost:3000
3. Verificar que existan:
   - token: "eyJ..."
   - user: {"id":1,"username":"testuser",...}
```

#### **B) Verificar llamadas a las APIs:**
```
1. DevTools (F12) → Network
2. Hacer una búsqueda
3. Verificar llamadas a:
   - http://localhost:8083/search?... (200 OK)
4. Ver detalles de una clase:
   - http://localhost:8083/search/[id] (200 OK)
5. Hacer una reserva:
   - http://localhost:8082/bookings (201 Created)
```

#### **C) Verificar autorización:**
```
1. Network → Headers de la llamada a /bookings
2. Verificar que incluye:
   Authorization: Bearer eyJ...
```

---

## 🔍 Paso 4: Verificar Unit Tests (1 min)

```bash
# Ir al directorio de activities-api
cd backend/activities-api

# Ejecutar tests
go test -v ./services/

# Resultado esperado: 9/9 tests PASS
```

**✅ Verificación esperada:**
```
=== RUN   TestCreateSchedule_Success
--- PASS: TestCreateSchedule_Success (0.00s)
=== RUN   TestCreateSchedule_ActivityNotFound
--- PASS: TestCreateSchedule_ActivityNotFound (0.00s)
...
PASS
ok  	activities-api/services	0.015s
```

---

## 📸 Paso 5: Capturas para el profesor (opcional)

Si el profesor pide evidencia, tomar capturas de:

1. **docker-compose ps** mostrando todos los servicios UP
2. **Salida de ./scripts/test-all.sh** con todos los ✅
3. **Página de Login** del frontend
4. **Página de Búsqueda** con resultados
5. **Página de Congrats** después de reservar
6. **go test -v ./services/** mostrando 9/9 PASS

---

## 🐛 Troubleshooting

### Problema: "Cannot connect to database"
```bash
# Esperar más tiempo para que las bases de datos inicialicen
docker-compose logs mysql
docker-compose logs mongodb

# Reiniciar servicios
docker-compose restart users-api activities-api
```

### Problema: "Frontend no carga"
```bash
# Ver logs del frontend
docker-compose logs frontend

# Verificar que el build se hizo correctamente
docker-compose up --build frontend
```

### Problema: "No aparecen clases en Search"
```bash
# Verificar que los datos se crearon
./scripts/test-activities-api.sh

# Verificar que Solr tiene los documentos
curl http://localhost:8983/solr/schedules/select?q=*:*
```

### Problema: "Error al reservar"
```bash
# Verificar logs de bookings-api
docker-compose logs bookings-api

# Verificar que el token es válido
# En DevTools → Application → Local Storage → verificar 'token'
```

---

## ✅ Checklist Final

Antes de la entrega, verificar:

- [ ] `docker-compose ps` muestra todos los servicios UP
- [ ] `./scripts/test-all.sh` pasa todos los tests (✅)
- [ ] `go test -v ./services/` muestra 9/9 PASS
- [ ] Frontend carga en http://localhost:3000
- [ ] Puedo hacer login con testuser/user123
- [ ] Veo clases en la página de búsqueda
- [ ] Puedo ver detalles de una clase
- [ ] Puedo hacer una reserva
- [ ] Veo la página de Congrats después de reservar

---

## 📊 Resumen de Tecnologías

**Backend:**
- Go 1.24 con Gin framework
- MySQL (users-api)
- MongoDB (activities-api, bookings-api)
- Apache Solr (search-api)
- RabbitMQ (eventos entre microservicios)
- Memcached (cache distribuido)

**Frontend:**
- React 18 + Vite
- React Router 6
- Axios para HTTP
- Nginx para producción

**DevOps:**
- Docker multi-stage builds
- Docker Compose para orquestación
- Health checks en todos los servicios

**Concurrencia (BONUS):**
- Goroutines + Channels + WaitGroup
- Implementada en:
  - activities-api/CreateSchedule (3 goroutines paralelas)
  - bookings-api/CreateBooking (3 goroutines paralelas)

**Testing:**
- Unit tests (Go): 9 tests en schedule_service_test.go
- Integration tests (Bash): 4 scripts para todas las APIs
- E2E: Flujo manual en frontend

---

## 🎯 Cumplimiento de Requisitos del PDF

| Requisito | Estado | Evidencia |
|-----------|--------|-----------|
| Login → Búsqueda → Detalle → Acción → Congrats | ✅ | Flujo completo en frontend |
| Frontend funcionando | ✅ | http://localhost:3000 |
| Backend funcionando | ✅ | 4 microservicios + APIs |
| Al menos un microservicio con tests | ✅ | activities-api: 9 unit tests |
| Concurrencia (opcional) | ✅ BONUS | Goroutines en 2 servicios |

---

## 📞 Contacto

Si encuentras algún problema durante las pruebas, verifica:
1. Que todos los puertos estén libres
2. Que Docker tenga suficiente memoria asignada (al menos 4GB)
3. Que los logs no muestren errores críticos

**¡Éxito en la entrega! 🚀**
