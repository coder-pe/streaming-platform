# API actual (backend)

Fecha de corte: 2026-02-21.

Base URL: `/api`

## Públicas

- `POST /auth/login`
- `POST /auth/register`
- `POST /auth/refresh`
- `GET /videos`
- `GET /videos/{id}`
- `POST /videos/{id}/view`
- `GET /users/{id}`
- `GET /users/{id}/stats`

## Protegidas (JWT)

- `GET /user/profile`
- `PUT /user/profile`
- `POST /auth/change-password`
- `GET /user/stats`
- `POST /users/{id}/avatar`
- `GET /user/settings`
- `PUT /user/settings`
- `GET /favorites`
- `POST /favorites/{videoId}`
- `DELETE /favorites/{videoId}`
- `GET /history`
- `POST /history/{videoId}/progress`
- `GET /continue-watching`
- `POST /videos`
- `PUT /videos/{id}`
- `DELETE /videos/{id}`
- `GET /stream/{id}/master.m3u8`
- `GET /stream/{id}/{quality}/playlist.m3u8`
- `GET /stream/{id}/{quality}/{segment}`

## Desalineaciones detectadas en frontend

- `web/static/js/services/VideoService.js` usa rutas no existentes:
  - `GET /videos/search` (backend usa `GET /videos?query=...`)
  - `POST /videos/upload` (backend usa `POST /videos`)
- `web/static/js/services/UserService.js` usa rutas no existentes:
  - `PUT /users/{id}`
  - `PUT /users/{id}/password`
  - `POST/DELETE /users/{id}/follow`
  - `GET /users/{id}/followers`
  - `GET /users/{id}/following`

## Nota

No existen endpoints de exámenes, niveles, certificados o pagos en el estado actual.
