# Fase 2: Integración de Servicios Externos - Cambios Realizados

Fecha: 2025-12-27

## Objetivo
Integrar servicios externos para mejorar la escalabilidad y funcionalidad de la plataforma: **RabbitMQ** (cola persistente), **MinIO** (almacenamiento distribuido preparado), y **Elasticsearch** (búsqueda avanzada).

## Cambios Implementados

### 1. Configuración de Servicios Externos

#### Actualizado `pkg/config/config.go`
- ✅ Agregados campos para RabbitMQ:
  - `RabbitMQURL` (default: `amqp://admin:password123@localhost:5672/`)

- ✅ Agregados campos para MinIO:
  - `MinIOEndpoint` (default: `localhost:9000`)
  - `MinIOAccessKey` (default: `minioadmin`)
  - `MinIOSecretKey` (default: `minioadmin123`)
  - `MinIOBucket` (default: `videos`)
  - `MinIOUseSSL` (default: `false`)

- ✅ Agregado campo para Elasticsearch:
  - `ElasticsearchURL` (default: `http://localhost:9200`)

- ✅ Agregada función helper `getEnvAsBool()`

### 2. Cliente RabbitMQ

#### Creado `pkg/database/rabbitmq.go`
Cliente wrapper para RabbitMQ con las siguientes funcionalidades:
- ✅ Conexión y reconexión automática
- ✅ Declaración de colas durables
- ✅ Publicación de mensajes persistentes
- ✅ Consumo de mensajes con ACK manual
- ✅ QoS configuration (prefetch: 10)
- ✅ Health check
- ✅ Operaciones de administración (purge, delete, info)

#### Creado `internal/core/ports/output/queue_repository.go`
Puerto de salida (interfaz) para gestión de colas:
```go
type QueueRepository interface {
    PublishJob(ctx context.Context, job JobMessage) error
    ConsumeJobs(ctx context.Context, queueName string, handler func(JobMessage) error) error
    GetQueueInfo(ctx context.Context, queueName string) (int, int, error)
    PurgeQueue(ctx context.Context, queueName string) (int, error)
    Close() error
    HealthCheck() error
}
```

#### Creado `internal/adapters/output/persistence/rabbitmq/queue_repository.go`
Implementación del puerto usando RabbitMQ:
- ✅ Sistema de reintentos (3 intentos con delay de 5s)
- ✅ Reconexión automática en caso de fallo
- ✅ Mapeo de tipos de trabajo a colas específicas:
  - `transcoding` → `transcoding_queue`
  - `thumbnail` → `thumbnail_queue`
  - `analytics` → `analytics_queue`
  - `notification` → `notification_queue`
- ✅ Manejo de reintentos en mensajes fallidos
- ✅ Logging detallado de operaciones

### 3. Cliente MinIO (Preparado para Fase 3)

#### Creado `pkg/database/minio.go`
Cliente wrapper para MinIO/S3:
- ✅ Conexión y creación automática de buckets
- ✅ Upload/Download de archivos
- ✅ Eliminación de archivos (simple y múltiple)
- ✅ Verificación de existencia de archivos
- ✅ Listado de archivos con prefijo
- ✅ URLs pre-firmadas temporales
- ✅ Copia de archivos
- ✅ Health check

#### Creado `internal/core/ports/output/storage_repository.go`
Puerto de salida para almacenamiento distribuido:
```go
type StorageRepository interface {
    UploadFile(ctx context.Context, path string, reader io.Reader, size int64, contentType string) error
    DownloadFile(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteFile(ctx context.Context, path string) error
    FileExists(ctx context.Context, path string) (bool, error)
    GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
    // ... más métodos
}
```

#### Creado `internal/adapters/output/persistence/minio/storage_repository.go`
Implementación del puerto usando MinIO.

**Nota:** MinIO está completamente integrado y listo para usar, pero NO se usa en el flujo actual (Fase 3).

### 4. Cliente Elasticsearch

