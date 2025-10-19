# 🧪 Testing - Guía de Pruebas

## 📋 Índice

1. [Introducción](#introducción)
2. [Tipos de Tests](#tipos-de-tests)
3. [Ejecutar Tests](#ejecutar-tests)
4. [Tests Unitarios](#tests-unitarios)
5. [Estructura de Tests](#estructura-de-tests)
6. [Mocks](#mocks)
7. [Cobertura de Código](#cobertura-de-código)
8. [Mejores Prácticas](#mejores-prácticas)

---

## 📚 Introducción

Este proyecto utiliza **testing basado en mocks** para probar la lógica de negocio de manera aislada, sin depender de bases de datos reales o servicios externos.

### ✅ Ventajas del Testing con Mocks

- **Rapidez**: Los tests se ejecutan en milisegundos (no requieren DB)
- **Aislamiento**: Cada test prueba UNA sola cosa
- **Confiabilidad**: No dependen de servicios externos
- **Facilidad**: Puedes simular cualquier escenario (incluso errores)

---

## 📊 Tipos de Tests

### 1. Tests Unitarios (Unit Tests)

**Ubicación**: `internal/core/services/*_test.go`

Prueban la **lógica de negocio** de manera aislada usando mocks.

```go
// Ejemplo: Probar que el login funciona correctamente
func TestAuthService_Login_Success(t *testing.T) {
    // Arrange: Preparar mocks
    mockUserRepo := newMockUserRepository()
    mockCacheRepo := newMockCacheRepository()

    // Act: Ejecutar la acción
    accessToken, refreshToken, profile, err := authService.Login(...)

    // Assert: Verificar el resultado
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

### 2. Tests de Integración (Integration Tests)

**Ubicación**: `internal/adapters/output/persistence/postgres/*_test.go` (futuro)

Prueban la **interacción con la base de datos real**.

```go
// Ejemplo: Probar que el repositorio guarda correctamente
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB() // Base de datos de prueba
    repo := NewUserRepository(db)

    user := &domain.User{...}
    err := repo.Create(context.Background(), user)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

### 3. Tests E2E (End-to-End)

**Ubicación**: `tests/e2e/` (futuro)

Prueban el **flujo completo** desde HTTP hasta la base de datos.

```go
// Ejemplo: Probar el endpoint de login
func TestLoginEndpoint(t *testing.T) {
    server := setupTestServer()

    resp := httptest.Post("/api/auth/login", ...)

    if resp.StatusCode != 200 {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }
}
```

---

## 🚀 Ejecutar Tests

### Ejecutar TODOS los tests

```bash
go test ./...
```

### Ejecutar tests de un paquete específico

```bash
# Tests del core
go test ./internal/core/services/

# Tests de un servicio específico
go test ./internal/core/services/ -run TestAuthService

# Tests de un método específico
go test ./internal/core/services/ -run TestAuthService_Login
```

### Ejecutar tests con salida detallada

```bash
go test ./internal/core/services/ -v
```

**Salida esperada:**
```
=== RUN   TestAuthService_Register_Success
--- PASS: TestAuthService_Register_Success (0.10s)
=== RUN   TestAuthService_Login_Success
--- PASS: TestAuthService_Login_Success (0.19s)
PASS
ok      streaming-platform/internal/core/services       1.471s
```

---

## 🧪 Tests Unitarios

### Anatomía de un Test Unitario

Cada test sigue la estructura **AAA (Arrange, Act, Assert)**:

```go
func TestAuthService_Register_Success(t *testing.T) {
    // ========== ARRANGE (Preparar) ==========
    // 1. Crear mocks
    mockUserRepo := newMockUserRepository()
    mockCacheRepo := newMockCacheRepository()

    // 2. Crear el servicio con los mocks
    authService := NewAuthService(mockUserRepo, mockCacheRepo, "test-secret")

    // 3. Preparar datos de entrada
    email := "test@example.com"
    password := "SecurePass123!"
    firstName := "Test"
    lastName := "User"
    role := "user"

    // ========== ACT (Actuar) ==========
    // Ejecutar el método que queremos probar
    accessToken, refreshToken, profile, err := authService.Register(
        context.Background(),
        email,
        password,
        firstName,
        lastName,
        role,
    )

    // ========== ASSERT (Verificar) ==========
    // Verificar que el resultado es el esperado
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }

    if accessToken == "" {
        t.Error("Expected access token to be generated")
    }

    if profile.Email != email {
        t.Errorf("Expected email %s, got %s", email, profile.Email)
    }
}
```

### Tests de Casos de Error

Es IMPORTANTE probar que los errores se manejen correctamente:

```go
func TestAuthService_Login_InvalidCredentials(t *testing.T) {
    // Arrange
    mockUserRepo := newMockUserRepository()
    mockCacheRepo := newMockCacheRepository()
    authService := NewAuthService(mockUserRepo, mockCacheRepo, "secret")

    // Registrar un usuario con password "Correct123!"
    authService.Register(ctx, "test@example.com", "Correct123!", ...)

    // Act - Intentar login con password INCORRECTA
    _, _, _, err := authService.Login(ctx, "test@example.com", "Wrong123!")

    // Assert - Verificar que retorna el error correcto
    if err != domain.ErrInvalidCredentials {
        t.Errorf("Expected ErrInvalidCredentials, got %v", err)
    }
}
```

---

## 🎭 Mocks

### ¿Qué es un Mock?

Un **mock** es una implementación FALSA de una interfaz que simula el comportamiento real pero sin tocar la base de datos o servicios externos.

### Ejemplo de Mock del UserRepository

```go
// Mock del UserRepository
type mockUserRepository struct {
    users map[string]*domain.User  // Almacén en memoria
}

func newMockUserRepository() *mockUserRepository {
    return &mockUserRepository{
        users: make(map[string]*domain.User),
    }
}

// Implementar los métodos de la interfaz
func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
    // Simular que guardamos en DB, pero realmente guardamos en memoria
    m.users[user.Email] = user
    return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    user, exists := m.users[email]
    if !exists {
        return nil, domain.ErrUserNotFound
    }
    return user, nil
}
```

### ¿Por qué usar Mocks?

```
SIN MOCKS:
Test → Service → PostgreSQL → Disco Duro
    ❌ Lento (100-500ms por test)
    ❌ Requiere base de datos configurada
    ❌ Puede fallar por problemas de red/DB
    ❌ Difícil simular errores

CON MOCKS:
Test → Service → Mock (memoria)
    ✅ Rápido (1-10ms por test)
    ✅ No requiere base de datos
    ✅ Nunca falla por problemas externos
    ✅ Puedes simular CUALQUIER escenario
```

---

## 📁 Estructura de Tests

```
streaming-platform/
├── internal/
│   ├── core/
│   │   └── services/
│   │       ├── auth_service.go         ← Código de producción
│   │       └── auth_service_test.go    ← Tests unitarios ✅
│   │
│   └── adapters/
│       └── output/
│           └── persistence/
│               └── postgres/
│                   ├── user_repository.go
│                   └── user_repository_test.go  ← Tests de integración (futuro)
│
└── tests/
    └── e2e/
        └── login_test.go  ← Tests E2E (futuro)
```

### Convenciones de Nombres

| Tipo de Test | Nombre del Archivo | Ubicación |
|-------------|-------------------|-----------|
| Test Unitario | `auth_service_test.go` | Al lado del archivo que prueba |
| Test de Integración | `user_repository_integration_test.go` | Al lado del repository |
| Test E2E | `login_e2e_test.go` | En `tests/e2e/` |

---

## 📊 Cobertura de Código

### Ver cobertura de tests

```bash
# Ejecutar tests con cobertura
go test ./internal/core/services/ -cover

# Salida:
# PASS
# coverage: 78.5% of statements
# ok      streaming-platform/internal/core/services       1.471s
```

### Generar reporte HTML de cobertura

```bash
# 1. Generar archivo de cobertura
go test ./internal/core/services/ -coverprofile=coverage.out

# 2. Ver reporte en HTML
go tool cover -html=coverage.out
```

Esto abrirá un navegador mostrando qué líneas de código están cubiertas (verde) y cuáles no (rojo).

### Cobertura por paquete

```bash
go test ./... -cover

# Salida:
# ok      streaming-platform/internal/core/services       1.471s  coverage: 78.5%
# ok      streaming-platform/internal/adapters/input/http 0.234s  coverage: 0.0%
```

---

## ✅ Mejores Prácticas

### 1. Nombres Descriptivos

❌ **Malo:**
```go
func TestLogin(t *testing.T) { ... }
```

✅ **Bueno:**
```go
func TestAuthService_Login_Success(t *testing.T) { ... }
func TestAuthService_Login_InvalidCredentials(t *testing.T) { ... }
func TestAuthService_Login_UserNotFound(t *testing.T) { ... }
```

**Patrón:** `Test<Servicio>_<Método>_<Escenario>`

### 2. Un Test = Una Cosa

❌ **Malo:**
```go
func TestAuthService(t *testing.T) {
    // Probar login
    // Probar register
    // Probar logout
    // ... demasiado en un solo test
}
```

✅ **Bueno:**
```go
func TestAuthService_Login_Success(t *testing.T) { ... }
func TestAuthService_Register_Success(t *testing.T) { ... }
func TestAuthService_Logout_Success(t *testing.T) { ... }
```

### 3. Probar Casos Felices Y Casos de Error

Para cada método, prueba:

- ✅ **Caso feliz** (cuando todo funciona)
- ✅ **Casos de error** (validaciones, errores de negocio)
- ✅ **Casos límite** (valores nulos, vacíos, extremos)

**Ejemplo:**

```go
// Caso feliz
func TestAuthService_Register_Success(t *testing.T) { ... }

// Casos de error
func TestAuthService_Register_InvalidEmail(t *testing.T) { ... }
func TestAuthService_Register_DuplicateEmail(t *testing.T) { ... }
func TestAuthService_Register_WeakPassword(t *testing.T) { ... }

// Casos límite
func TestAuthService_Register_EmptyFields(t *testing.T) { ... }
```

### 4. No Usar Datos Reales en Tests

❌ **Malo:**
```go
email := "miguel@example.com"  // Email real
password := "myRealPassword123!"
```

✅ **Bueno:**
```go
email := "test@example.com"
password := "TestPass123!"
```

### 5. Limpiar Después de Cada Test

Si usas bases de datos de prueba:

```go
func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)  // ← Limpiar al final

    repo := NewUserRepository(db)
    // ... test ...
}
```

### 6. Usar Table-Driven Tests para Múltiples Casos

```go
func TestAuthService_Register_Validation(t *testing.T) {
    tests := []struct {
        name      string
        email     string
        password  string
        wantError error
    }{
        {
            name:      "Invalid email",
            email:     "not-an-email",
            password:  "Valid123!",
            wantError: domain.ErrInvalidEmail,
        },
        {
            name:      "Weak password",
            email:     "test@example.com",
            password:  "weak",
            wantError: domain.ErrWeakPassword,
        },
        // ... más casos
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, _, _, err := authService.Register(ctx, tt.email, tt.password, ...)
            if err != tt.wantError {
                t.Errorf("Expected error %v, got %v", tt.wantError, err)
            }
        })
    }
}
```

---

## 🎯 Qué Tests Tenemos Actualmente

### ✅ AuthService Tests (8 tests)

**Ubicación**: `internal/core/services/auth_service_test.go`

| Test | Descripción | Estado |
|------|-------------|--------|
| `TestAuthService_Register_Success` | Registrar usuario correctamente | ✅ PASS |
| `TestAuthService_Register_InvalidEmail` | Rechazar email inválido | ✅ PASS |
| `TestAuthService_Register_DuplicateEmail` | Rechazar email duplicado | ✅ PASS |
| `TestAuthService_Login_Success` | Login exitoso | ✅ PASS |
| `TestAuthService_Login_InvalidCredentials` | Rechazar password incorrecta | ✅ PASS |
| `TestAuthService_Login_UserNotFound` | Rechazar usuario inexistente | ✅ PASS |
| `TestAuthService_ValidateToken_Success` | Validar token JWT correctamente | ✅ PASS |
| `TestAuthService_ValidateToken_InvalidToken` | Rechazar token inválido | ✅ PASS |

**Ejecutar estos tests:**
```bash
go test ./internal/core/services/ -v
```

---

## 🚦 Próximos Pasos

### 1. Agregar más tests unitarios

```bash
# Tests para UserService
internal/core/services/user_service_test.go

# Tests para VideoService
internal/core/services/video_service_test.go

# Tests para StreamingService
internal/core/services/streaming_service_test.go
```

### 2. Agregar tests de integración

```bash
# Tests para PostgreSQL repositories
internal/adapters/output/persistence/postgres/user_repository_test.go
internal/adapters/output/persistence/postgres/video_repository_test.go
```

### 3. Agregar tests E2E

```bash
# Tests end-to-end
tests/e2e/auth_test.go
tests/e2e/video_test.go
tests/e2e/streaming_test.go
```

---

## 📚 Recursos Adicionales

### Documentación Oficial de Go Testing

- [Testing Package](https://pkg.go.dev/testing)
- [Subtests and Table-Driven Tests](https://go.dev/blog/subtests)

### Guías Relacionadas

- [ARQUITECTURA_HEXAGONAL.md](./ARQUITECTURA_HEXAGONAL.md) - Arquitectura del proyecto
- [GUIA_RAPIDA.md](./GUIA_RAPIDA.md) - Referencia rápida
- [EJEMPLO_COMPLETO.md](./EJEMPLO_COMPLETO.md) - Ejemplo paso a paso

---

## 💡 Tips Rápidos

```bash
# Ejecutar solo tests que fallan
go test ./... -failfast

# Ejecutar tests en paralelo
go test ./... -parallel 4

# Ejecutar tests con timeout
go test ./... -timeout 30s

# Ver qué tests se están ejecutando
go test ./... -v

# Ejecutar tests y mostrar cobertura
go test ./... -cover

# Ejecutar un test específico por nombre
go test ./internal/core/services/ -run TestAuthService_Login_Success
```

---

**¡Recuerda!** Los tests son tan importantes como el código de producción. Un buen conjunto de tests te da **confianza** para hacer cambios sin miedo a romper cosas.

Happy Testing! 🧪✨
