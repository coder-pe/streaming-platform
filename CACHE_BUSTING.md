# Cache Busting - Guía de Uso

## ¿Qué es Cache Busting?

Cache Busting es una técnica para forzar al navegador a descargar la versión más reciente de tus archivos estáticos (CSS, JS) en lugar de usar la versión en caché.

## ¿Por qué lo necesitamos?

Cuando actualizas tu código JavaScript o CSS, los navegadores pueden seguir usando la versión antigua almacenada en caché, causando:
- ❌ Características nuevas que no funcionan
- ❌ Bugs que ya corregiste pero que siguen apareciendo
- ❌ Usuarios que necesitan hacer "Ctrl+F5" o limpiar caché manualmente

## Solución Implementada

### 1. Versionado con Query Strings

Todos los archivos estáticos tienen un parámetro `?v=VERSION`:

```html
<!-- Antes -->
<link rel="stylesheet" href="/static/css/styles.css">
<script src="/static/js/app.js" type="module"></script>

<!-- Después -->
<link rel="stylesheet" href="/static/css/styles.css?v=2025.01.19.1200">
<script src="/static/js/app.js?v=2025.01.19.1200" type="module"></script>
```

Cuando cambias la versión, el navegador ve una URL diferente y descarga el archivo nuevo.

### 2. Headers de Cache Control

El servidor configura headers HTTP para controlar el caché:

| Tipo de Archivo | Cache Control | Duración |
|-----------------|---------------|----------|
| HTML (`index.html`) | `no-cache, no-store, must-revalidate` | Nunca cachear |
| Archivos versionados (`?v=`) | `public, max-age=31536000, immutable` | 1 año |
| CSS/JS sin versión | `public, max-age=86400` | 1 día |
| Imágenes, fuentes | `public, max-age=604800` | 1 semana |

## Cómo Usar

### Opción 1: Script Automático (Recomendado)

```bash
# Actualizar versión y ejecutar
make dev

# O solo actualizar versión
make update-version
```

El script automáticamente:
1. Genera una versión (hash Git o timestamp)
2. Actualiza `version.js`
3. Actualiza `index.html` con la nueva versión

### Opción 2: Manual

1. Edita `web/static/js/config/version.js`:
```javascript
export const APP_VERSION = '2025.01.19.1300'; // Cambiar aquí
```

2. Actualiza `web/index.html`:
```html
<link rel="stylesheet" href="/static/css/styles.css?v=2025.01.19.1300">
<script src="/static/js/app.js?v=2025.01.19.1300" type="module"></script>
```

### Opción 3: Durante el Desarrollo

Si estás desarrollando activamente:

1. **Deshabilitar caché en DevTools:**
   - Abre DevTools (F12)
   - Ve a la pestaña Network
   - Marca "Disable cache"
   - Mantén DevTools abierto mientras desarrollas

2. **Hard Refresh:**
   - Windows/Linux: `Ctrl + Shift + R` o `Ctrl + F5`
   - Mac: `Cmd + Shift + R`

## Flujo de Trabajo Recomendado

### Durante el Desarrollo
```bash
# Desarrollo normal
go run cmd/server/main.go

# Con caché deshabilitado en DevTools
```

### Antes de Commit/Deploy
```bash
# Actualizar versión y compilar
make build

# O con tests
make deploy
```

### Workflow Git
```bash
# Hacer cambios
git add .
git commit -m "Update frontend features"

# Actualizar versión (usa git hash)
./scripts/update-version.sh

# Commit la nueva versión
git add web/static/js/config/version.js web/index.html
git commit -m "Update cache busting version"
git push
```

## Verificar que Funciona

### 1. Ver Headers en el Navegador

1. Abre DevTools (F12)
2. Ve a la pestaña Network
3. Recarga la página
4. Haz clic en `styles.css` o `app.js`
5. Verifica los headers de respuesta:

```
Cache-Control: public, max-age=31536000, immutable
```

### 2. Probar Actualización

1. Cambia algo en tu CSS (ej: color de fondo)
2. Ejecuta `make update-version`
3. Reinicia el servidor
4. Recarga el navegador **una sola vez** (F5 normal)
5. Deberías ver los cambios inmediatamente

## Comandos Útiles del Makefile

```bash
make help           # Ver todos los comandos disponibles
make build          # Compilar (actualiza versión automáticamente)
make dev            # Desarrollo (actualiza versión y ejecuta)
make update-version # Solo actualizar versión
make clean          # Limpiar archivos compilados
make deploy         # Tests + Build para producción
```

## Troubleshooting

### "Los cambios no se reflejan después de actualizar la versión"

1. Verifica que la versión cambió en `index.html`:
```bash
grep "?v=" web/index.html
```

2. Verifica los headers de respuesta en DevTools

3. Prueba con hard refresh: `Ctrl + Shift + R`

### "El script update-version.sh no funciona"

```bash
# Dar permisos de ejecución
chmod +x scripts/update-version.sh

# Ejecutar directamente
./scripts/update-version.sh
```

### "Quiero usar mi propio formato de versión"

Edita `scripts/update-version.sh` y cambia la línea de VERSION:

```bash
# Ejemplo: usar número incremental
VERSION="1.2.3"

# Ejemplo: fecha personalizada
VERSION=$(date +"%Y%m%d-%H%M")
```

## Integración con CI/CD

### GitHub Actions

```yaml
- name: Update cache busting version
  run: |
    ./scripts/update-version.sh

- name: Build
  run: make build
```

### GitLab CI

```yaml
build:
  script:
    - ./scripts/update-version.sh
    - make build
```

## Mejores Prácticas

✅ **DO:**
- Actualizar versión antes de cada deploy
- Usar `make dev` durante desarrollo
- Commitear cambios de versión junto con tus cambios
- Verificar que los usuarios ven la nueva versión después del deploy

❌ **DON'T:**
- No uses la misma versión para múltiples deploys
- No edites manualmente los archivos si usas el script
- No olvides actualizar la versión antes de deploy

## Referencias

- [MDN - HTTP Caching](https://developer.mozilla.org/en-US/docs/Web/HTTP/Caching)
- [Google - Cache Busting](https://web.dev/http-cache/)
