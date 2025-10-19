# 🎓 Ejemplo Completo: Login de Usuario

Este documento muestra el flujo COMPLETO de una petición de login, desde que el usuario envía la petición HTTP hasta que recibe la respuesta.

---

## 📋 Índice

1. [Vista General](#vista-general)
2. [Paso a Paso Detallado](#paso-a-paso-detallado)
3. [Código Completo](#código-completo)
4. [Diagrama de Secuencia](#diagrama-de-secuencia)
5. [¿Qué pasa si...?](#qué-pasa-si)

---

## 🎯 Vista General

**Petición del Usuario:**
```bash
POST http://localhost:8080/api/auth/login
Content-Type: application/json

{
  "email": "miguel@example.com",
  "password": "MiPassword123!"
}
```

**Flujo Simplificado:**
```
Usuario → HTTP Handler → Auth Service → User Repository → PostgreSQL
                ↓            ↓              ↓
              JSON      Lógica de      Query SQL
                        Negocio
```

---

## 📝 Paso a Paso Detallado

### Paso 1: Petición HTTP llega al servidor

```
POST /api/auth/login HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{
  "email": "miguel@example.com",
  "password": "MiPassword123!"
}
```

**¿Quién la recibe?**
- El router de Gorilla Mux
- Está configurado en `main.go`

```go
// cmd/server/main.go
router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
```

---

### Paso 2: HTTP Handler procesa la petición

**Archivo:** `internal/adapters/input/http/auth_handler.go`

```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // 1️⃣ Decodificar el JSON de la petición
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "Invalid request format")
        return
    }
    // req.Email = "miguel@example.com"
    // req.Password = "MiPassword123!"

    // 2️⃣ Llamar al servicio del CORE (a través del port)
    accessToken, refreshToken, user, err := h.authService.Login(
        r.Context(),
        req.Email,
        req.Password,
    )

    // 3️⃣ Manejar errores
    if err != nil {
        h.handleAuthError(w, err)  // Convierte error del dominio a HTTP
        return
    }

    // 4️⃣ Configurar cookie con refresh token
    httputil.SetRefreshTokenCookie(w, refreshToken)

    // 5️⃣ Retornar respuesta JSON
    httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
        "access_token": accessToken,
        "user": user,
    })
}
```

**Responsabilidades del Handler:**
- ✅ Leer y parsear JSON
- ✅ Llamar al servicio
- ✅ Convertir errores del dominio a códigos HTTP
- ✅ Generar respuesta HTTP
- ❌ NO tiene lógica de negocio
- ❌ NO accede a la base de datos

---

### Paso 3: Auth Service ejecuta la lógica de negocio

**Archivo:** `internal/core/services/auth_service.go`

```go
func (s *authService) Login(ctx context.Context, email, password string) (string, string, *domain.UserProfile, error) {

    // 1️⃣ VALIDAR EMAIL
    if err := validator.ValidateEmail(email); err != nil {
        return "", "", nil, domain.ErrInvalidCredentials
    }
    // ✅ Email válido

    // 2️⃣ OBTENER USUARIO (usa el Output Port)
    user, err := s.userRepo.GetByEmail(ctx, email)
    // Llamada al puerto → el servicio NO sabe si es PostgreSQL, MongoDB, etc.

    if err != nil {
        if err == domain.ErrUserNotFound {
            return "", "", nil, domain.ErrInvalidCredentials
            // 🔒 Seguridad: No revelar si el email existe
        }
        return "", "", nil, fmt.Errorf("failed to get user: %w", err)
    }
    // user = { ID: "123", Email: "miguel@example.com", PasswordHash: "$2a$10$..." }

    // 3️⃣ VERIFICAR QUE EL USUARIO ESTÉ ACTIVO
    if !user.IsActive {
        return "", "", nil, domain.ErrUserInactive
    }
    // ✅ Usuario activo

    // 4️⃣ VERIFICAR CONTRASEÑA
    if !s.verifyPassword(password, user.PasswordHash) {
        return "", "", nil, domain.ErrInvalidCredentials
    }
    // Compara: bcrypt.CompareHashAndPassword(hash, password)
    // ✅ Contraseña correcta

    // 5️⃣ GENERAR ACCESS TOKEN (JWT)
    accessToken, err := s.generateAccessToken(user)
    if err != nil {
        return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
    }
    // accessToken = "eyJhbGciOiJIUzI1NiIs..."
    // Contiene: userID, email, role
    // Expira en: 15 minutos

    // 6️⃣ GENERAR REFRESH TOKEN
    refreshToken, err := s.generateRefreshToken(user.ID)
    if err != nil {
        return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
    }
    // refreshToken = "a7f3c9e2..."
    // Se guarda en Redis
    // Expira en: 7 días

    // 7️⃣ ACTUALIZAR ÚLTIMA FECHA DE LOGIN
    if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
        // ⚠️ Log pero no falla el login
        fmt.Printf("Warning: failed to update last login: %v\n", err)
    }

    // 8️⃣ CREAR PERFIL DE USUARIO PARA LA RESPUESTA
    profile := &domain.UserProfile{
        ID:        user.ID,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Role:      user.Role,
        Avatar:    user.Avatar,
        CreatedAt: user.CreatedAt,
    }
    // ℹ️ UserProfile NO incluye password_hash (seguridad)

    // 9️⃣ CACHEAR USUARIO EN REDIS
    if err := s.cacheRepo.CacheUser(ctx, user); err != nil {
        // ⚠️ Log pero no falla el login
        fmt.Printf("Warning: failed to cache user: %v\n", err)
    }

    // 🔟 RETORNAR TOKENS Y PERFIL
    return accessToken, refreshToken, profile, nil
}
```

**Responsabilidades del Service:**
- ✅ Validaciones de negocio
- ✅ Lógica de autenticación
- ✅ Generación de tokens
- ✅ Orquestación de llamadas a repositorios
- ❌ NO tiene código SQL
- ❌ NO tiene código HTTP

---

### Paso 4: User Repository obtiene el usuario

**Archivo:** `internal/adapters/output/persistence/postgres/user_repository.go`

```go
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {

    // 1️⃣ DEFINIR QUERY SQL
    query := `
        SELECT id, email, password_hash, first_name, last_name,
               role, avatar, is_active, created_at, updated_at
        FROM users
        WHERE email = $1
    `

    // 2️⃣ CREAR ESTRUCTURA PARA EL RESULTADO
    user := &domain.User{}

    // 3️⃣ EJECUTAR QUERY
    err := r.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
        &user.ID,
        &user.Email,
        &user.PasswordHash,
        &user.FirstName,
        &user.LastName,
        &user.Role,
        &user.Avatar,
        &user.IsActive,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    // 4️⃣ MANEJAR ERRORES
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domain.ErrUserNotFound  // Error del dominio
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }

    // 5️⃣ RETORNAR USUARIO
    return user, nil
    // user = {
    //   ID: "550e8400-e29b-41d4-a716-446655440000",
    //   Email: "miguel@example.com",
    //   PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
    //   FirstName: "Miguel",
    //   LastName: "Mamani",
    //   Role: "student",
    //   IsActive: true,
    //   ...
    // }
}
```

**Responsabilidades del Repository:**
- ✅ Ejecutar queries SQL
- ✅ Mapear resultados SQL a entidades del dominio
- ✅ Convertir errores SQL a errores del dominio
- ❌ NO tiene lógica de negocio
- ❌ NO tiene validaciones de negocio

---

### Paso 5: PostgreSQL ejecuta la query

```sql
-- En PostgreSQL
SELECT id, email, password_hash, first_name, last_name,
       role, avatar, is_active, created_at, updated_at
FROM users
WHERE email = 'miguel@example.com';
```

**Resultado:**
```
┌──────────────────────────────────────┬─────────────────────┬──────────────────────────────────────┬────────────┬─────────┬─────────┬────────┬───────────┬──────────────────────┬──────────────────────┐
│ id                                   │ email               │ password_hash                        │ first_name │ last... │ role    │ avatar │ is_active │ created_at           │ updated_at           │
├──────────────────────────────────────┼─────────────────────┼──────────────────────────────────────┼────────────┼─────────┼─────────┼────────┼───────────┼──────────────────────┼──────────────────────┤
│ 550e8400-e29b-41d4-a716-446655440000 │ miguel@example.com  │ $2a$10$N9qo8uLOickgx2ZMRZoMyeI... │ Miguel     │ Mamani  │ student │ NULL   │ true      │ 2024-10-01 10:00:00  │ 2024-10-18 15:30:00  │
└──────────────────────────────────────┴─────────────────────┴──────────────────────────────────────┴────────────┴─────────┴─────────┴────────┴───────────┴──────────────────────┴──────────────────────┘
```

---

### Paso 6: Los datos regresan al Service

El repository retorna la entidad `domain.User` al servicio.

El servicio:
1. ✅ Verifica que el usuario esté activo
2. ✅ Verifica la contraseña
3. ✅ Genera los tokens
4. ✅ Retorna el resultado

---

### Paso 7: El resultado regresa al Handler

El handler:
1. ✅ Recibe `accessToken`, `refreshToken`, `user`
2. ✅ Configura la cookie con el refresh token
3. ✅ Genera respuesta JSON

---

### Paso 8: Respuesta HTTP al cliente

```
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: refresh_token=a7f3c9e2...; HttpOnly; Secure; SameSite=Strict; Path=/api/auth; Max-Age=604800

{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "miguel@example.com",
    "firstName": "Miguel",
    "lastName": "Mamani",
    "role": "student",
    "avatar": null,
    "created_at": "2024-10-01T10:00:00Z"
  }
}
```

---

## 📊 Diagrama de Secuencia

```
Cliente            Handler          Service         Repository        PostgreSQL
  │                  │                │                 │                  │
  │──POST /login────▶│                │                 │                  │
  │  {email,pass}    │                │                 │                  │
  │                  │                │                 │                  │
  │                  │──Login()──────▶│                 │                  │
  │                  │                │                 │                  │
  │                  │                │──GetByEmail()──▶│                  │
  │                  │                │                 │                  │
  │                  │                │                 │──SELECT * FROM──▶│
  │                  │                │                 │   users WHERE... │
  │                  │                │                 │                  │
  │                  │                │                 │◀────User Data────│
  │                  │                │                 │                  │
  │                  │                │◀────User────────│                  │
  │                  │                │                 │                  │
  │                  │                │ VerifyPassword  │                  │
  │                  │                │ GenerateTokens  │                  │
  │                  │                │                 │                  │
  │                  │                │──UpdateLast─────▶                  │
  │                  │                │  Login()        │                  │
  │                  │                │                 │                  │
  │                  │◀───Tokens──────│                 │                  │
  │                  │    + User      │                 │                  │
  │                  │                │                 │                  │
  │◀────200 OK───────│                │                 │                  │
  │  {token, user}   │                │                 │                  │
  │  + Cookie        │                │                 │                  │
```

---

## 🔍 ¿Qué pasa si...?

### Escenario 1: Email no existe

```
1. Handler recibe petición
   ↓
2. Service llama a GetByEmail()
   ↓
3. Repository ejecuta SQL → 0 filas
   ↓
4. Repository retorna domain.ErrUserNotFound
   ↓
5. Service retorna domain.ErrInvalidCredentials
   (No revela si el email existe)
   ↓
6. Handler convierte a HTTP 401 Unauthorized
   ↓
7. Cliente recibe:
   {
     "error": "Invalid credentials"
   }
```

### Escenario 2: Contraseña incorrecta

```
1. Handler recibe petición
   ↓
2. Service obtiene usuario correctamente
   ↓
3. Service verifica password → ❌ No coincide
   ↓
4. Service retorna domain.ErrInvalidCredentials
   ↓
5. Handler convierte a HTTP 401 Unauthorized
   ↓
6. Cliente recibe:
   {
     "error": "Invalid credentials"
   }
```

### Escenario 3: Usuario inactivo

```
1. Handler recibe petición
   ↓
2. Service obtiene usuario correctamente
   ↓
3. Service verifica user.IsActive → false
   ↓
4. Service retorna domain.ErrUserInactive
   ↓
5. Handler convierte a HTTP 403 Forbidden
   ↓
6. Cliente recibe:
   {
     "error": "User account is inactive"
   }
```

### Escenario 4: Error de base de datos

```
1. Handler recibe petición
   ↓
2. Service llama a GetByEmail()
   ↓
3. PostgreSQL está caído → Error de conexión
   ↓
4. Repository retorna error
   ↓
5. Service propaga el error
   ↓
6. Handler convierte a HTTP 500 Internal Server Error
   ↓
7. Cliente recibe:
   {
     "error": "Internal server error"
   }

(El error real se loguea pero NO se muestra al usuario)
```

---

## 💡 Puntos Clave

### 1. Separación de Responsabilidades

| Capa | Responsabilidad | NO debe hacer |
|------|-----------------|---------------|
| **Handler** | Manejar HTTP | Lógica de negocio, SQL |
| **Service** | Lógica de negocio | HTTP, SQL |
| **Repository** | Acceso a datos | Lógica de negocio, HTTP |

### 2. Uso de Interfaces (Ports)

```go
// El servicio NO sabe que es PostgreSQL
type authService struct {
    userRepo output.UserRepository  // ← Interfaz, no implementación
}

// Podría ser PostgreSQL, MongoDB, MySQL, etc.
```

**Ventaja:** Si cambias de PostgreSQL a MongoDB, solo cambias el adapter, no el service.

### 3. Errores del Dominio

```go
// Definidos en internal/core/domain/errors.go
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrUserInactive      = errors.New("user account is inactive")
)
```

**Ventaja:** Consistencia en los errores en toda la aplicación.

### 4. Validaciones en el Service

```go
// ✅ BIEN: Service valida
func (s *authService) Login(email, password string) {
    if err := validator.ValidateEmail(email); err != nil {
        return domain.ErrInvalidCredentials
    }
}

// ❌ MAL: Handler valida
func (h *AuthHandler) Login(w, r) {
    if !isValidEmail(req.Email) {  // ← NO
        return
    }
}
```

**Ventaja:** Si agregas GraphQL, CLI, etc., no duplicas validaciones.

### 5. Testeable

Puedes testear el service sin base de datos:

```go
func TestLogin(t *testing.T) {
    // Mock repository
    mockRepo := &mockUserRepository{
        users: map[string]*domain.User{
            "test@example.com": {...},
        },
    }

    service := NewAuthService(mockRepo, mockCache, "secret")
    token, _, user, err := service.Login(ctx, "test@example.com", "password")

    // Asserts...
}
```

---

## 📚 Archivos Involucrados

```
cmd/server/main.go
  ├─ Inyección de dependencias
  └─ Configuración de rutas

internal/adapters/input/http/auth_handler.go
  ├─ Recibe petición HTTP
  ├─ Llama al servicio
  └─ Retorna respuesta HTTP

internal/core/services/auth_service.go
  ├─ Lógica de negocio
  ├─ Validaciones
  ├─ Generación de tokens
  └─ Orquestación

internal/core/ports/input/auth_service.go
  └─ Interfaz del servicio

internal/core/ports/output/user_repository.go
  └─ Interfaz del repositorio

internal/adapters/output/persistence/postgres/user_repository.go
  ├─ Queries SQL
  └─ Mapeo de datos

internal/core/domain/user.go
  └─ Entidad User

internal/core/domain/errors.go
  └─ Errores del dominio

pkg/validator/validator.go
  └─ Validaciones reutilizables

pkg/httputil/response.go
  └─ Helpers HTTP

pkg/jwt/jwt.go
  └─ Generación y validación de tokens
```

---

**Este es el flujo completo de una petición en Arquitectura Hexagonal. Cada capa tiene una responsabilidad clara y está desacoplada de las demás.**
