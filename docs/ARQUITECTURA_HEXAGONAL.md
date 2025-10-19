# 📐 Arquitectura Hexagonal - Guía Completa

## 📚 Tabla de Contenidos

1. [¿Qué es la Arquitectura Hexagonal?](#qué-es-la-arquitectura-hexagonal)
2. [¿Por qué usar Arquitectura Hexagonal?](#por-qué-usar-arquitectura-hexagonal)
3. [Conceptos Fundamentales](#conceptos-fundamentales)
4. [Estructura del Proyecto](#estructura-del-proyecto)
5. [Flujo de Datos](#flujo-de-datos)
6. [Ejemplos Prácticos](#ejemplos-prácticos)
7. [Cómo Agregar Nuevas Funcionalidades](#cómo-agregar-nuevas-funcionalidades)
8. [Testing](#testing)
9. [Buenas Prácticas](#buenas-prácticas)

---

## 🎯 ¿Qué es la Arquitectura Hexagonal?

La **Arquitectura Hexagonal** (también conocida como **Ports & Adapters**) es un patrón de diseño que separa la **lógica de negocio** (el núcleo de tu aplicación) de los **detalles técnicos** (bases de datos, APIs, frameworks).

### Analogía Simple: Una Casa 🏠

Imagina tu aplicación como una **casa**:

- **El núcleo (core)** = Las habitaciones y funciones principales de la casa
- **Los puertos (ports)** = Las puertas y ventanas (interfaces de entrada/salida)
- **Los adapters** = Diferentes formas de acceder (puerta principal, puerta trasera, ventanas)

```
           🚪 HTTP API
              ↓
        ┌─────────────┐
        │   ADAPTER   │  ← Adapta HTTP a lo que el core entiende
        │   (HTTP)    │
        └─────────────┘
              ↓
        ┌─────────────┐
        │    PORT     │  ← Interfaz (contrato)
        │  (Service)  │
        └─────────────┘
              ↓
    ╔═══════════════════╗
    ║      CORE         ║  ← Tu lógica de negocio
    ║   (Servicios)     ║     (NO sabe nada de HTTP, SQL, etc.)
    ╚═══════════════════╝
              ↓
        ┌─────────────┐
        │    PORT     │  ← Interfaz (contrato)
        │(Repository) │
        └─────────────┘
              ↓
        ┌─────────────┐
        │   ADAPTER   │  ← Adapta el core a PostgreSQL
        │ (PostgreSQL)│
        └─────────────┘
              ↓
           🗄️ Database
```

---

## 💡 ¿Por qué usar Arquitectura Hexagonal?

### Problema sin Arquitectura Hexagonal ❌

```go
// ❌ MAL: El servicio depende directamente de la base de datos
func Login(email, password string) (string, error) {
    // Código mezclado con SQL, HTTP, lógica de negocio
    db.Query("SELECT * FROM users WHERE email = ?", email)
    // Si cambio de PostgreSQL a MongoDB, tengo que reescribir TODO
}
```

### Solución con Arquitectura Hexagonal ✅

```go
// ✅ BIEN: El servicio solo usa interfaces
type AuthService struct {
    userRepo UserRepository  // ← Interface, no implementación concreta
    cache    CacheRepository // ← Interface, no implementación concreta
}

func (s *AuthService) Login(email, password string) (string, error) {
    user := s.userRepo.GetByEmail(email)  // ← No sabe si es PostgreSQL, MongoDB, etc.
    // Lógica de negocio pura
}
```

### Beneficios 🎁

1. **Independencia de la base de datos**
   - Puedes cambiar PostgreSQL por MongoDB sin tocar la lógica de negocio

2. **Fácil de testear**
   - Puedes crear mocks de las interfaces para testing

3. **Código más limpio y organizado**
   - Cada parte tiene una responsabilidad clara

4. **Escalabilidad**
   - Puedes agregar nuevos adapters (GraphQL, gRPC) sin modificar el core

5. **Mantenibilidad**
   - Los cambios están aislados en capas específicas

---

## 🧩 Conceptos Fundamentales

### 1️⃣ **Core (Núcleo)**

El **corazón** de tu aplicación. Contiene:

#### **a) Domain (Dominio)**
Las **entidades** y **reglas de negocio**. Son objetos que representan conceptos del mundo real.

```go
// internal/core/domain/user.go

type User struct {
    ID           uuid.UUID
    Email        string
    PasswordHash string
    FirstName    string
    LastName     string
    Role         string
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// Reglas de negocio del dominio
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrUserInactive      = errors.New("user account is inactive")
)
```

**¿Qué va aquí?**
- ✅ Estructuras de datos (User, Video, etc.)
- ✅ Errores del dominio
- ✅ Constantes del negocio (VideoStatus, UserRole, etc.)
- ❌ NO va código de base de datos
- ❌ NO va código HTTP

#### **b) Ports (Puertos)**

Son **interfaces** (contratos). Definen **QUÉ** se puede hacer, pero NO **CÓMO** se hace.

**Input Ports** (Puertos de Entrada) - Lo que tu aplicación **PUEDE HACER**:

```go
// internal/core/ports/input/auth_service.go

type AuthService interface {
    Login(ctx context.Context, email, password string) (accessToken, refreshToken string, user *domain.UserProfile, error)
    Register(ctx context.Context, email, password, firstName, lastName, role string) (accessToken, refreshToken string, user *domain.UserProfile, error)
    Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error
}
```

**Output Ports** (Puertos de Salida) - Lo que tu aplicación **NECESITA**:

```go
// internal/core/ports/output/user_repository.go

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### **c) Services (Servicios)**

Implementan los **Input Ports** y contienen la **lógica de negocio**.

```go
// internal/core/services/auth_service.go

type authService struct {
    userRepo  output.UserRepository  // ← Usa el OUTPUT port (interfaz)
    cacheRepo output.CacheRepository // ← Usa el OUTPUT port (interfaz)
    jwtSecret string
}

// Implementa el INPUT port
func (s *authService) Login(ctx context.Context, email, password string) (string, string, *domain.UserProfile, error) {
    // 1. Validar email
    if err := validator.ValidateEmail(email); err != nil {
        return "", "", nil, domain.ErrAuthInvalidCredentials
    }

    // 2. Obtener usuario (usa el port, no sabe si es PostgreSQL o MongoDB)
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return "", "", nil, domain.ErrAuthInvalidCredentials
    }

    // 3. Verificar contraseña
    if !s.verifyPassword(password, user.PasswordHash) {
        return "", "", nil, domain.ErrAuthInvalidCredentials
    }

    // 4. Generar tokens
    accessToken, _ := s.generateAccessToken(user)
    refreshToken, _ := s.generateRefreshToken(user.ID)

    // 5. Retornar resultado
    return accessToken, refreshToken, profile, nil
}
```

**Características importantes:**
- ✅ Solo depende de **interfaces** (ports)
- ✅ Contiene la **lógica de negocio**
- ✅ Es **fácil de testear** (puedes crear mocks)
- ❌ NO tiene código SQL
- ❌ NO tiene código HTTP

### 2️⃣ **Adapters (Adaptadores)**

Implementan los **puertos** y se conectan con el mundo exterior.

#### **Input Adapters** (Adaptadores de Entrada)

Convierten peticiones externas al formato que el core entiende.

```go
// internal/adapters/input/http/auth_handler.go

type AuthHandler struct {
    authService input.AuthService  // ← Usa el INPUT port (interfaz)
    logger      logger.Logger
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1. Leer petición HTTP
    var req LoginRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Llamar al servicio del core (a través del port)
    accessToken, refreshToken, user, err := h.authService.Login(
        r.Context(),
        req.Email,
        req.Password,
    )

    // 3. Convertir respuesta a HTTP
    if err != nil {
        httputil.WriteError(w, http.StatusUnauthorized, err.Error())
        return
    }

    httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "user":          user,
    })
}
```

#### **Output Adapters** (Adaptadores de Salida)

Implementan los **Output Ports** y se conectan con infraestructura (BD, caché, etc.).

```go
// internal/adapters/output/persistence/postgres/user_repository.go

type userRepository struct {
    db *sql.DB  // ← Detalles técnicos (PostgreSQL)
}

// Implementa el OUTPUT port
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    query := `
        SELECT id, email, password_hash, first_name, last_name, role, is_active
        FROM users
        WHERE email = $1
    `

    user := &domain.User{}
    err := r.db.QueryRowContext(ctx, query, email).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.FirstName,
        &user.LastName,
        &user.Role,
        &user.IsActive,
    )

    if err == sql.ErrNoRows {
        return nil, domain.ErrUserNotFound
    }

    return user, err
}
```

---

## 📂 Estructura del Proyecto

```
streaming-platform/
│
├── cmd/
│   └── server/
│       └── main.go              # ← INYECCIÓN DE DEPENDENCIAS
│
├── internal/
│   │
│   ├── core/                    # ← NÚCLEO (no depende de nada externo)
│   │   ├── domain/              # Entidades y reglas de negocio
│   │   │   ├── user.go          # User, UserProfile, UserStats
│   │   │   ├── video.go         # Video, VideoSearchRequest
│   │   │   ├── video_file.go    # VideoFile, HLSPlaylist
│   │   │   └── errors.go        # Errores del dominio
│   │   │
│   │   ├── ports/               # Interfaces (contratos)
│   │   │   ├── input/           # Lo que la app PUEDE HACER
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── user_service.go
│   │   │   │   ├── video_service.go
│   │   │   │   └── streaming_service.go
│   │   │   │
│   │   │   └── output/          # Lo que la app NECESITA
│   │   │       ├── user_repository.go
│   │   │       ├── video_repository.go
│   │   │       └── cache_repository.go
│   │   │
│   │   └── services/            # Lógica de negocio
│   │       ├── auth_service.go
│   │       ├── user_service.go
│   │       ├── video_service.go
│   │       └── streaming_service.go
│   │
│   └── adapters/                # ← ADAPTADORES (implementan los ports)
│       │
│       ├── input/               # Adaptadores de ENTRADA
│       │   └── http/            # HTTP → Core
│       │       ├── auth_handler.go
│       │       ├── user_handler.go
│       │       ├── video_handler.go
│       │       └── streaming_handler.go
│       │
│       └── output/              # Adaptadores de SALIDA
│           └── persistence/     # Core → Infraestructura
│               ├── postgres/    # Core → PostgreSQL
│               │   ├── user_repository.go
│               │   └── video_repository.go
│               │
│               └── redis/       # Core → Redis
│                   └── cache_repository.go
│
├── pkg/                         # Utilidades compartidas
│   ├── httputil/               # Helpers HTTP
│   ├── validator/              # Validaciones
│   ├── dbutil/                 # Helpers DB
│   └── jwt/                    # JWT utilities
│
└── docs/
    └── ARQUITECTURA_HEXAGONAL.md  # ← Este archivo
```

---

## 🔄 Flujo de Datos

### Ejemplo: Usuario hace Login

```
1. HTTP Request (JSON)
   ↓
2. Input Adapter (HTTP Handler)
   - Convierte JSON a formato del core
   ↓
3. Input Port (AuthService interface)
   - Define el contrato
   ↓
4. Core Service (authService)
   - Ejecuta lógica de negocio
   - Valida email/password
   - Genera tokens
   ↓
5. Output Port (UserRepository interface)
   - Define qué necesita el core
   ↓
6. Output Adapter (PostgreSQL Repository)
   - Ejecuta SQL
   - Obtiene usuario de la BD
   ↓
7. Retorna resultado al Core
   ↓
8. Core retorna al Input Adapter
   ↓
9. Input Adapter convierte a HTTP Response
   ↓
10. HTTP Response (JSON)
```

### Diagrama Visual

```
┌─────────────────────────────────────────────────────────┐
│                    CLIENTE (Browser)                    │
└─────────────────────────────────────────────────────────┘
                           │
                    HTTP POST /login
                    { email, password }
                           │
                           ↓
┌─────────────────────────────────────────────────────────┐
│         INPUT ADAPTER (HTTP Handler)                    │
│  internal/adapters/input/http/auth_handler.go          │
│                                                          │
│  1. Lee JSON del request                                │
│  2. Valida formato                                      │
│  3. Llama al servicio                                   │
└─────────────────────────────────────────────────────────┘
                           │
                  Llama a: Login(email, password)
                           │
                           ↓
┌─────────────────────────────────────────────────────────┐
│         INPUT PORT (AuthService Interface)              │
│  internal/core/ports/input/auth_service.go             │
│                                                          │
│  type AuthService interface {                           │
│      Login(...) (string, string, *UserProfile, error)  │
│  }                                                       │
└─────────────────────────────────────────────────────────┘
                           │
                    Implementado por ↓
┌─────────────────────────────────────────────────────────┐
│              CORE SERVICE                               │
│  internal/core/services/auth_service.go                │
│                                                          │
│  1. Valida email format                                 │
│  2. Busca usuario (llama a userRepo)                   │
│  3. Verifica password                                   │
│  4. Genera tokens JWT                                   │
│  5. Retorna tokens                                      │
└─────────────────────────────────────────────────────────┘
                           │
                  Llama a: GetByEmail(email)
                           │
                           ↓
┌─────────────────────────────────────────────────────────┐
│       OUTPUT PORT (UserRepository Interface)            │
│  internal/core/ports/output/user_repository.go         │
│                                                          │
│  type UserRepository interface {                        │
│      GetByEmail(...) (*User, error)                    │
│  }                                                       │
└─────────────────────────────────────────────────────────┘
                           │
                    Implementado por ↓
┌─────────────────────────────────────────────────────────┐
│         OUTPUT ADAPTER (PostgreSQL Repo)                │
│  internal/adapters/output/persistence/postgres/        │
│  user_repository.go                                     │
│                                                          │
│  1. Ejecuta query SQL                                   │
│  2. SELECT * FROM users WHERE email = ?                │
│  3. Mapea resultado a domain.User                      │
│  4. Retorna User o error                               │
└─────────────────────────────────────────────────────────┘
                           │
                           ↓
                    ┌─────────────┐
                    │  PostgreSQL │
                    │   Database  │
                    └─────────────┘
```

---

## 🎓 Ejemplos Prácticos

### Ejemplo 1: Crear un nuevo video

```go
// 1. El usuario envía una petición HTTP
POST /api/videos
{
    "title": "Tutorial de Go",
    "description": "Aprende Go desde cero",
    "category": "programming"
}

// 2. El HTTP Handler recibe la petición
// internal/adapters/input/http/video_handler.go
func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
    var req CreateVideoRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Convierte a entidad del dominio
    video := &domain.Video{
        Title:       req.Title,
        Description: req.Description,
        Category:    req.Category,
    }

    // 3. Llama al servicio (a través del port)
    err := h.videoService.CreateVideo(r.Context(), video)

    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }

    httputil.WriteSuccess(w, http.StatusCreated, video)
}

// 4. El servicio ejecuta la lógica de negocio
// internal/core/services/video_service.go
func (s *videoService) CreateVideo(ctx context.Context, video *domain.Video) error {
    // Validaciones
    if err := validator.ValidateRequired("title", video.Title); err != nil {
        return err
    }

    // Establecer valores por defecto
    video.Status = domain.VideoStatusUploading
    video.ViewCount = 0

    // 5. Llama al repositorio (a través del port)
    return s.videoRepo.Create(ctx, video)
}

// 6. El repositorio persiste en PostgreSQL
// internal/adapters/output/persistence/postgres/video_repository.go
func (r *videoRepository) Create(ctx context.Context, video *domain.Video) error {
    query := `
        INSERT INTO videos (id, title, description, category, status, view_count)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    video.ID = uuid.New()

    _, err := r.db.ExecContext(ctx, query,
        video.ID,
        video.Title,
        video.Description,
        video.Category,
        video.Status,
        video.ViewCount,
    )

    return err
}
```

### Ejemplo 2: Testing con Mocks

La arquitectura hexagonal hace el testing muy fácil:

```go
// internal/core/services/auth_service_test.go

// 1. Crear un mock del UserRepository
type mockUserRepository struct {
    users map[string]*domain.User
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    user, exists := m.users[email]
    if !exists {
        return nil, domain.ErrUserNotFound
    }
    return user, nil
}

// 2. Testear el servicio sin base de datos real
func TestAuthService_Login(t *testing.T) {
    // Crear mocks
    mockRepo := &mockUserRepository{
        users: map[string]*domain.User{
            "test@example.com": {
                ID:           uuid.New(),
                Email:        "test@example.com",
                PasswordHash: "$2a$10$...", // Hash de "password123"
                IsActive:     true,
            },
        },
    }

    mockCache := &mockCacheRepository{}

    // Crear servicio con mocks
    service := services.NewAuthService(mockRepo, mockCache, "secret")

    // Ejecutar test
    token, _, user, err := service.Login(context.Background(), "test@example.com", "password123")

    // Verificar
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if token == "" {
        t.Error("Expected token, got empty string")
    }
    if user.Email != "test@example.com" {
        t.Errorf("Expected email test@example.com, got %s", user.Email)
    }
}
```

---

## 🚀 Cómo Agregar Nuevas Funcionalidades

### Caso: Agregar sistema de comentarios

#### Paso 1: Definir entidades del dominio

```go
// internal/core/domain/comment.go

type Comment struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    VideoID   uuid.UUID
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}

var (
    ErrCommentNotFound = errors.New("comment not found")
    ErrCommentTooLong  = errors.New("comment exceeds maximum length")
)
```

#### Paso 2: Crear el Input Port (interfaz del servicio)

```go
// internal/core/ports/input/comment_service.go

type CommentService interface {
    CreateComment(ctx context.Context, userID, videoID uuid.UUID, content string) (*domain.Comment, error)
    GetComments(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*domain.Comment, error)
    DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error
}
```

#### Paso 3: Crear el Output Port (interfaz del repositorio)

```go
// internal/core/ports/output/comment_repository.go

type CommentRepository interface {
    Create(ctx context.Context, comment *domain.Comment) error
    GetByVideoID(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*domain.Comment, error)
    Delete(ctx context.Context, id uuid.UUID) error
}
```

#### Paso 4: Implementar el servicio (lógica de negocio)

```go
// internal/core/services/comment_service.go

type commentService struct {
    commentRepo output.CommentRepository
    videoRepo   output.VideoRepository
}

func NewCommentService(commentRepo output.CommentRepository, videoRepo output.VideoRepository) input.CommentService {
    return &commentService{
        commentRepo: commentRepo,
        videoRepo:   videoRepo,
    }
}

func (s *commentService) CreateComment(ctx context.Context, userID, videoID uuid.UUID, content string) (*domain.Comment, error) {
    // Validaciones
    if len(content) > 500 {
        return nil, domain.ErrCommentTooLong
    }

    // Verificar que el video existe
    _, err := s.videoRepo.GetByID(ctx, videoID)
    if err != nil {
        return nil, err
    }

    // Crear comentario
    comment := &domain.Comment{
        ID:        uuid.New(),
        UserID:    userID,
        VideoID:   videoID,
        Content:   content,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // Persistir
    err = s.commentRepo.Create(ctx, comment)
    if err != nil {
        return nil, err
    }

    return comment, nil
}
```

#### Paso 5: Implementar el Output Adapter (PostgreSQL)

```go
// internal/adapters/output/persistence/postgres/comment_repository.go

type commentRepository struct {
    db *sql.DB
}

func NewCommentRepository(db *sql.DB) output.CommentRepository {
    return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *domain.Comment) error {
    query := `
        INSERT INTO comments (id, user_id, video_id, content, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    _, err := r.db.ExecContext(ctx, query,
        comment.ID,
        comment.UserID,
        comment.VideoID,
        comment.Content,
        comment.CreatedAt,
        comment.UpdatedAt,
    )

    return err
}
```

#### Paso 6: Crear el Input Adapter (HTTP Handler)

```go
// internal/adapters/input/http/comment_handler.go

type CommentHandler struct {
    commentService input.CommentService
    logger         logger.Logger
}

func NewCommentHandler(commentService input.CommentService, logger logger.Logger) *CommentHandler {
    return &CommentHandler{
        commentService: commentService,
        logger:         logger,
    }
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        VideoID string `json:"video_id"`
        Content string `json:"content"`
    }

    json.NewDecoder(r.Body).Decode(&req)

    userID := r.Context().Value("user_id").(uuid.UUID)
    videoID, _ := uuid.Parse(req.VideoID)

    comment, err := h.commentService.CreateComment(r.Context(), userID, videoID, req.Content)

    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, err.Error())
        return
    }

    httputil.WriteJSON(w, http.StatusCreated, comment)
}
```

#### Paso 7: Configurar en main.go

```go
// cmd/server/main.go

