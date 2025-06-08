# StreamLearn - Plataforma de Streaming de Video Tutoriales

Una plataforma completa de streaming de video tutoriales construida con Go (backend) y Vanilla JavaScript (frontend), implementando arquitectura limpia y tecnologías modernas.

## 🚀 Características

### Backend
- **Arquitectura limpia** con separación clara de responsabilidades
- **API RESTful** con autenticación JWT
- **Procesamiento asíncrono** de video con worker pools
- **Streaming HLS** para reproducción adaptiva
- **Transcodificación automática** con FFmpeg a múltiples calidades
- **Cache multicapa** con Redis
- **Base de datos PostgreSQL** con migraciones
- **Message queue** con RabbitMQ
- **Monitoreo** con Prometheus y Grafana

### Frontend
- **Progressive Web App (PWA)** con funcionalidad offline
- **Vanilla JavaScript** modular con ES6
- **Reproductor HLS** con Video.js
- **Interfaz responsive** con CSS moderno
- **Upload con progreso** y drag & drop
- **Búsqueda y filtros** avanzados
- **Gestión de perfil** y favoritos

## 🛠️ Tecnologías

### Backend
- **Go 1.21** - Lenguaje principal
- **Gorilla Mux** - Router HTTP
- **PostgreSQL** - Base de datos principal
- **Redis** - Cache y sesiones
- **FFmpeg** - Procesamiento de video
- **RabbitMQ** - Message queue
- **Docker** - Containerización

### Frontend
- **Vanilla JavaScript (ES6+)** - Sin frameworks
- **HTML5 & CSS3** - Estructura y estilos
- **Video.js** - Reproductor de video
- **Service Workers** - Funcionalidad PWA
- **IndexedDB** - Storage local

### Infraestructura
- **Docker Compose** - Orquestación local
- **Nginx** - Reverse proxy
- **MinIO** - Object storage (S3 compatible)
- **Elasticsearch** - Búsqueda avanzada
- **Prometheus & Grafana** - Monitoreo

## 📋 Requisitos Previos

- Docker y Docker Compose
- Git
- (Opcional) Go 1.21+ para desarrollo

## 🚀 Instalación y Configuración

### 1. Clonar el repositorio

```bash
git clone <repository-url>
cd streaming-platform
```

### 2. Configurar variables de entorno

Crear archivo `.env`:

```bash
# Server
PORT=8080
LOG_LEVEL=info

# Database
DATABASE_URL=postgres://admin:password123@localhost:5432/streaming_platform?sslmode=disable
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Storage
STORAGE_PATH=./storage
CDN_BASE_URL=http://localhost:8080/static

# FFmpeg
FFMPEG_PATH=ffmpeg

# Worker Pool
WORKER_POOL_SIZE=4

# Email (opcional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### 3. Inicializar con Docker Compose

```bash
# Construir e iniciar todos los servicios
docker-compose up -d

# Ver logs
docker-compose logs -f app

# Verificar estado de servicios
docker-compose ps
```

### 4. Configurar base de datos

```bash
# Ejecutar migraciones (si están configuradas)
docker-compose exec app ./main migrate

# O crear tablas manualmente
docker-compose exec postgres psql -U admin -d streaming_platform -f /docker-entrypoint-initdb.d/init.sql
```

## 🖥️ Uso

### Acceder a la aplicación
- **Aplicación web**: http://localhost:8080
- **API**: http://localhost:8080/api
- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **RabbitMQ Management**: http://localhost:15672 (admin/password123)
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin123)

### Funcionalidades principales

1. **Registro/Login**: Crear cuenta o iniciar sesión
2. **Explorar videos**: Buscar y filtrar contenido
3. **Reproducir videos**: Streaming HLS adaptivo
4. **Subir videos**: Upload con procesamiento automático
5. **Gestionar perfil**: Videos propios, favoritos, historial

## 📁 Estructura del Proyecto

```
streaming-platform/
├── cmd/server/           # Punto de entrada
├── internal/
│   ├── domain/          # Entidades y reglas de negocio
│   ├── usecase/         # Casos de uso
│   ├── delivery/        # Handlers HTTP y Workers
│   └── repository/      # Acceso a datos
├── pkg/                 # Utilidades compartidas
├── web/                 # Frontend
│   ├── static/
│   │   ├── css/        # Estilos
│   │   ├── js/         # JavaScript
│   │   └── assets/     # Recursos
│   └── index.html      # Página principal
├── storage/            # Almacenamiento de archivos
├── docker-compose.yml  # Orquestación
├── Dockerfile          # Imagen de la app
└── README.md
```

## 🔧 Desarrollo

### Desarrollo local (sin Docker)

1. **Instalar dependencias**:
```bash
go mod download
```

2. **Ejecutar servicios de dependencias**:
```bash
docker-compose up -d postgres redis rabbitmq minio elasticsearch
```

3. **Ejecutar aplicación**:
```bash
go run cmd/server/main.go
```

### Testing

```bash
# Tests unitarios
go test ./...

# Tests con coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Build para producción

```bash
# Build binario
go build -o bin/server cmd/server/main.go

# Build imagen Docker
docker build -t streaming-platform:latest .
```

## 📊 Monitoreo y Logs

### Logs de aplicación
```bash
# Ver logs en tiempo real
docker-compose logs -f app

# Logs específicos
docker-compose logs app | grep ERROR
```

### Métricas en Grafana
1. Acceder a http://localhost:3000
2. Login: admin/admin123
3. Importar dashboards pre-configurados
4. Monitorear performance y uso

### Health Checks
```bash
# Verificar salud de la aplicación
curl http://localhost:8080/api/health

# Verificar servicios individuales
docker-compose exec app curl redis:6379
docker-compose exec app pg_isready -h postgres -U admin
```

## 🔒 Seguridad

### Configuraciones importantes

1. **JWT Secret**: Cambiar en producción
2. **Database passwords**: Usar contraseñas seguras
3. **HTTPS**: Configurar certificados SSL
4. **CORS**: Configurar dominios permitidos
5. **Rate limiting**: Ajustar límites según necesidades

### Best practices implementadas

- Validación de input
- Sanitización de datos
- Headers de seguridad
- Autenticación JWT
- Autorización basada en roles
- Logs de auditoría

## 🚀 Deployment

### Docker Swarm
```bash
# Inicializar swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml streaming-platform
```

### Kubernetes
```bash
# Aplicar manifiestos (crear primero)
kubectl apply -f k8s/

# Verificar deployment
kubectl get pods
kubectl get services
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear branch de feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Abrir Pull Request

## 📝 Roadmap

- [ ] Integración con CDN externo
- [ ] Sistema de comentarios
- [ ] Live streaming
- [ ] Subtítulos automáticos
- [ ] Recomendaciones por IA
- [ ] App móvil nativa
- [ ] Integración con LMS
- [ ] Analytics avanzados

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver `LICENSE` para más detalles.

## 👥 Equipo

- **Backend**: Arquitectura limpia con Go
- **Frontend**: PWA con Vanilla JavaScript
- **DevOps**: Docker, Nginx, Monitoreo
- **Video**: FFmpeg, HLS, Streaming

## 📞 Soporte

Para soporte técnico:
- Crear issue en GitHub
- Revisar documentación en `/docs`
- Consultar logs de aplicación

---

¡Gracias por usar StreamLearn! 🎓📹

