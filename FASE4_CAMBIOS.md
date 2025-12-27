# Fase 4: Completar Favoritos e Historial - Cambios Realizados

Fecha: 2025-12-27

## Objetivo
Completar la implementación de favoritos e historial de visualización con endpoints HTTP funcionales y persistencia completa en PostgreSQL.

## Cambios Implementados

### 1. UserService - Gestión Completa de Favoritos ✅ COMPLETADO

#### Modificado `internal/core/services/user_service.go`
- ✅ Agregado `favoriteRepo` y `watchHistoryRepo` como dependencias
- ✅ **AddToFavorites**: Usa FavoriteRepository en lugar de cache
- ✅ **RemoveFromFavorites**: Usa FavoriteRepository con invalidación de cache
- ✅ **GetFavorites**: Obtiene favoritos paginados con información del video
- ✅ **GetWatchHistory**: Implementado con paginación
- ✅ **UpdateWatchProgress**: Guarda en PostgreSQL (sincronización entre dispositivos)
- ✅ **GetContinueWatching**: Obtiene videos para continuar viendo

**Cambios Clave:**
```go
type userService struct {
    userRepo         output.UserRepository
    cacheRepo        output.CacheRepository
    favoriteRepo     output.FavoriteRepository      // ✅ Nuevo
    watchHistoryRepo output.WatchHistoryRepository  // ✅ Nuevo
}

func NewUserService(
    userRepo output.UserRepository,
    cacheRepo output.CacheRepository,
    favoriteRepo output.FavoriteRepository,          // ✅ Nuevo
    watchHistoryRepo output.WatchHistoryRepository,  // ✅ Nuevo
) input.UserService
```

**Flujo de Favoritos:**
```
Usuario → AddToFavorites → PostgreSQL.favorites
                         → Cache invalidate

Usuario → GetFavorites → PostgreSQL JOIN videos
                       → [Video{...}, Video{...}]
```

**Flujo de Historial:**
```
Reproductor → UpdateWatchProgress (cada 30s)
           → PostgreSQL.watch_history (UPSERT)
           → Sincronizado entre dispositivos

Usuario → GetContinueWatching
       → PostgreSQL WHERE position > 10s AND < 90%
       → [WatchHistory{Video: {...}}]
```

### 2. WatchHistoryRepository ✅ IMPLEMENTADO

#### Creado `internal/adapters/output/persistence/postgres/watch_history_repository.go`
- ✅ **SaveWatchHistory**: UPSERT con ON CONFLICT (user_id, video_id)
- ✅ **GetWatchHistory**: Obtener progreso específico
- ✅ **GetUserWatchHistory**: Historial completo con paginación y JOIN a videos
- ✅ **DeleteWatchHistory**: Eliminar del historial
- ✅ **GetContinueWatching**: Videos a medio ver (>10s y <90%)

**Query Inteligente - Continue Watching:**
```sql
WHERE wh.position > 10
  AND v.duration > 0
  AND (CAST(wh.position AS FLOAT) / CAST(v.duration AS FLOAT)) < 0.9
  AND v.status = 'ready'
  AND v.is_public = true
ORDER BY wh.watched_at DESC
```

**Ventajas:**
- Videos vistos hace poco primero (watched_at DESC)
- Excluye videos apenas iniciados (<10s)
- Excluye videos casi terminados (>90%)
- Solo videos disponibles (ready + public)

### 3. Endpoints HTTP ✅ COMPLETADOS

#### Modificado `internal/adapters/input/http/user_handler.go`
Agregados nuevos endpoints:

**Favoritos:**
- ✅ `GetFavorites`: Lista videos favoritos con paginación
- ✅ `AddToFavorites`: Agregar video a favoritos
- ✅ `RemoveFromFavorites`: Quitar de favoritos

**Historial:**
- ✅ `GetWatchHistory`: Historial completo con paginación
- ✅ `UpdateWatchProgress`: Registrar progreso de visualización
- ✅ `GetContinueWatching`: Videos para continuar viendo

### 4. Rutas HTTP Registradas ✅ CONFIGURADAS

#### Modificado `cmd/server/main.go`

**Rutas de Favoritos:**
```go
protected.HandleFunc("/favorites", userHandler.GetFavorites).Methods("GET")
protected.HandleFunc("/favorites/{videoId}", userHandler.AddToFavorites).Methods("POST")
protected.HandleFunc("/favorites/{videoId}", userHandler.RemoveFromFavorites).Methods("DELETE")
```

