# 📊 Diagrama de Arquitectura - Streaming Platform

## 🏗️ Vista General de la Arquitectura Hexagonal

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLIENTES EXTERNOS                                  │
│                    (Navegadores, Apps Móviles, APIs)                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INPUT ADAPTERS (Puertos de Entrada)                 │
│                         internal/adapters/input/http/                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐             │
│  │  AuthHandler    │  │  UserHandler    │  │  VideoHandler   │             │
│  │                 │  │                 │  │                 │             │
│  │ - Login()       │  │ - GetProfile()  │  │ - UploadVideo() │             │
│  │ - Register()    │  │ - UpdateProfile │  │ - GetVideo()    │             │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘             │
│                                                                              │
│  ┌────────────────────────────┐                                             │
│  │  StreamingHandler          │                                             │
│  │                            │                                             │
│  │ - GetMasterPlaylist()      │                                             │
│  │ - GetPlaylist()            │                                             │
│  │ - GetSegment()             │                                             │
│  └────────────────────────────┘                                             │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INPUT PORTS (Interfaces)                            │
│                         internal/core/ports/input/                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  interface AuthService         interface UserService                        │
│  interface VideoService        interface StreamingService                   │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CORE - SERVICIOS                                  │
│                         internal/core/services/                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                        LÓGICA DE NEGOCIO                             │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │                                                                      │   │
│  │  authService:                                                        │   │
│  │  • Validación de credenciales                                        │   │
│  │  • Generación de tokens JWT                                          │   │
│  │  • Gestión de sesiones                                               │   │
│  │                                                                      │   │
│  │  userService:                                                        │   │
│  │  • Gestión de perfiles                                               │   │
│  │  • Actualización de usuarios                                         │   │
│  │  • Validaciones de negocio                                           │   │
│  │                                                                      │   │
│  │  videoService:                                                       │   │
│  │  • Carga y procesamiento de videos                                   │   │
│  │  • Gestión de metadata                                               │   │
│  │  • Validaciones de formato                                           │   │
│  │                                                                      │   │
│  │  streamingService:                                                   │   │
│  │  • Generación de playlists HLS                                       │   │
│  │  • Control de calidad adaptativo                                     │   │
│  │  • Gestión de tokens de streaming                                    │   │
│  │                                                                      │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CORE - DOMINIO                                    │
│                         internal/core/domain/                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                      │
│  │    User      │  │    Video     │  │  VideoFile   │                      │
│  ├──────────────┤  ├──────────────┤  ├──────────────┤                      │
│  │ ID           │  │ ID           │  │ ID           │                      │
│  │ Email        │  │ Title        │  │ Quality      │                      │
│  │ Username     │  │ Description  │  │ Resolution   │                      │
│  │ Password     │  │ UserID       │  │ Bitrate      │                      │
│  │ Role         │  │ Duration     │  │ FilePath     │                      │
│  │ CreatedAt    │  │ FileSize     │  │ VideoID      │                      │
│  └──────────────┘  │ Status       │  └──────────────┘                      │
│                    │ Visibility   │                                         │
│  Errores del       └──────────────┘                                         │
│  Dominio:                                                                    │
│  • ErrInvalidCredentials                                                     │
│  • ErrUserNotFound                                                           │
│  • ErrVideoNotFound                                                          │
│  • ErrUnauthorized                                                           │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         OUTPUT PORTS (Interfaces)                           │
│                         internal/core/ports/output/                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  interface UserRepository      interface VideoRepository                    │
│  interface CacheRepository                                                  │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       OUTPUT ADAPTERS (Implementaciones)                    │
│                       internal/adapters/output/persistence/                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────┐         ┌──────────────────────────┐          │
│  │   PostgreSQL Adapters    │         │     Redis Adapter        │          │
│  ├──────────────────────────┤         ├──────────────────────────┤          │
│  │                          │         │                          │          │
│  │ • userRepository         │         │ • cacheRepository        │          │
│  │   - Create()             │         │   - Set()                │          │
│  │   - GetByID()            │         │   - Get()                │          │
│  │   - GetByEmail()         │         │   - Delete()             │          │
│  │   - Update()             │         │   - CacheUser()          │          │
│  │   - Delete()             │         │   - CacheVideo()         │          │
│  │   - List()               │         │   - SetSession()         │          │
│  │                          │         │                          │          │
│  │ • videoRepository        │         └──────────────────────────┘          │
│  │   - Create()             │                                                │
│  │   - GetByID()            │                                                │
│  │   - GetByUserID()        │                                                │
│  │   - Update()             │                                                │
│  │   - Delete()             │                                                │
│  │   - List()               │                                                │
│  │   - UpdateStatus()       │                                                │
│  │                          │                                                │
│  └──────────────────────────┘                                                │
│                                                                              │
└───────────────────────────────┬─────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         INFRAESTRUCTURA EXTERNA                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────┐         ┌─────────────────┐                            │
│  │   PostgreSQL    │         │      Redis      │                            │
│  │   (Database)    │         │     (Cache)     │                            │
│  └─────────────────┘         └─────────────────┘                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## 🔄 Flujo de una Petición (Ejemplo: Login)