func main() {
    // ... código existente ...

    // Output Adapters
    commentRepo := postgresAdapter.NewCommentRepository(db)

    // Core Services
    commentService := services.NewCommentService(commentRepo, videoRepo)

    // Input Adapters
    commentHandler := httpHandlers.NewCommentHandler(commentService, log)

    // Routes
    router.HandleFunc("/api/videos/{id}/comments", commentHandler.CreateComment).Methods("POST")
    router.HandleFunc("/api/videos/{id}/comments", commentHandler.GetComments).Methods("GET")
}
```

---

## 🧪 Testing

### Estrategia de Testing por Capa

#### 1. Testing del Domain

```go
// internal/core/domain/user_test.go

func TestUserValidation(t *testing.T) {
    user := &domain.User{
        Email:     "invalid-email",
        FirstName: "",
    }

    // Test validaciones del dominio
    err := user.Validate()
    if err == nil {
        t.Error("Expected validation error")
    }
}
```

#### 2. Testing de Services (con mocks)

```go
// internal/core/services/video_service_test.go

type mockVideoRepository struct {
    videos map[uuid.UUID]*domain.Video
}

func (m *mockVideoRepository) Create(ctx context.Context, video *domain.Video) error {
    m.videos[video.ID] = video
    return nil
}

func TestVideoService_CreateVideo(t *testing.T) {
    mockRepo := &mockVideoRepository{videos: make(map[uuid.UUID]*domain.Video)}
    mockCache := &mockCacheRepository{}

    service := services.NewVideoService(mockRepo, mockCache)

    video := &domain.Video{
        Title:       "Test Video",
        Category:    "programming",
        InstructorID: uuid.New(),
    }

    err := service.CreateVideo(context.Background(), video)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if video.Status != domain.VideoStatusUploading {
        t.Errorf("Expected status %s, got %s", domain.VideoStatusUploading, video.Status)
    }
}
```

#### 3. Testing de Adapters (con base de datos de test)

```go
// internal/adapters/output/persistence/postgres/user_repository_test.go