#### Creado `pkg/database/elasticsearch.go`
Cliente wrapper para Elasticsearch 7.x:
- ✅ Indexación y actualización de documentos
- ✅ Eliminación de documentos
- ✅ Búsqueda con queries complejas
- ✅ Bulk indexing
- ✅ Gestión de índices (create, delete, exists)
- ✅ Refresh de índices
- ✅ Health check

#### Creado `internal/core/ports/output/search_repository.go`
Puerto de salida para búsqueda avanzada:
```go
type SearchRepository interface {
    IndexVideo(ctx context.Context, video *domain.Video) error
    UpdateVideoIndex(ctx context.Context, video *domain.Video) error
    DeleteVideoIndex(ctx context.Context, videoID string) error
    SearchVideos(ctx context.Context, query string, filters map[string]interface{}, page, limit int) ([]*domain.Video, int64, error)
    SuggestVideos(ctx context.Context, partial string, limit int) ([]string, error)
    HealthCheck() error
    InitializeIndex(ctx context.Context) error
}
```

#### Creado `internal/adapters/output/persistence/elasticsearch/search_repository.go`
Implementación del puerto usando Elasticsearch:
- ✅ Índice `videos` con mapping optimizado para búsqueda
- ✅ Búsqueda multi-match en title, description y tags
- ✅ Boost en título (weight: 3x)
- ✅ Filtrado por categoría, tags, estado y visibilidad
- ✅ Ordenamiento por relevancia y popularidad
- ✅ Paginación eficiente
- ✅ Sugerencias de autocompletado (preparado)

### 5. Integración en la Aplicación

#### Modificado `cmd/server/main.go`
- ✅ Agregadas conexiones a RabbitMQ, MinIO y Elasticsearch
- ✅ Inicialización de repositorios adaptadores
- ✅ Inicialización automática del índice de Elasticsearch
- ✅ Consumo de trabajos desde RabbitMQ en goroutine
- ✅ Conexión del WorkerPool con QueueRepository
- ✅ Graceful shutdown de todos los servicios

**Flujo de Queue:**
```go
// Publicar trabajo (VideoHandler)
queueRepo.PublishJob(ctx, job)
    → RabbitMQ (persistente)

// Consumir trabajo (main.go)
queueRepo.ConsumeJobs(ctx, "transcoding_queue", handler)
    → WorkerPool.ProcessJobFromQueue(job)
    → TranscodingWorker.ProcessJob(job)
```

#### Modificado `internal/adapters/input/http/video_handler.go`
- ✅ Reemplazado `workerPool` por `queueRepo`
- ✅ Publicación de trabajos a RabbitMQ en lugar de canal en memoria
- ✅ Mejores mensajes de log

#### Modificado `internal/infrastructure/workers/worker_pool.go`
- ✅ Agregado método `ProcessJobFromQueue(jobMessage interface{}) error`
- ✅ Conversión de mensajes de RabbitMQ a formato interno del WorkerPool

### 6. Dependencias Agregadas

```bash
go get github.com/rabbitmq/amqp091-go      # RabbitMQ client
go get github.com/minio/minio-go/v7        # MinIO/S3 client
go get github.com/elastic/go-elasticsearch/v7  # Elasticsearch client
```

## Arquitectura Resultante

### Antes (Fase 1):
```
Upload → VideoHandler → Canal en memoria (jobQueue)
                      → WorkerPool → TranscodingWorker

Problema: Jobs se pierden al reiniciar ❌
```

### Ahora (Fase 2):
```
Upload → VideoHandler → RabbitMQ (persistente, durable)
                      ↓
                   Consumer (goroutine)
                      ↓
                   WorkerPool → TranscodingWorker

Ventaja: Jobs sobreviven reinicios ✅
```

### Elasticsearch (Preparado):
```
Video creado/actualizado → Indexar en Elasticsearch
Búsqueda de usuario → Query a Elasticsearch (full-text, relevancia)
```