```
1. Cliente HTTP
     │
     │ POST /api/auth/login
     ▼
2. AuthHandler.Login()
     │ (Input Adapter)
     │ - Decodifica JSON
     │ - Valida formato HTTP
     ▼
3. AuthService.Login()
     │ (Core Service)
     │ - Valida email
     │ - Obtiene usuario
     │ - Verifica password
     │ - Genera tokens JWT
     ▼
4. UserRepository.GetByEmail()
     │ (Output Port → PostgreSQL Adapter)
     │ - Ejecuta query SQL
     │ - Mapea resultado a domain.User
     ▼
5. PostgreSQL Database
     │
     │ Retorna datos
     ▼
6. domain.User ← UserRepository
     │
     │ Retorna entidad del dominio
     ▼
7. AuthService
     │ - Verifica password
     │ - Genera JWT tokens
     │ - Cachea sesión en Redis
     ▼
8. AuthHandler
     │ - Crea response JSON
     │ - Setea cookies
     ▼
9. Cliente HTTP
     Recibe: { accessToken, refreshToken, user }
```

## 📦 Dependencias del Proyecto

```
┌─────────────────────────────────────────────────────────────────┐
│                         cmd/server                              │
│                         (main.go)                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Inicializa:                                                     │
│  1. Configuración (pkg/config)                                   │
│  2. Logger (pkg/logger)                                          │
│  3. Bases de datos (pkg/database)                                │
│  4. Output Adapters (postgres, redis)                            │
│  5. Core Services (con inyección de dependencias)                │
│  6. Input Adapters (HTTP handlers)                               │
│  7. Router y Middleware                                          │
│  8. Worker Pool                                                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## 🛠️ Paquetes de Utilidades (pkg/)

```
pkg/
├── config/          → Configuración de la aplicación
├── database/        → Conexiones a PostgreSQL y Redis
├── dbutil/          → Helpers para operaciones de BD
├── httputil/        → Helpers para HTTP (responses, cookies)
├── jwt/             → Generación y validación de JWT
├── logger/          → Sistema de logging
├── utils/           → Utilidades generales
└── validator/       → Validaciones (email, required, etc.)
```

## 🎯 Inyección de Dependencias (main.go)

```go
// 1. OUTPUT ADAPTERS (Implementaciones concretas)
userRepo := postgresAdapter.NewUserRepository(db)
videoRepo := postgresAdapter.NewVideoRepository(db)
cacheRepo := redisAdapter.NewCacheRepository(redisClient)

// 2. CORE SERVICES (Reciben interfaces, no implementaciones)
authService := services.NewAuthService(userRepo, cacheRepo, cfg.JWTSecret)
userService := services.NewUserService(userRepo, cacheRepo)
videoService := services.NewVideoService(videoRepo, cacheRepo)
streamingService := services.NewStreamingService(videoRepo, cacheRepo, cfg.CDNBaseURL, cfg.JWTSecret)

