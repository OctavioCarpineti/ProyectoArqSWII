# 🔧 Solr Indexing Fix - Explicación Detallada

## 🐛 Problema Detectado

Cuando ejecutabas `test-search-api.sh`, el Test 10 fallaba con:

```
Test 10: Obtener schedule por ID desde Solr
⚠️ No se pudo obtener desde Solr (HTTP 404)
Response: {"error":"Schedule not found"}
```

## 🔍 Diagnóstico

### ¿Qué estaba pasando?

El flujo asíncrono funciona así:
1. `activities-api` crea un schedule → publica evento a RabbitMQ
2. RabbitMQ recibe el evento y lo encola en `schedules_queue`
3. `search-api` consumer recibe el mensaje de la queue
4. `search-api` intenta indexar en Solr
5. **❌ AQUÍ FALLABA** - Solr rechazaba el documento

### ¿Por qué fallaba?

Al revisar los logs de `search-api`:

```bash
docker-compose logs search-api | grep -i error
```

Encontramos dos errores:

#### Error 1: Campo `activity_price` no existe
```json
{
  "error": {
    "msg": "ERROR: [doc=69166800f9e393aa8488b120] unknown field 'activity_price'",
    "code": 400
  }
}
```

**Causa**:
- En el código Go (`backend/search-api/domain/schedule_search.go`), el campo estaba definido como:
  ```go
  ActivityPrice float64 `json:"activity_price"`
  ```
- En el schema de Solr (`config/solr/schedules/conf/schema.xml`), el campo se llamaba:
  ```xml
  <field name="price" type="double" .../>
  ```

#### Error 2: Campos `created_at` y `updated_at` no existen
```json
{
  "error": {
    "msg": "ERROR: [doc=...] unknown field 'created_at'",
    "code": 400
  }
}
```

**Causa**: Los campos `created_at` y `updated_at` estaban en el struct Go pero faltaban en el schema de Solr.

---

## ✅ Solución Aplicada

### Fix 1: Corregir nombre del campo `price`

**Archivo**: `backend/search-api/domain/schedule_search.go`

```diff
- ActivityPrice    float64 `json:"activity_price"`
+ ActivityPrice    float64 `json:"price"` // Changed from "activity_price" to match Solr schema
```

### Fix 2: Agregar campos de timestamp al schema

**Archivo**: `config/solr/schedules/conf/schema.xml`

```xml
<field name="created_at" type="date" indexed="true" stored="true"/>
<field name="updated_at" type="date" indexed="true" stored="true"/>
```

### Pasos ejecutados:

1. Modificar `schedule_search.go` (línea 16)
2. Modificar `schema.xml` (líneas 46-47)
3. Rebuild `search-api`:
   ```bash
   docker-compose build search-api
   ```
4. Restart servicios afectados:
   ```bash
   docker-compose restart search-api solr
   ```
5. Purgar mensajes fallidos de RabbitMQ:
   ```bash
   docker-compose exec rabbitmq rabbitmqctl purge_queue schedules_queue
   ```

---

## 🧪 Verificación del Fix

### Antes del fix:
```bash
./scripts/debug-solr.sh
```
Output:
```
Total de documentos indexados: 0
⚠️ NO HAY DOCUMENTOS INDEXADOS
```

### Después del fix:
```bash
./scripts/debug-solr.sh
```
Output:
```
Total de documentos indexados: 1
✅ Hay 1 documento(s) indexado(s)
```

### Logs del consumer:
```bash
docker-compose logs search-api | grep -E "indexando|✅|❌"
```

**Antes**:
```
❌ Error procesando mensaje: failed to index in solr: solr returned status 400
```

**Después**:
```
✅ Schedule 69166b2d9f42b8bb2e343b44 indexado exitosamente
```

---

## 📚 Cómo Debuggear en el Futuro

### Scripts de debugging

Creamos dos scripts para facilitar el debugging del flujo asíncrono:

#### 1. `./scripts/debug-rabbitmq.sh`

Verifica:
- ✅ RabbitMQ está corriendo
- 📡 Exchange `schedules_exchange` existe
- 📬 Queue `schedules_queue` existe
- 🔗 Bindings correctos
- 🔌 Consumers activos
- 📨 Estado de mensajes

**Uso**:
```bash
./scripts/debug-rabbitmq.sh
```

#### 2. `./scripts/debug-solr.sh`

Verifica:
- ✅ Solr está corriendo
- 📚 Core `schedules` existe
- 📄 Cantidad de documentos indexados
- 📋 Listado de documentos
- 🔎 Pruebas de búsqueda

**Uso**:
```bash
./scripts/debug-solr.sh
```

### Debugging manual

#### Ver logs del consumer en tiempo real:
```bash
docker-compose logs -f search-api | grep -E "Mensaje recibido|indexando|Error"
```

#### Verificar mensajes en RabbitMQ:
```bash
# Web UI
open http://localhost:15672
# Usuario: guest / Contraseña: guest

# CLI
docker-compose exec rabbitmq rabbitmqctl list_queues
```

#### Query directo a Solr:
```bash
curl "http://localhost:8983/solr/schedules/select?q=*:*&wt=json&indent=true"
```

#### Purgar queue si hay mensajes atascados:
```bash
docker-compose exec rabbitmq rabbitmqctl purge_queue schedules_queue
```

---

## ✅ Checklist de Verificación

Cuando implementes cambios en el flujo search-api → Solr, verifica:

- [ ] El struct Go (`ScheduleSearch`) tiene campos JSON que coinciden con el schema de Solr
- [ ] El schema de Solr (`schema.xml`) tiene definidos todos los campos que se indexan
- [ ] Los tipos de datos coinciden (string, int, double, date)
- [ ] Los campos required en Solr están presentes en el Go struct
- [ ] Después de cambiar el schema, reiniciar Solr: `docker-compose restart solr`
- [ ] Después de cambiar código Go, rebuild: `docker-compose build search-api`
- [ ] Verificar logs del consumer para errores
- [ ] Usar scripts de debugging para validar flujo completo

---

## 🎯 Conclusión

El problema era un **mismatch entre el schema de Solr y el struct de Go**.

**Lecciones aprendidas**:
1. Siempre verificar logs del consumer cuando Solr no indexa
2. Los nombres de campos JSON en Go deben coincidir con los campos en Solr schema
3. Usar scripts de debugging para diagnosticar problemas del flujo asíncrono
4. El flujo RabbitMQ → Consumer → Solr funciona correctamente cuando hay schema matching

**Estado actual**: ✅ TODO FUNCIONANDO
- El consumer procesa eventos correctamente
- Solr indexa documentos sin errores
- Las búsquedas retornan resultados

---

## 📝 Referencias

- Logs del consumer: `docker-compose logs -f search-api`
- RabbitMQ UI: http://localhost:15672
- Solr UI: http://localhost:8983
- Script debug RabbitMQ: `./scripts/debug-rabbitmq.sh`
- Script debug Solr: `./scripts/debug-solr.sh`