### MinIO (Preparado para Fase 3):
```
Upload → Filesystem local (actual)
       → MinIO (futuro, distribuido, escalable)
```

## Archivos Creados

### Clientes (pkg/database/):
1. `rabbitmq.go` - Cliente RabbitMQ
2. `minio.go` - Cliente MinIO
3. `elasticsearch.go` - Cliente Elasticsearch

### Puertos de Salida (internal/core/ports/output/):
4. `queue_repository.go` - Puerto para colas
5. `storage_repository.go` - Puerto para almacenamiento
6. `search_repository.go` - Puerto para búsqueda

### Adaptadores (internal/adapters/output/persistence/):
7. `rabbitmq/queue_repository.go` - Implementación RabbitMQ
8. `minio/storage_repository.go` - Implementación MinIO
9. `elasticsearch/search_repository.go` - Implementación Elasticsearch

## Archivos Modificados

1. **pkg/config/config.go**
   - Líneas 40-51: Agregados campos de configuración
   - Líneas 83-95: Carga de variables de entorno
   - Líneas 114-121: Función helper getEnvAsBool

2. **cmd/server/main.go**
   - Líneas 15-22: Imports de nuevos adaptadores
   - Líneas 53-79: Conexiones a servicios externos
   - Líneas 82-88: Creación de repositorios adaptadores
   - Líneas 90-95: Inicialización de índice Elasticsearch
   - Líneas 110-118: Consumo de trabajos desde RabbitMQ
   - Línea 123: VideoHandler con queueRepo

3. **internal/adapters/input/http/video_handler.go**
   - Línea 19: Import de output ports
   - Líneas 25-39: Struct y constructor modificados
   - Líneas 182-195: Publicación a RabbitMQ

4. **internal/infrastructure/workers/worker_pool.go**
   - Líneas 120-144: Nuevo método ProcessJobFromQueue

## Estado de Integración

### ✅ ACTIVO - RabbitMQ
- Cola persistente funcionando
- Jobs sobreviven reinicios
- Reintentos automáticos
- Logging completo

### ✅ PREPARADO - Elasticsearch
- Índice creado automáticamente
- Mapping optimizado
- **Pendiente:** Indexación automática en VideoService (Fase 3)
- **Pendiente:** Endpoint de búsqueda usando Elasticsearch

### ✅ PREPARADO - MinIO
- Cliente completamente funcional
- **Pendiente:** Migración de filesystem a MinIO (Fase 3)
- **Pendiente:** Upload directo a MinIO en VideoHandler
- **Pendiente:** Serving de videos desde MinIO

## Beneficios Obtenidos

### 1. Persistencia de Trabajos (RabbitMQ)
- ✅ Jobs no se pierden al reiniciar el servidor
- ✅ Distribución de carga entre múltiples workers
- ✅ Reintentos automáticos en caso de fallo
- ✅ Monitoreo de colas vía RabbitMQ Management (puerto 15672)

### 2. Escalabilidad Horizontal
- ✅ Múltiples instancias pueden consumir de la misma cola
- ✅ MinIO permite almacenamiento distribuido (cuando se active)
- ✅ Elasticsearch permite búsquedas rápidas a gran escala

### 3. Mejor Observabilidad
- ✅ RabbitMQ Management UI para monitoreo de colas
- ✅ MinIO Console para gestión de archivos
- ✅ Elasticsearch para analytics de búsquedas

### 4. Arquitectura Limpia
- ✅ Separación de responsabilidades (puertos y adaptadores)
- ✅ Fácil cambio de implementaciones (ej: RabbitMQ → Kafka)
- ✅ Testeable (mocks de puertos)

## Próximos Pasos (Fase 3)

### 1. Activar Elasticsearch Completamente
- [ ] Indexar automáticamente en VideoService.CreateVideo/UpdateVideo
- [ ] Crear endpoint GET /api/videos/search usando SearchRepository
- [ ] Agregar autocompletado con SuggestVideos
- [ ] Re-indexar videos existentes