// 3. INPUT ADAPTERS (Reciben servicios del core)
authHandler := httpHandlers.NewAuthHandler(authService, log)
userHandler := httpHandlers.NewUserHandler(userService, log)
videoHandler := httpHandlers.NewVideoHandler(videoService, log)
streamingHandler := httpHandlers.NewStreamingHandler(streamingService, log)
```

## 🔒 Separación de Responsabilidades

### ✅ Core (Corazón de la aplicación)
- **NO depende** de frameworks
- **NO depende** de bases de datos específicas
- **NO depende** de HTTP
- **SOLO** contiene lógica de negocio pura

### ✅ Adapters (Detalles de implementación)
- **Dependen** del Core (importan ports y domain)
- **Implementan** las interfaces (ports)
- **Pueden ser reemplazados** sin afectar el core

### ✅ Main (Composición)
- **Conecta** todos los componentes
- **Inyecta** dependencias
- **Configura** la aplicación

## 📊 Matriz de Dependencias

```
┌──────────────┬─────────┬──────────┬──────────┬─────────┐
│   Capa       │ Domain  │  Ports   │ Services │ Adapters│
├──────────────┼─────────┼──────────┼──────────┼─────────┤
│ Domain       │    -    │    ✗     │    ✗     │    ✗    │
│ Ports        │    ✓    │    -     │    ✗     │    ✗    │
│ Services     │    ✓    │    ✓     │    -     │    ✗    │
│ Adapters     │    ✓    │    ✓     │    ✗     │    -    │
│ Main         │    ✓    │    ✓     │    ✓     │    ✓    │
└──────────────┴─────────┴──────────┴──────────┴─────────┘

✓ = Puede importar
✗ = NO puede importar
```

## 🧪 Estrategia de Testing

```
┌─────────────────────────────────────────────────────────┐
│              TESTS UNITARIOS (Services)                 │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  • Usan MOCKS de los repositorios                        │
│  • Prueban LÓGICA DE NEGOCIO aislada                     │
│  • NO requieren base de datos                            │
│  • Rápidos y confiables                                  │
│                                                          │
│  Ejemplo:                                                │
│  type mockUserRepository struct { ... }                  │
│  func TestAuthService_Login(t *testing.T) { ... }        │
│                                                          │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│           TESTS DE INTEGRACIÓN (Repositories)           │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  • Usan base de datos de testing                        │
│  • Prueban queries SQL reales                            │
│  • Verifican mapeo de datos                              │
│                                                          │
│  Ejemplo:                                                │
│  func TestUserRepository_Create(t *testing.T) {          │
│      db := setupTestDB()                                 │
│      repo := NewUserRepository(db)                       │
│      ...                                                 │
│  }                                                       │
│                                                          │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│              TESTS E2E (End-to-End)                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  • Prueban flujo completo HTTP → DB → HTTP              │
│  • Usan servidor de testing                             │
│  • Verifican comportamiento real                         │
│                                                          │
│  Ejemplo:                                                │
│  func TestLoginEndpoint(t *testing.T) {                  │
│      server := setupTestServer()                         │
│      resp := httptest.Post("/api/auth/login", ...)       │
│      ...                                                 │
│  }                                                       │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 📈 Beneficios de esta Arquitectura

### 1. ✅ Testabilidad
```go
// Fácil de testear porque usamos interfaces
mockRepo := &mockUserRepository{}
service := NewAuthService(mockRepo, mockCache, "secret")
// Ahora puedo probar el servicio sin BD real
```

### 2. ✅ Mantenibilidad
```
Si cambio de PostgreSQL a MongoDB:
- ❌ NO toco el Core
- ❌ NO toco los Handlers
- ✅ SOLO creo un nuevo adapter MongoDB
```

### 3. ✅ Escalabilidad
```
Puedo agregar nuevos adapters:
- HTTP Handler → gRPC Handler
- PostgreSQL → MongoDB
- Redis → Memcached
Sin tocar la lógica de negocio
```

### 4. ✅ Independencia
```
El Core NO sabe:
- Si usa PostgreSQL o MySQL
- Si es HTTP o gRPC
- Si usa Redis o Memcached
```

## 🎓 Resumen para Recordar

```
INPUT ADAPTERS (HTTP)
      ↓
INPUT PORTS (Interfaces de servicios)
      ↓
SERVICES (Lógica de negocio)
      ↓
OUTPUT PORTS (Interfaces de repos)
      ↓
OUTPUT ADAPTERS (PostgreSQL/Redis)
      ↓
INFRAESTRUCTURA (Bases de datos)
```

**Regla de Oro:** Las dependencias apuntan HACIA ADENTRO (hacia el Core)

---

**Para más detalles, consulta:**
- [ARQUITECTURA_HEXAGONAL.md](./ARQUITECTURA_HEXAGONAL.md) - Guía completa
- [GUIA_RAPIDA.md](./GUIA_RAPIDA.md) - Referencia rápida
- [EJEMPLO_COMPLETO.md](./EJEMPLO_COMPLETO.md) - Ejemplo paso a paso