func TestUserRepository_Create(t *testing.T) {
    // Usar base de datos de test
    db := setupTestDB(t)
    defer db.Close()

    repo := postgres.NewUserRepository(db)

    user := &domain.User{
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
    }

    err := repo.Create(context.Background(), user)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    // Verificar que se guardó
    retrieved, err := repo.GetByEmail(context.Background(), "test@example.com")
    if err != nil || retrieved == nil {
        t.Error("User not found after creation")
    }
}
```

#### 4. Testing de HTTP Handlers (integration tests)

```go
// internal/adapters/input/http/auth_handler_test.go

func TestAuthHandler_Login(t *testing.T) {
    // Setup
    mockService := &mockAuthService{}
    handler := NewAuthHandler(mockService, testLogger)

    // Crear request
    reqBody := `{"email":"test@example.com","password":"password123"}`
    req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
    w := httptest.NewRecorder()

    // Ejecutar
    handler.Login(w, req)

    // Verificar
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }

    var response map[string]interface{}
    json.NewDecoder(w.Body).Decode(&response)

    if response["access_token"] == nil {
        t.Error("Expected access_token in response")
    }
}
```

---

## ✨ Buenas Prácticas

### 1. El Core NUNCA debe importar código de Adapters

❌ **MAL:**
```go
// internal/core/services/auth_service.go
import "streaming-platform/internal/adapters/output/persistence/postgres"  // ❌ NO!

