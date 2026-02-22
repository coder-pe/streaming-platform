# Estado actual y brecha de producto

Fecha de auditoría: 2026-02-21.

## Objetivo de negocio solicitado

Construir una plataforma tipo AlgoExpert pero más completa, con:

- Exámenes teóricos y prácticos.
- Niveles por progreso.
- Certificados por nivel.
- Mecanismos de motivación para completar rutas completas.
- Monetización sostenible.

## Veredicto

La implementación actual no cumple todavía esos objetivos de negocio.

## Qué sí está implementado

- Gestión de usuarios: registro/login/refresh.
- Catálogo de videos públicos con filtros.
- Subida y procesamiento de video (cola + transcodificación).
- Streaming HLS por calidades.
- Perfil, favoritos, historial y continuar viendo.
- Infraestructura base (PostgreSQL, Redis, MinIO, Meilisearch, Prometheus, Grafana).

## Qué falta para cumplir el objetivo

- Exámenes teóricos: no hay entidades, tablas, servicios ni endpoints.
- Exámenes prácticos: no hay sandbox/evaluador/submisiones.
- Niveles: no existe modelo de nivel, XP o reglas de desbloqueo.
- Certificados: no existe emisión, verificación ni almacenamiento de certificados.
- Monetización: no hay planes, suscripciones, pagos, facturación ni control de acceso por plan.
- Motivación: no hay sistema de streaks, metas, recompensas o rutas gamificadas.

## Hallazgos relevantes de implementación

- El backend expone rutas funcionales para video/streaming/auth/perfil, pero no para producto educativo avanzado.
- El frontend contiene llamadas a endpoints que no existen en el backend actual (por ejemplo búsqueda en `/videos/search` y upload en `/videos/upload`), lo que indica desalineación de integración.
- Hay un TODO explícito en control de acceso de streaming para compra/suscripción sin implementar.
- El script SQL base tiene inconsistencia de nombres de columnas en el `INSERT` del usuario admin (`firstName`/`lastName` vs `first_name`/`last_name`).

## Prioridad recomendada (orden)

1. Corregir desalineación frontend-backend y estabilizar flujos actuales.
2. Implementar dominio educativo: cursos, módulos, lecciones, intentos, evaluación, progreso.
3. Agregar niveles y reglas de avance.
4. Agregar certificados por nivel con validación pública.
5. Integrar monetización (planes, suscripción, pagos webhooks).
6. Añadir gamificación (streaks, objetivos semanales, recompensas).

## Definición mínima de éxito (MVP alineado a tu objetivo)

- Cada curso tiene evaluación teórica y práctica.
- El estudiante no avanza de nivel sin cumplir criterios.
- Al aprobar un nivel se genera certificado verificable.
- El acceso a contenido premium depende de una suscripción activa.
- El panel del alumno muestra progreso de ruta completa y señales de motivación.
