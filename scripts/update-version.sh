#!/bin/bash
# Script para actualizar la versión de cache busting

# Generar timestamp o usar hash de git
if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
    # Usar commit hash corto de Git
    VERSION=$(git rev-parse --short HEAD)
    echo "Using Git hash: $VERSION"
else
    # Usar timestamp
    VERSION=$(date +"%Y.%m.%d.%H%M")
    echo "Using timestamp: $VERSION"
fi

# Actualizar version.js
VERSION_FILE="web/static/js/config/version.js"
cat > "$VERSION_FILE" << EOF
/**
 * @file version.js
 * @description Configuración de versión de la aplicación para cache busting
 */

// Actualizado automáticamente: $(date)
export const APP_VERSION = '$VERSION';

// Para uso en query strings: /static/js/app.js?v=$VERSION
export const getVersionedUrl = (url) => {
  const separator = url.includes('?') ? '&' : '?';
  return \`\${url}\${separator}v=\${APP_VERSION}\`;
};
EOF

echo "✅ Version updated to: $VERSION"

# Actualizar index.html
INDEX_FILE="web/index.html"
if [ -f "$INDEX_FILE" ]; then
    # Reemplazar versiones en CSS y JS
    sed -i.bak "s/styles\.css?v=[^\"']*/styles.css?v=$VERSION/g" "$INDEX_FILE"
    sed -i.bak "s/app\.js?v=[^\"']*/app.js?v=$VERSION/g" "$INDEX_FILE"
    rm "${INDEX_FILE}.bak" 2>/dev/null
    echo "✅ index.html updated"
fi

echo "🎉 Cache busting version updated successfully!"
