# 📚 Documentación del Proyecto - Streaming Platform

## Índice de Documentos

Esta carpeta contiene toda la documentación técnica del proyecto de plataforma de streaming construido con **Arquitectura Hexagonal**.

---

## 📖 Documentos Disponibles

### 1. 🏗️ [ARQUITECTURA_HEXAGONAL.md](./ARQUITECTURA_HEXAGONAL.md)

**Descripción**: Guía completa de la arquitectura hexagonal del proyecto.

**Contenido**:
- ¿Qué es la Arquitectura Hexagonal?
- ¿Por qué usarla?
- Conceptos fundamentales (Domain, Ports, Adapters)
- Estructura completa del proyecto
- Flujo de datos en la aplicación
- Ejemplo práctico completo (Login)
- Estrategias de testing
- Mejores prácticas y antipatrones
- Comparación con otras arquitecturas

**Cuándo leerlo**: Cuando quieras entender a profundidad la arquitectura del proyecto o estés comenzando con Hexagonal Architecture.

---

### 2. ⚡ [GUIA_RAPIDA.md](./GUIA_RAPIDA.md)

**Descripción**: Referencia rápida y cheat sheet para desarrollo diario.

**Contenido**:
- Tabla "¿Dónde pongo mi código?"
- Reglas de oro (PUEDES vs NO PUEDES)
- Flujo de trabajo para agregar features
- Plantillas de código listas para usar:
  - Nueva entidad del dominio
  - Nuevo servicio
  - Nuevo repositorio
  - Nuevo handler HTTP
- Template de tests
- Checklist para nueva feature
- Comandos útiles
- FAQ

**Cuándo leerlo**: Cuando estés desarrollando y necesites una referencia rápida de cómo estructurar tu código.

---

### 3. 📝 [EJEMPLO_COMPLETO.md](./EJEMPLO_COMPLETO.md)

**Descripción**: Walkthrough paso a paso de un flujo completo (Login).

**Contenido**:
- Explicación detallada de cada paso del flujo de Login
- Análisis de cada archivo involucrado
- Diagrama de secuencia
- Código real del proyecto
- Escenarios de error ("¿Qué pasa si...?")
- Lista de todos los archivos involucrados

**Cuándo leerlo**: Cuando quieras entender cómo fluyen los datos a través de todas las capas de la arquitectura con un ejemplo concreto.

---

### 4. 📊 [DIAGRAMA_ARQUITECTURA.md](./DIAGRAMA_ARQUITECTURA.md)

**Descripción**: Diagramas visuales de la arquitectura del proyecto.

**Contenido**:
- Vista general de la arquitectura hexagonal
- Diagrama de capas y componentes
- Flujo de una petición HTTP
- Gráfico de dependencias
- Inyección de dependencias en main.go
- Separación de responsabilidades
- Matriz de dependencias
- Estrategia de testing
- Beneficios de la arquitectura

**Cuándo leerlo**: Cuando prefieras visualizar la arquitectura con diagramas en lugar de leer texto.

---

### 5. 🧪 [TESTING.md](./TESTING.md)

**Descripción**: Guía completa de testing y pruebas.

**Contenido**:
- Introducción al testing con mocks
- Tipos de tests (Unitarios, Integración, E2E)
- Cómo ejecutar tests
- Anatomía de un test unitario
- Qué son los mocks y por qué usarlos
- Estructura de archivos de tests
- Cobertura de código
- Mejores prácticas
- Tests actuales del proyecto
- Próximos pasos

**Cuándo leerlo**: Cuando necesites escribir tests o entender cómo probar el código de manera efectiva.

---

## 🎯 Flujo de Aprendizaje Recomendado

### Para Principiantes en Hexagonal Architecture

```
1. ARQUITECTURA_HEXAGONAL.md (Sección 1-4)
   └─> Entender los conceptos básicos

2. DIAGRAMA_ARQUITECTURA.md
   └─> Visualizar la estructura

3. EJEMPLO_COMPLETO.md
   └─> Ver un caso real paso a paso

4. GUIA_RAPIDA.md
   └─> Referencia rápida para desarrollo

5. TESTING.md
   └─> Aprender a escribir tests
```

### Para Desarrolladores con Experiencia

```
1. DIAGRAMA_ARQUITECTURA.md
   └─> Vista general rápida

2. GUIA_RAPIDA.md
   └─> Plantillas y referencias

3. TESTING.md
   └─> Estrategia de testing

4. ARQUITECTURA_HEXAGONAL.md (como referencia)
   └─> Profundizar cuando sea necesario
```

### Para Revisar Código Existente

```
1. EJEMPLO_COMPLETO.md
   └─> Entender el flujo de un caso real

2. DIAGRAMA_ARQUITECTURA.md
   └─> Ubicar componentes en la arquitectura

3. GUIA_RAPIDA.md (FAQ y reglas de oro)
   └─> Verificar si se siguen las convenciones
```

---

## 📂 Estructura del Proyecto

```
streaming-platform/
├── cmd/
│   └── server/
│       └── main.go                    ← Punto de entrada, DI
│
├── internal/
│   ├── core/                          ← NÚCLEO (lógica de negocio)
│   │   ├── domain/                    ← Entidades y errores
│   │   ├── ports/
│   │   │   ├── input/                 ← Interfaces de servicios
│   │   │   └── output/                ← Interfaces de repositorios
│   │   └── services/                  ← Implementación de lógica
│   │
│   └── adapters/                      ← DETALLES DE IMPLEMENTACIÓN
│       ├── input/
│       │   └── http/                  ← HTTP Handlers
│       └── output/
│           └── persistence/
│               ├── postgres/          ← PostgreSQL repositories
│               └── redis/             ← Redis cache
│
├── pkg/                               ← Utilidades compartidas
│   ├── config/
│   ├── database/
│   ├── jwt/
│   ├── logger/
│   └── validator/
│
└── docs/                              ← DOCUMENTACIÓN
    ├── ARQUITECTURA_HEXAGONAL.md
    ├── GUIA_RAPIDA.md
    ├── EJEMPLO_COMPLETO.md
    ├── DIAGRAMA_ARQUITECTURA.md
    ├── TESTING.md
    └── README_DOCS.md (este archivo)
```