### 2. Migrar a MinIO
- [ ] Actualizar TranscodingWorker para subir a MinIO
- [ ] Actualizar StreamingHandler para servir desde MinIO
- [ ] Migrar videos existentes de filesystem a MinIO
- [ ] URLs pre-firmadas para acceso temporal

### 3. Funcionalidades Pendientes de Fase 1
- [ ] Sistema de favoritos persistente (usando FavoriteRepository)
- [ ] Historial de visualización (usando WatchHistoryRepository)
- [ ] Sistema de comentarios
- [ ] Sistema de ratings
- [ ] Notificaciones
- [ ] Analytics completo

### 4. Mejoras de Producción
- [ ] Tests para nuevos adaptadores
- [ ] Circuit breakers para servicios externos
- [ ] Métricas de Prometheus para RabbitMQ/MinIO/Elasticsearch
- [ ] Health checks en endpoint /health
- [ ] Rate limiting avanzado por usuario

## Testing

### Para verificar la integración:

1. **Iniciar todos los servicios**:
   ```bash
   docker-compose up -d
   go run cmd/server/main.go
   ```

2. **Verificar conexiones**:
   - PostgreSQL: `localhost:5432`
   - Redis: `localhost:6379`
   - RabbitMQ Management: `http://localhost:15672` (admin/password123)
   - MinIO Console: `http://localhost:9001` (minioadmin/minioadmin123)
   - Elasticsearch: `http://localhost:9200`

3. **Subir un video** (mismo proceso que Fase 1):
   ```bash
   curl -X POST http://localhost:8080/api/videos \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -F "video=@test.mp4" \
     -F 'metadata={"title":"Test","description":"Test","category":"tech"}'
   ```

4. **Verificar en RabbitMQ**:
   - Acceder a http://localhost:15672
   - Ver queue `transcoding_queue`
   - Deberías ver el mensaje procesándose

5. **Verificar logs**:
   ```
   INFO: Job published successfully to queue transcoding_queue: transcoding
   INFO: Started consuming messages from queue: transcoding_queue
   INFO: Processing job: transcoding
   INFO: Starting transcoding for video ...
   INFO: Transcoding completed for video ...
   INFO: Job transcoding processed successfully
   ```

6. **Verificar Elasticsearch** (cuando se indexe):
   ```bash
   curl http://localhost:9200/videos/_count
   ```

## Notas Técnicas

### RabbitMQ
- Usa AMQP 0.9.1 protocol
- Mensajes marcados como persistentes (DeliveryMode: Persistent)
- Colas declaradas como durables
- QoS set a 10 (prefetch count)
- ACK manual para control de flujo

### MinIO
- Compatible con S3 API
- Bucket creado automáticamente si no existe
- **No está en uso todavía** (filesystem local actual)

### Elasticsearch
- Versión 7.17.5
- Índice `videos` con mapping customizado
- Single-node mode (desarrollo)
- **Indexación manual pendiente** (no automática aún)

## Troubleshooting

### Error: "failed to connect to RabbitMQ"
- Verificar que RabbitMQ esté corriendo: `docker-compose ps rabbitmq`
- Verificar URL en .env o config

### Error: "failed to create MinIO client"
- Verificar que MinIO esté corriendo: `docker-compose ps minio`
- Verificar endpoint y credenciales

### Error: "failed to create Elasticsearch client"
- Verificar que Elasticsearch esté corriendo: `docker-compose ps elasticsearch`
- Verificar URL en .env

### Jobs no se procesan
- Verificar logs del consumer
- Verificar RabbitMQ Management para ver mensajes en cola
- Verificar que WorkerPool esté corriendo

---

**Estado**: ✅ **Fase 2 Completada - Servicios Externos Integrados**

**Compilación**: ✅ Sin errores

**Próximo**: Fase 3 - Activar completamente Elasticsearch y MinIO