type authService struct {
    userRepo *postgres.UserRepository  // ❌ Dependencia concreta
}
```

✅ **BIEN:**
```go
// internal/core/services/auth_service.go
import "streaming-platform/internal/core/ports/output"  // ✅ SÍ!

type authService struct {
    userRepo output.UserRepository  // ✅ Interfaz
}
```

### 2. Los Adapters deben ser intercambiables

Deberías poder cambiar PostgreSQL por MongoDB sin tocar el core:

```go
// Antes: PostgreSQL
userRepo := postgresAdapter.NewUserRepository(db)

// Después: MongoDB (solo cambias esta línea)
userRepo := mongoAdapter.NewUserRepository(mongoClient)

// El servicio sigue funcionando igual
authService := services.NewAuthService(userRepo, cacheRepo, jwtSecret)
```

### 3. Las entidades del Domain deben ser "tontas"

```go
✅ BIEN:
type User struct {
    ID        uuid.UUID
    Email     string
    FirstName string
}

❌ MAL:
type User struct {
    ID        uuid.UUID
    Email     string
    db        *sql.DB        // ❌ No mezclar con infraestructura
    Save()    error          // ❌ No mezclar con persistencia
}
```

### 4. Un Port = Una responsabilidad

```go
✅ BIEN:
type UserRepository interface {
    Create(...)
    GetByID(...)
    Update(...)
}

