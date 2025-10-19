# 🚀 Guía Rápida - Arquitectura Hexagonal

## 📝 Cheat Sheet

### ¿Dónde poner mi código?

| Necesito... | Lo pongo en... | Ejemplo |
|------------|----------------|---------|
| Una estructura de datos del negocio | `internal/core/domain/` | `type User struct {...}` |
| Un error del negocio | `internal/core/domain/errors.go` | `var ErrUserNotFound = ...` |
| Definir QUÉ puede hacer la app | `internal/core/ports/input/` | `type UserService interface {...}` |
| Definir QUÉ necesita la app | `internal/core/ports/output/` | `type UserRepository interface {...}` |
| La lógica de negocio | `internal/core/services/` | `func (s *userService) CreateUser(...) {...}` |
| Código que maneja HTTP | `internal/adapters/input/http/` | `func (h *Handler) CreateUser(w, r) {...}` |
| Código que accede a PostgreSQL | `internal/adapters/output/persistence/postgres/` | `func (r *repo) Create(...) {...}` |
| Código que accede a Redis | `internal/adapters/output/persistence/redis/` | `func (r *cache) Set(...) {...}` |

---

## 🎯 Reglas de Oro

### ✅ PUEDES hacer:

1. **Core** puede usar **Ports**
   ```go
   // ✅ Service usa Repository port
   type userService struct {
       userRepo output.UserRepository  // Interfaz
   }
   ```

2. **Adapters** pueden implementar **Ports**
   ```go
   // ✅ PostgreSQL implementa el port
   func (r *userRepository) Create(...) error {
       // SQL aquí
   }
   ```

3. **Adapters** pueden usar **Domain**
   ```go
   // ✅ Handler retorna entidades del dominio
   func GetUser() *domain.User {
       return user
   }
   ```

### ❌ NO PUEDES hacer:

1. **Core** NO puede importar **Adapters**
   ```go
   // ❌ MAL
   import "internal/adapters/output/persistence/postgres"
   ```

2. **Domain** NO puede tener dependencias externas
   ```go
   // ❌ MAL
   type User struct {
       db *sql.DB  // No!
   }
   ```

3. **Ports** NO pueden tener implementaciones
   ```go
   // ❌ MAL
   type UserRepository interface {
       Create() error {  // ← No puede tener cuerpo
           db.Insert()
       }
   }
   ```

---

## 🔄 Flujo de Trabajo Típico

### Para agregar una nueva feature:

```
1. Domain      → Define las entidades (User, Video, etc.)
                 internal/core/domain/

2. Input Port  → Define QUÉ hace el servicio
                 internal/core/ports/input/

3. Output Port → Define QUÉ necesita el servicio
                 internal/core/ports/output/

4. Service     → Implementa la lógica de negocio
                 internal/core/services/

5. Output Adapter → Implementa acceso a BD
                    internal/adapters/output/persistence/

6. Input Adapter  → Implementa HTTP handler
                    internal/adapters/input/http/

7. main.go     → Conecta todo con inyección de dependencias
```

---

## 📋 Plantillas de Código

### 1. Nueva Entidad del Dominio

```go
// internal/core/domain/nombre.go

package domain

import (
    "time"
    "github.com/google/uuid"
)

type NombreEntidad struct {
    ID        uuid.UUID `json:"id"`
    Campo1    string    `json:"campo1"`
    Campo2    int       `json:"campo2"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Errores específicos del dominio
var (
    ErrNombreNotFound = errors.New("nombre not found")
    ErrNombreInvalid  = errors.New("nombre is invalid")
)
```

### 2. Nuevo Input Port (Servicio)

```go
// internal/core/ports/input/nombre_service.go

package input

import (
    "context"
    "streaming-platform/internal/core/domain"
    "github.com/google/uuid"
)

type NombreService interface {
    Create(ctx context.Context, entity *domain.NombreEntidad) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.NombreEntidad, error)
    Update(ctx context.Context, entity *domain.NombreEntidad) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, page, limit int) ([]*domain.NombreEntidad, int64, error)
}
```

### 3. Nuevo Output Port (Repositorio)

```go
// internal/core/ports/output/nombre_repository.go

package output

import (
    "context"
    "streaming-platform/internal/core/domain"
    "github.com/google/uuid"
)

type NombreRepository interface {
    Create(ctx context.Context, entity *domain.NombreEntidad) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.NombreEntidad, error)
    Update(ctx context.Context, entity *domain.NombreEntidad) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, page, limit int) ([]*domain.NombreEntidad, int64, error)
}
```

### 4. Implementación del Servicio

```go
// internal/core/services/nombre_service.go

package services

