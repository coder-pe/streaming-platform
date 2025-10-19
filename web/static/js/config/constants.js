/**
 * @file constants.js
 * @description Constantes de configuración de la aplicación
 */

export const API_CONFIG = {
  BASE_URL: '/api',
  TIMEOUT: 30000, // 30 segundos
  MAX_RETRIES: 3,
};

export const STORAGE_KEYS = {
  AUTH_TOKEN: 'authToken',
  REFRESH_TOKEN: 'refreshToken',
  USER_DATA: 'userData',
  PREFERENCES: 'userPreferences',
};

export const VIDEO_CONFIG = {
  MAX_FILE_SIZE: 100 * 1024 * 1024, // 100MB
  ACCEPTED_FORMATS: ['video/mp4', 'video/webm', 'video/ogg'],
  DEFAULT_QUALITY: '720p',
  PLAYBACK_RATES: [0.5, 1, 1.25, 1.5, 2],
};

export const UI_CONFIG = {
  TOAST_DURATION: 5000,
  MODAL_ANIMATION_DURATION: 300,
  DEBOUNCE_DELAY: 300,
  PAGINATION_LIMIT: 12,
};

export const ROUTES = {
  HOME: 'home',
  VIDEOS: 'videos',
  UPLOAD: 'upload',
  PROFILE: 'profile',
};

export const VIDEO_STATUS = {
  UPLOADING: 'uploading',
  PROCESSING: 'processing',
  READY: 'ready',
  FAILED: 'failed',
};

export const USER_ROLES = {
  ADMIN: 'admin',
  INSTRUCTOR: 'instructor',
  USER: 'user',
};

export default {
  API_CONFIG,
  STORAGE_KEYS,
  VIDEO_CONFIG,
  UI_CONFIG,
  ROUTES,
  VIDEO_STATUS,
  USER_ROLES,
};