type EmailService interface {
    SendWelcomeEmail(...)
    SendPasswordReset(...)
}

❌ MAL:
type UserRepository interface {
    Create(...)
    GetByID(...)
    SendEmail(...)  // ❌ Mezcla responsabilidades
}
```

### 5. Errores del dominio en el Core

```go
// internal/core/domain/errors.go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidEmail      = errors.New("invalid email format")
    ErrWeakPassword      = errors.New("password is too weak")
)

// Los servicios usan estos errores
func (s *authService) Register(...) error {
    if !isValidEmail(email) {
        return domain.ErrInvalidEmail  // ✅ Error del dominio
    }
}
```

### 6. Validaciones en el Service, no en el Handler

❌ **MAL:**
```go
// Handler
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    if len(req.Email) == 0 {  // ❌ Validación en el handler
        return
    }
    h.userService.CreateUser(...)
}
```

✅ **BIEN:**
```go
// Handler (solo convierte formatos)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    err := h.userService.CreateUser(req.Email, ...)
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, err.Error())
    }
}

// Service (contiene las validaciones)
func (s *userService) CreateUser(email string, ...) error {
    if err := validator.ValidateEmail(email); err != nil {  // ✅ Validación en el service
        return err
    }
    // ...
}
```

### 7. Inyección de Dependencias en main.go

Todo se conecta en `main.go`:

```go
func main() {
    // 1. Infraestructura
    db := database.NewPostgres(...)
    redis := database.NewRedis(...)

    // 2. Output Adapters (de afuera hacia adentro)
    userRepo := postgresAdapter.NewUserRepository(db)
    cacheRepo := redisAdapter.NewCacheRepository(redis)

    // 3. Core Services (inyectan dependencias)
    authService := services.NewAuthService(userRepo, cacheRepo, jwtSecret)
    userService := services.NewUserService(userRepo, cacheRepo)

    // 4. Input Adapters (usan los servicios)
    authHandler := httpHandlers.NewAuthHandler(authService, logger)
    userHandler := httpHandlers.NewUserHandler(userService, logger)

    // 5. Configurar rutas
    router.HandleFunc("/login", authHandler.Login)
    router.HandleFunc("/users", userHandler.GetUser)
}
```

---

## 📖 Resumen

### La Arquitectura Hexagonal en una frase:

> **"El núcleo de tu aplicación (core) solo depende de interfaces (ports), y los detalles técnicos (adapters) implementan esas interfaces."**

### Flujo mental al desarrollar:

1. **¿Qué necesito hacer?** → Define el **Input Port** (interfaz del servicio)
2. **¿Qué datos necesito?** → Define el **Output Port** (interfaz del repositorio)
3. **¿Cómo lo hago?** → Implementa el **Service** (lógica de negocio)
4. **¿De dónde vienen los datos?** → Implementa el **Output Adapter** (PostgreSQL, Redis, etc.)
5. **¿Cómo llega la petición?** → Implementa el **Input Adapter** (HTTP, GraphQL, CLI, etc.)

### Ventajas principales:

- ✅ **Independencia de frameworks** - Puedes cambiar de Gorilla Mux a Gin sin afectar el core
- ✅ **Independencia de base de datos** - Puedes cambiar PostgreSQL por MongoDB
- ✅ **Testeable** - Puedes testear la lógica sin base de datos real
- ✅ **Mantenible** - Cada capa tiene responsabilidades claras
- ✅ **Escalable** - Agregar funcionalidades es sistemático y predecible

---

## 🎓 Recursos Adicionales

- [Hexagonal Architecture (Alistair Cockburn)](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)

---

**¿Preguntas? Consulta los ejemplos en el código o revisa los tests para ver más casos de uso.**