import (
    "context"
    "fmt"

    "streaming-platform/internal/core/domain"
    "streaming-platform/internal/core/ports/input"
    "streaming-platform/internal/core/ports/output"
    "streaming-platform/pkg/validator"

    "github.com/google/uuid"
)

type nombreService struct {
    nombreRepo output.NombreRepository
    cacheRepo  output.CacheRepository
}

func NewNombreService(nombreRepo output.NombreRepository, cacheRepo output.CacheRepository) input.NombreService {
    return &nombreService{
        nombreRepo: nombreRepo,
        cacheRepo:  cacheRepo,
    }
}

func (s *nombreService) Create(ctx context.Context, entity *domain.NombreEntidad) error {
    // 1. Validaciones
    if err := validator.ValidateRequired("campo1", entity.Campo1); err != nil {
        return err
    }

    // 2. Lógica de negocio
    entity.ID = uuid.New()

    // 3. Persistir
    if err := s.nombreRepo.Create(ctx, entity); err != nil {
        return fmt.Errorf("failed to create: %w", err)
    }

    // 4. Cachear (opcional)
    _ = s.cacheRepo.Set(ctx, fmt.Sprintf("nombre:%s", entity.ID), entity, 30*time.Minute)

    return nil
}

func (s *nombreService) GetByID(ctx context.Context, id uuid.UUID) (*domain.NombreEntidad, error) {
    // 1. Verificar caché
    var entity domain.NombreEntidad
    if err := s.cacheRepo.Get(ctx, fmt.Sprintf("nombre:%s", id), &entity); err == nil {
        return &entity, nil
    }

    // 2. Obtener de BD
    result, err := s.nombreRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Cachear
    _ = s.cacheRepo.Set(ctx, fmt.Sprintf("nombre:%s", id), result, 30*time.Minute)

    return result, nil
}
```

### 5. Implementación del Output Adapter (PostgreSQL)

```go
// internal/adapters/output/persistence/postgres/nombre_repository.go

package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "streaming-platform/internal/core/domain"
    "streaming-platform/internal/core/ports/output"
    "streaming-platform/pkg/dbutil"

    "github.com/google/uuid"
)

type nombreRepository struct {
    db *sql.DB
}

func NewNombreRepository(db *sql.DB) output.NombreRepository {
    return &nombreRepository{db: db}
}

func (r *nombreRepository) Create(ctx context.Context, entity *domain.NombreEntidad) error {
    query := `
        INSERT INTO nombre_tabla (id, campo1, campo2, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `

    entity.CreatedAt = time.Now()
    entity.UpdatedAt = time.Now()

    _, err := r.db.ExecContext(ctx, query,
        entity.ID,
        entity.Campo1,
        entity.Campo2,
        entity.CreatedAt,
        entity.UpdatedAt,
    )

    if err != nil {
        return fmt.Errorf("failed to insert: %w", err)
    }

    return nil
}

func (r *nombreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NombreEntidad, error) {
    query := `
        SELECT id, campo1, campo2, created_at, updated_at
        FROM nombre_tabla
        WHERE id = $1
    `

    entity := &domain.NombreEntidad{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &entity.ID,
        &entity.Campo1,
        &entity.Campo2,
        &entity.CreatedAt,
        &entity.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        return nil, domain.ErrNombreNotFound
    }

    if err != nil {
        return nil, fmt.Errorf("failed to query: %w", err)
    }

    return entity, nil
}
```

### 6. Implementación del Input Adapter (HTTP Handler)

```go
// internal/adapters/input/http/nombre_handler.go

package http

import (
    "encoding/json"
    "net/http"

    "streaming-platform/internal/core/domain"
    "streaming-platform/internal/core/ports/input"
    "streaming-platform/pkg/httputil"
    "streaming-platform/pkg/logger"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
)

type NombreHandler struct {
    nombreService input.NombreService
    logger        logger.Logger
}

func NewNombreHandler(nombreService input.NombreService, logger logger.Logger) *NombreHandler {
    return &NombreHandler{
        nombreService: nombreService,
        logger:        logger,
    }
}