---

## 🚀 Quick Start

### 1. Leer la Documentación

```bash
# Abrir la guía rápida
cat docs/GUIA_RAPIDA.md

# Abrir un ejemplo completo
cat docs/EJEMPLO_COMPLETO.md
```

### 2. Compilar el Proyecto

```bash
# Compilar todo
go build ./...

# Compilar solo el servidor
go build ./cmd/server
```

### 3. Ejecutar Tests

```bash
# Ejecutar todos los tests
go test ./...

# Ejecutar tests con verbose
go test ./... -v

# Ejecutar tests con cobertura
go test ./... -cover
```

### 4. Ejecutar el Servidor

```bash
# Ejecutar el servidor
./server

# O directamente con go run
go run ./cmd/server
```

---

## 🎓 Conceptos Clave

### Arquitectura Hexagonal en 3 Puntos

1. **Core (Hexágono Central)**: Lógica de negocio pura, sin dependencias externas
2. **Ports (Interfaces)**: Contratos que definen QUÉ puede hacer la aplicación
3. **Adapters (Implementaciones)**: CÓMO se comunica con el mundo exterior

### Flujo de Dependencias

```
INPUT ADAPTERS (HTTP)
      ↓
INPUT PORTS (Interfaces)
      ↓
SERVICES (Lógica de negocio) ← NÚCLEO
      ↓
OUTPUT PORTS (Interfaces)
      ↓
OUTPUT ADAPTERS (PostgreSQL/Redis)
```

### Regla de Oro

**Las dependencias apuntan HACIA ADENTRO**

- ✅ Adapters pueden importar Core
- ✅ Core usa Ports (interfaces)
- ❌ Core NO puede importar Adapters

---

## 📋 Checklist de Desarrollo

Al agregar una nueva feature:

- [ ] 1. Definir entidad en `internal/core/domain/`
- [ ] 2. Definir errores en `internal/core/domain/errors.go`
- [ ] 3. Definir Input Port en `internal/core/ports/input/`
- [ ] 4. Definir Output Port en `internal/core/ports/output/`
- [ ] 5. Implementar Service en `internal/core/services/`
- [ ] 6. Implementar Repository en `internal/adapters/output/persistence/`
- [ ] 7. Implementar Handler en `internal/adapters/input/http/`
- [ ] 8. Conectar en `main.go` con inyección de dependencias
- [ ] 9. Escribir tests unitarios
- [ ] 10. Escribir tests de integración
- [ ] 11. Probar endpoints

---

## 🔧 Comandos Útiles

### Desarrollo

```bash
# Compilar
go build ./...

# Ejecutar tests
go test ./...

# Ver cobertura
go test ./... -cover

# Ver estructura del proyecto
tree -L 3 internal/

# Buscar TODOs
grep -r "TODO" internal/
```

### Testing

```bash
# Tests de un paquete específico
go test ./internal/core/services/

# Test específico
go test ./internal/core/services/ -run TestAuthService_Login

# Tests con verbose
go test ./... -v

# Generar reporte de cobertura HTML
go test ./internal/core/services/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 📞 Recursos Adicionales

### Documentación Externa

- [Go Testing](https://pkg.go.dev/testing)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### Herramientas

- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP Router
- [Go-Redis](https://github.com/redis/go-redis) - Redis Client
- [pgx](https://github.com/jackc/pgx) - PostgreSQL Driver
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Password Hashing
- [UUID](https://github.com/google/uuid) - UUID Generation

---

## 🤝 Contribuir

Al contribuir código al proyecto:

1. **Lee la documentación** (especialmente GUIA_RAPIDA.md)
2. **Sigue la arquitectura hexagonal**
3. **Escribe tests** para tu código
4. **Sigue las convenciones** de nombres y estructura
5. **Documenta** funcionalidad compleja

---

## 📝 Notas Finales

Esta documentación está viva y debe actualizarse cuando:

- Se agreguen nuevas features importantes
- Se cambien patrones arquitectónicos
- Se descubran mejores prácticas
- Se encuentren errores en la documentación

**Última actualización**: 2024

**Versión del proyecto**: 1.0.0

**Arquitectura**: Hexagonal (Ports & Adapters)

**Lenguaje**: Go 1.21+

---

## 🎯 TL;DR (Resumen Ejecutivo)

**¿Nuevo en el proyecto?** Lee en este orden:

1. `DIAGRAMA_ARQUITECTURA.md` - Visualizar la estructura
2. `EJEMPLO_COMPLETO.md` - Ver un caso real
3. `GUIA_RAPIDA.md` - Plantillas para empezar a desarrollar

**¿Necesitas desarrollar?**

1. Usa las plantillas de `GUIA_RAPIDA.md`
2. Sigue el checklist de desarrollo
3. Escribe tests (ver `TESTING.md`)

**¿Dudas sobre la arquitectura?**

1. Consulta `ARQUITECTURA_HEXAGONAL.md`
2. Revisa el FAQ en `GUIA_RAPIDA.md`
3. Analiza el código existente con `EJEMPLO_COMPLETO.md` como guía

---

Happy Coding! 🚀✨
