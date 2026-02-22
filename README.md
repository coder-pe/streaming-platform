# Streaming Platform (Go + Web)

Plataforma de streaming de tutoriales con backend en Go y frontend en JavaScript modular.

## Estado actual

El sistema actual está orientado a:

- Autenticación (JWT + refresh token).
- Catálogo de videos públicos con búsqueda por query/categoría/tags.
- Subida de video con cola de transcodificación (Redis + workers + FFmpeg).
- Reproducción HLS por calidades.
- Perfil de usuario, favoritos e historial/continuar viendo.
- Infra de soporte: PostgreSQL, Redis, MinIO, Meilisearch, Prometheus y Grafana.

El detalle de brechas respecto a objetivos de negocio (exámenes, niveles, certificados y monetización) está en `docs/ESTADO_ACTUAL.md`.

## Documentación vigente

- `docs/README_DOCS.md`: índice de documentación mantenida.
- `docs/ESTADO_ACTUAL.md`: auditoría funcional y brechas contra objetivos.
- `docs/API_ACTUAL.md`: endpoints reales del backend y notas de integración frontend.

## Requisitos

- Docker y Docker Compose.
- Go 1.21+ (opcional para ejecutar fuera de contenedores).

## Ejecución rápida

1. Levantar servicios:

```bash
docker-compose up -d
```

2. Ejecutar backend (local):

```bash
go run ./cmd/server
```

3. Ejecutar tests:

```bash
go test ./...
```

4. Ejecutar smoke test de API (con backend arriba):

```bash
./scripts/smoke_api.sh
```

## Stack técnico

- Backend: Go, Gorilla Mux, PostgreSQL, Redis.
- Streaming: FFmpeg + HLS.
- Búsqueda: Meilisearch.
- Storage: MinIO + filesystem local.
- Frontend: Vanilla JS modular + PWA.

## Nota de alcance

Aún no existen módulos de:

- Exámenes teóricos/prácticos.
- Sistema de niveles.
- Certificación por nivel.
- Suscripciones/pagos.

Esas capacidades están definidas como siguiente etapa de producto en `docs/ESTADO_ACTUAL.md`.

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