**Rutas de Historial:**
```go
protected.HandleFunc("/history", userHandler.GetWatchHistory).Methods("GET")
protected.HandleFunc("/history/{videoId}/progress", userHandler.UpdateWatchProgress).Methods("POST")
protected.HandleFunc("/continue-watching", userHandler.GetContinueWatching).Methods("GET")
```

**Rutas de Configuración:**
```go
protected.HandleFunc("/user/settings", userHandler.GetSettings).Methods("GET")
protected.HandleFunc("/user/settings", userHandler.UpdateSettings).Methods("PUT")
```

### 5. Modelo de Dominio Actualizado ✅

#### Modificado `internal/core/domain/user.go`
```go
type WatchHistory struct {
    ID        uuid.UUID `json:"id" db:"id"`
    UserID    uuid.UUID `json:"userId" db:"user_id"`
    VideoID   uuid.UUID `json:"videoId" db:"video_id"`
    Video     *Video    `json:"video,omitempty"` // ✅ Agregado
    Position  int       `json:"position" db:"position"`
    Quality   string    `json:"quality" db:"quality"`
    WatchedAt time.Time `json:"watchedAt" db:"watched_at"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
```

### 6. Puerto de Entrada Actualizado ✅

#### Modificado `internal/core/ports/input/user_service.go`
```go
type UserService interface {
    // ... métodos existentes ...

    // ✅ Agregados en Fase 4
    GetContinueWatching(ctx context.Context, userID uuid.UUID, limit int) ([]domain.WatchHistory, error)
    UpdateWatchProgress(ctx context.Context, userID, videoID uuid.UUID, position int, quality string) error
}
```

### 7. Integración en main.go ✅

#### Modificado `cmd/server/main.go`
```go
// Output Adapters
favoriteRepo := postgresAdapter.NewFavoriteRepository(db)
watchHistoryRepo := postgresAdapter.NewWatchHistoryRepository(db)

// Core Services
userService := services.NewUserService(userRepo, cacheRepo, favoriteRepo, watchHistoryRepo)
```

## Arquitectura Resultante

### Flujo Completo de Favoritos

```
1. AGREGAR FAVORITO
   User → POST /api/favorites/VIDEO_ID
       → UserHandler.AddToFavorites
       → UserService.AddToFavorites
       → FavoriteRepository.AddFavorite
       → PostgreSQL INSERT (ON CONFLICT DO NOTHING)
       → Cache invalidate

2. LISTAR FAVORITOS
   User → GET /api/favorites?page=1&limit=20
       → UserHandler.GetFavorites
       → UserService.GetFavorites
       → FavoriteRepository.GetUserFavorites
       → PostgreSQL SELECT JOIN videos
       → [Video{...}, Video{...}]

3. QUITAR FAVORITO
   User → DELETE /api/favorites/VIDEO_ID
       → UserHandler.RemoveFromFavorites
       → UserService.RemoveFromFavorites
       → FavoriteRepository.RemoveFavorite
       → PostgreSQL DELETE
       → Cache invalidate
```

### Flujo Completo de Historial

```
1. VER VIDEO
   Frontend Player → Cada 30 segundos
                   → POST /api/history/VIDEO_ID/progress
                   → Body: {position: 157, quality: "720p"}
                   → UserHandler.UpdateWatchProgress
                   → UserService.UpdateWatchProgress
                   → WatchHistoryRepository.SaveWatchHistory
                   → PostgreSQL UPSERT (ON CONFLICT UPDATE position)

2. CONTINUAR VIENDO
   User → GET /api/continue-watching?limit=10
       → UserHandler.GetContinueWatching
       → UserService.GetContinueWatching
       → WatchHistoryRepository.GetContinueWatching
       → PostgreSQL WHERE position > 10 AND < 90%
       → [WatchHistory{Video: {...}, Position: 157}]

3. HISTORIAL COMPLETO
   User → GET /api/history?page=1&limit=20
       → UserHandler.GetWatchHistory
       → UserService.GetWatchHistory
       → WatchHistoryRepository.GetUserWatchHistory
       → PostgreSQL SELECT JOIN videos ORDER BY watched_at DESC
       → {data: [...], total: 47, page: 1, limit: 20}
```

### Sincronización Entre Dispositivos

```
Dispositivo A (Web) → Ve video hasta 2:37
                    → POST /api/history/.../progress {position: 157}
                    → PostgreSQL UPDATE watch_history

Dispositivo B (Mobile) → GET /api/continue-watching
                       → PostgreSQL SELECT
                       → Video {position: 157}  ✅ Sincronizado
                       → Reproduce desde 2:37
