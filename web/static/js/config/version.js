/**
 * @file version.js
 * @description Configuración de versión de la aplicación para cache busting
 */

// Se actualiza automáticamente con cada build/deploy
// Formato: YYYY.MM.DD.HHMM o usar un hash git
export const APP_VERSION = '2025.01.19.1200';

// Para uso en query strings: /static/js/app.js?v=2025.01.19.1200
export const getVersionedUrl = (url) => {
  const separator = url.includes('?') ? '&' : '?';
  return `${url}${separator}v=${APP_VERSION}`;
};