// POST /api/nombres
func (h *NombreHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Campo1 string `json:"campo1"`
        Campo2 int    `json:"campo2"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    entity := &domain.NombreEntidad{
        Campo1: req.Campo1,
        Campo2: req.Campo2,
    }

    if err := h.nombreService.Create(r.Context(), entity); err != nil {
        httputil.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }

    httputil.WriteJSON(w, http.StatusCreated, entity)
}

// GET /api/nombres/{id}
func (h *NombreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := uuid.Parse(vars["id"])
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "Invalid ID format")
        return
    }

    entity, err := h.nombreService.GetByID(r.Context(), id)
    if err != nil {
        if err == domain.ErrNombreNotFound {
            httputil.WriteNotFound(w, "Nombre not found")
            return
        }
        httputil.WriteInternalError(w, "Failed to get nombre")
        return
    }

    httputil.WriteJSON(w, http.StatusOK, entity)
}

// Registrar rutas
func (h *NombreHandler) RegisterRoutes(router *mux.Router) {
    router.HandleFunc("/api/nombres", h.Create).Methods("POST")
    router.HandleFunc("/api/nombres/{id}", h.GetByID).Methods("GET")
}
```

### 7. Conectar en main.go

```go
// cmd/server/main.go

func main() {
    // ... código existente ...

    // Output Adapters
    nombreRepo := postgresAdapter.NewNombreRepository(db)

    // Core Services
    nombreService := services.NewNombreService(nombreRepo, cacheRepo)

    // Input Adapters
    nombreHandler := httpHandlers.NewNombreHandler(nombreService, log)

    // Routes
    nombreHandler.RegisterRoutes(router)

    // ... resto del código ...
}
```

---

## 🧪 Template de Test

```go
// internal/core/services/nombre_service_test.go

package services

import (
    "context"
    "testing"

    "streaming-platform/internal/core/domain"

    "github.com/google/uuid"
)

// Mock Repository
type mockNombreRepository struct {
    entities map[uuid.UUID]*domain.NombreEntidad
}

func (m *mockNombreRepository) Create(ctx context.Context, entity *domain.NombreEntidad) error {
    m.entities[entity.ID] = entity
    return nil
}

func (m *mockNombreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NombreEntidad, error) {
    entity, exists := m.entities[id]
    if !exists {
        return nil, domain.ErrNombreNotFound
    }
    return entity, nil
}

// Mock Cache
type mockCacheRepository struct{}

func (m *mockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
    return nil
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
    return nil
}

// Test
func TestNombreService_Create(t *testing.T) {
    // Setup
    mockRepo := &mockNombreRepository{
        entities: make(map[uuid.UUID]*domain.NombreEntidad),
    }
    mockCache := &mockCacheRepository{}

    service := NewNombreService(mockRepo, mockCache)

    // Execute
    entity := &domain.NombreEntidad{
        Campo1: "test",
        Campo2: 123,
    }

    err := service.Create(context.Background(), entity)

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if entity.ID == uuid.Nil {
        t.Error("Expected ID to be set")
    }

    // Verify it was saved
    saved, err := mockRepo.GetByID(context.Background(), entity.ID)
    if err != nil || saved == nil {
        t.Error("Entity was not saved to repository")
    }
}
```

---

## 🎯 Checklist para Nueva Feature

- [ ] 1. Crear entidad en `internal/core/domain/`
- [ ] 2. Crear errores específicos en `internal/core/domain/errors.go`
- [ ] 3. Definir Input Port en `internal/core/ports/input/`
- [ ] 4. Definir Output Port en `internal/core/ports/output/`
- [ ] 5. Implementar servicio en `internal/core/services/`
- [ ] 6. Implementar repositorio PostgreSQL en `internal/adapters/output/persistence/postgres/`
- [ ] 7. Implementar handler HTTP en `internal/adapters/input/http/`
- [ ] 8. Conectar en `main.go` (output adapter → service → input adapter)
- [ ] 9. Escribir tests unitarios del servicio
- [ ] 10. Escribir tests de integración del repositorio
- [ ] 11. Probar endpoints HTTP

---

## 📞 Comandos Útiles

```bash
# Compilar todo
go build ./...

# Compilar solo el core
go build ./internal/core/...

# Compilar solo los adapters
go build ./internal/adapters/...

# Compilar el servidor
go build ./cmd/server

# Ejecutar tests
go test ./...

# Ejecutar tests del core
go test ./internal/core/...

# Ejecutar tests con cobertura
go test -cover ./...

# Ver estructura de directorios
tree internal/core internal/adapters
```

---

## 💡 FAQs

### ¿Puedo poner validaciones en el Handler?

**No**. Las validaciones de negocio van en el **Service**. El Handler solo valida formato HTTP.

### ¿Dónde pongo código de email?

Crea un **Output Port** llamado `EmailService` y un adapter que lo implemente.

### ¿Puedo usar el dominio en el Handler?

**Sí**. Los Handlers pueden importar y usar entidades del `domain`.

### ¿Qué pasa si necesito cambiar de base de datos?

Solo cambias el **Output Adapter**. El core sigue igual.

### ¿Cómo testeo sin base de datos?

Creas un **mock** del repositorio que implementa la interfaz.

---

**Esta es tu referencia rápida. Para más detalles, consulta [ARQUITECTURA_HEXAGONAL.md](./ARQUITECTURA_HEXAGONAL.md)**