```

## Archivos Creados (1 nuevo)

1. ✅ `internal/adapters/output/persistence/postgres/watch_history_repository.go` (302 líneas)
   - SaveWatchHistory con UPSERT
   - GetUserWatchHistory con JOIN
   - GetContinueWatching con filtros inteligentes

## Archivos Modificados (5)

1. **internal/core/services/user_service.go** (~100 líneas modificadas)
   - Líneas 22-27: Agregado favoriteRepo y watchHistoryRepo
   - Líneas 30-42: Constructor actualizado
   - Líneas 243-253: AddToFavorites usa repositorio
   - Líneas 255-265: RemoveFromFavorites usa repositorio
   - Líneas 267-291: GetFavorites con conversión de tipos
   - Líneas 294-316: GetWatchHistory implementado
   - Líneas 318-333: UpdateWatchProgress sin cache
   - Líneas 335-354: GetContinueWatching implementado

2. **internal/core/ports/input/user_service.go** (3 métodos agregados)
   - Líneas 32-36: GetContinueWatching y UpdateWatchProgress

3. **internal/adapters/input/http/user_handler.go** (3 handlers agregados)
   - Líneas 360-398: UpdateWatchProgress
   - Líneas 400-425: GetContinueWatching

4. **cmd/server/main.go** (integración de repositorios)
   - Líneas 85-86: Creación de favoriteRepo y watchHistoryRepo
   - Línea 102: UserService con nuevas dependencias
   - Líneas 159-167: Rutas HTTP registradas

5. **internal/core/domain/user.go** (campo agregado)
   - Línea 55: Video *Video en WatchHistory

## Estadísticas de Código

- **Líneas agregadas:** ~550 líneas
- **Archivos nuevos:** 1
- **Archivos modificados:** 5
- **Endpoints HTTP nuevos:** 6
- **Tests:** 0 (pendiente Fase 5)

## API Endpoints - Guía de Uso

### Favoritos

**1. Listar Favoritos**
```bash
GET /api/favorites?page=1&limit=20
Authorization: Bearer TOKEN

Response:
{
  "data": [
    {
      "id": "uuid",
      "title": "Tutorial de Golang",
      "thumbnail": "/storage/thumbnails/...",
      "duration": 3600,
      ...
    }
  ],
  "total_count": 47,
  "page": 1,
  "limit": 20
}
```

**2. Agregar a Favoritos**
```bash
POST /api/favorites/VIDEO_UUID
Authorization: Bearer TOKEN

Response:
{
  "message": "Added to favorites successfully"
}
```

**3. Quitar de Favoritos**
```bash
DELETE /api/favorites/VIDEO_UUID
Authorization: Bearer TOKEN

Response:
{
  "message": "Removed from favorites successfully"
}
```

### Historial de Visualización

**1. Registrar Progreso** (Llamado automáticamente por el reproductor)
```bash
POST /api/history/VIDEO_UUID/progress
Authorization: Bearer TOKEN
Content-Type: application/json

{
  "position": 157,    // Segundos
  "quality": "720p"
}

Response:
{
  "message": "Watch progress updated successfully"
}
```

**2. Videos para Continuar Viendo**
```bash
GET /api/continue-watching?limit=10
Authorization: Bearer TOKEN

Response:
{
  "data": [
    {
      "id": "uuid",
      "videoId": "uuid",
      "video": {
        "id": "uuid",
        "title": "Tutorial de Docker",
        "duration": 3600,
        ...
      },
      "position": 157,
      "quality": "720p",
      "watchedAt": "2025-12-27T10:30:00Z"
    }
  ]
}
```

**3. Historial Completo**
```bash
GET /api/history?page=1&limit=20
Authorization: Bearer TOKEN

Response:
{
  "data": [...],
  "total_count": 89,
  "page": 1,
  "limit": 20
}
```

## Integración en Frontend

### JavaScript - Reproductor de Video

```javascript
// En el componente del reproductor de video
class VideoPlayer {
  constructor(videoId, initialPosition = 0) {
    this.videoId = videoId;
    this.videoElement = document.querySelector('video');
    this.progressInterval = null;

    // Cargar posición inicial
    if (initialPosition > 10) {
      this.videoElement.currentTime = initialPosition;
    }

    // Registrar progreso cada 30 segundos
    this.startProgressTracking();
  }

  startProgressTracking() {
    this.progressInterval = setInterval(() => {
      this.saveProgress();
    }, 30000); // 30 segundos

    // También guardar al pausar o cerrar
    this.videoElement.addEventListener('pause', () => this.saveProgress());
    window.addEventListener('beforeunload', () => this.saveProgress());
  }

  async saveProgress() {
    const position = Math.floor(this.videoElement.currentTime);
    const quality = this.getCurrentQuality(); // e.g., "720p"

    try {
      await fetch(`/api/history/${this.videoId}/progress`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`
        },
        body: JSON.stringify({ position, quality })
      });
    } catch (error) {
      console.error('Failed to save progress:', error);
    }
  }

  destroy() {
    if (this.progressInterval) {
      clearInterval(this.progressInterval);
    }
    this.saveProgress(); // Último guardado
  }
}

// Uso en página de video
const player = new VideoPlayer(videoId, initialPosition);
```

### JavaScript - Continuar Viendo

```javascript
// En la página principal o dashboard
async function loadContinueWatching() {
  try {
    const response = await fetch('/api/continue-watching?limit=10', {
      headers: {
        'Authorization': `Bearer ${getToken()}`
      }
    });

    const { data } = await response.json();

    // Renderizar sección "Continuar Viendo"
    renderContinueWatchingSection(data);
  } catch (error) {
    console.error('Failed to load continue watching:', error);
  }
}

function renderContinueWatchingSection(videos) {
  const container = document.getElementById('continue-watching');

  videos.forEach(item => {
    const progress = (item.position / item.video.duration) * 100;

    const card = `
      <div class="video-card">
        <img src="${item.video.thumbnail}" alt="${item.video.title}">
        <div class="progress-bar">
          <div class="progress" style="width: ${progress}%"></div>
        </div>
        <h3>${item.video.title}</h3>
        <p>Visto: ${formatTime(item.position)} / ${formatTime(item.video.duration)}</p>
        <a href="/video/${item.videoId}">Continuar viendo</a>
      </div>
    `;

    container.innerHTML += card;
  });
}
```

## Mejoras de Rendimiento

### Queries Optimizadas

**Favoritos:**
- `idx_favorites_user_id`: O(log n) para listar favoritos
- `unique(user_id, video_id)`: O(1) para verificar duplicados
- JOIN optimizado con videos (índices en ambas tablas)

**Historial:**
- `unique(user_id, video_id)`: UPSERT rápido (O(1))
- `idx_watch_history_user_id`: Listar historial O(log n)
- `idx_watch_history_watched_at`: Ordenar por fecha O(log n)

### Continue Watching Performance

```sql
-- Query altamente optimizada
WHERE wh.position > 10
  AND v.duration > 0
  AND (CAST(wh.position AS FLOAT) / CAST(v.duration AS FLOAT)) < 0.9
  AND v.status = 'ready'
  AND v.is_public = true
ORDER BY wh.watched_at DESC
LIMIT 10
```

**Índices usados:**
1. `idx_watch_history_user_id` → Filtrar por usuario
2. `idx_watch_history_watched_at` → Ordenar por fecha
3. JOIN usa `idx_videos_id` → O(1) por video

**Tiempo estimado:** ~5-10ms para 10,000 registros

## Beneficios Obtenidos

### 1. Persistencia Completa
- ✅ Favoritos permanentes (PostgreSQL)
- ✅ Historial sincronizado entre dispositivos
- ✅ No se pierde progreso al cerrar sesión
- ✅ Continue watching inteligente

### 2. Experiencia de Usuario
- ✅ "Continuar viendo" solo muestra videos relevantes
- ✅ Sincronización automática cada 30s
- ✅ Favoritos accesibles desde cualquier dispositivo
- ✅ Historial completo con paginación

### 3. Arquitectura Limpia
- ✅ Separación de responsabilidades (Hexagonal)
- ✅ Repositorios reutilizables
- ✅ Servicios sin lógica de persistencia
- ✅ Handlers enfocados en HTTP

## Testing Manual

### 1. Probar Favoritos

```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.token')

# Agregar favorito
curl -X POST http://localhost:8080/api/favorites/VIDEO_UUID \
  -H "Authorization: Bearer $TOKEN"

# Listar favoritos
curl http://localhost:8080/api/favorites \
  -H "Authorization: Bearer $TOKEN" | jq

# Quitar favorito
curl -X DELETE http://localhost:8080/api/favorites/VIDEO_UUID \
  -H "Authorization: Bearer $TOKEN"
```

### 2. Probar Historial

```bash
# Registrar progreso
curl -X POST http://localhost:8080/api/history/VIDEO_UUID/progress \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"position": 157, "quality": "720p"}'

# Ver "Continuar viendo"
curl http://localhost:8080/api/continue-watching?limit=5 \
  -H "Authorization: Bearer $TOKEN" | jq

# Ver historial completo
curl http://localhost:8080/api/history?page=1&limit=10 \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 3. Probar Sincronización

```bash
# Dispositivo 1: Guardar progreso
curl -X POST http://localhost:8080/api/history/VIDEO_UUID/progress \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"position": 300, "quality": "1080p"}'

# Dispositivo 2: Obtener progreso (mismo usuario)
curl http://localhost:8080/api/continue-watching \
  -H "Authorization: Bearer $TOKEN" | jq '.data[0].position'
# Debe mostrar: 300
```

## Migración de Datos

### Ejecutar Migración de Favoritos

```bash
# Si la tabla no existe
psql -d streaming_platform_db -f scripts/db/002_add_favorites_table.sql

# Verificar
psql -d streaming_platform_db -c "\\dt favorites"
```

### Migración de Datos Legacy (si aplica)

Si tenías favoritos en Redis/cache, migrar a PostgreSQL:

```sql
-- Script de migración manual (adaptar según necesidad)
INSERT INTO favorites (id, user_id, video_id, created_at)
SELECT
  uuid_generate_v4(),
  'USER_UUID'::uuid,
  'VIDEO_UUID'::uuid,
  NOW()
ON CONFLICT (user_id, video_id) DO NOTHING;
```

## Próximos Pasos (Fase 5)

### 1. Testing Automatizado
- [ ] Tests unitarios para UserService
- [ ] Tests unitarios para repositorios
- [ ] Tests de integración para endpoints HTTP
- [ ] Tests de concurrencia (múltiples dispositivos)

### 2. Optimizaciones
- [ ] Cache de favoritos (Redis) con invalidación
- [ ] Batch updates para watch progress (reducir writes)
- [ ] Índices compuestos para queries complejas
- [ ] Materialized view para analytics

### 3. Funcionalidades Adicionales
- [ ] Listas de reproducción personalizadas
- [ ] Compartir favoritos con otros usuarios
- [ ] Exportar historial de visualización
- [ ] Notificaciones de nuevos videos en favoritos

### 4. Analytics
- [ ] Dashboard de estadísticas de visualización
- [ ] Videos más populares (por favoritos)
- [ ] Tiempo promedio de visualización
- [ ] Tasa de finalización de videos

## Troubleshooting

### Error: "favorites table does not exist"
```bash
# Ejecutar migración
psql -d streaming_platform_db -f scripts/db/002_add_favorites_table.sql
```

### Error: "watch_history table does not exist"
```bash
# La tabla debe existir desde init_schema.sql
# Verificar
psql -d streaming_platform_db -c "\\dt watch_history"

# Si no existe, revisar init_schema.sql y ejecutar
psql -d streaming_platform_db -f scripts/db/init_schema.sql
```

### Progreso no se guarda
1. Verificar que el frontend está llamando UpdateWatchProgress
2. Ver logs del servidor para errores
3. Verificar autenticación (JWT válido)
4. Comprobar que video_id es válido

### Continue Watching vacío
- Verificar que position > 10s
- Verificar que video no está >90% completo
- Verificar que videos son públicos y ready
- Ver datos en PostgreSQL:
  ```sql
  SELECT * FROM watch_history WHERE user_id = 'UUID';
  ```

## Notas Técnicas

### UpdateWatchProgress - Idempotencia

```go
// ON CONFLICT asegura idempotencia
INSERT INTO watch_history (id, user_id, video_id, position, quality, watched_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (user_id, video_id)
DO UPDATE SET
    position = EXCLUDED.position,
    quality = EXCLUDED.quality,
    watched_at = NOW(),
    updated_at = NOW()
```

**Ventajas:**
- Múltiples llamadas con misma posición → 1 UPDATE
- Seguro para concurrencia
- No crea duplicados

### Sincronización - Eventual Consistency

```
T0: Dispositivo A → position 100
T1: Dispositivo B → GET continue-watching → position 100 ✅
T2: Dispositivo A → position 150
T3: Dispositivo B → GET continue-watching → position 150 ✅
```

No hay conflictos porque:
- UPSERT usa ON CONFLICT
- Último UPDATE gana
- watched_at siempre actualizado

---

**Estado**: ✅ **Fase 4 Completada - Favoritos e Historial Funcionales**

**Compilación**: ✅ Sin errores

**Endpoints**: 6 nuevos endpoints HTTP

**Próximo**: Fase 5 - Testing, Optimizaciones y Analytics
