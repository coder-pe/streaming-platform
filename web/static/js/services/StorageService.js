/**
 * @file StorageService.js
 * @description Servicio para manejo de localStorage con funciones helper
 */

import { STORAGE_KEYS } from '../config/constants.js';

class StorageService {
  /**
   * Guardar item en localStorage
   */
  setItem(key, value) {
    try {
      const serialized = JSON.stringify(value);
      localStorage.setItem(key, serialized);
      return true;
    } catch (error) {
      console.error('Error saving to localStorage:', error);
      return false;
    }
  }

  /**
   * Obtener item de localStorage
   */
  getItem(key, defaultValue = null) {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : defaultValue;
    } catch (error) {
      console.error('Error reading from localStorage:', error);
      return defaultValue;
    }
  }

  /**
   * Eliminar item de localStorage
   */
  removeItem(key) {
    try {
      localStorage.removeItem(key);
      return true;
    } catch (error) {
      console.error('Error removing from localStorage:', error);
      return false;
    }
  }

  /**
   * Limpiar todo el localStorage
   */
  clear() {
    try {
      localStorage.clear();
      return true;
    } catch (error) {
      console.error('Error clearing localStorage:', error);
      return false;
    }
  }

  /**
   * Guardar token de autenticación
   */
  setAuthToken(token) {
    return this.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
  }

  /**
   * Obtener token de autenticación
   */
  getAuthToken() {
    return this.getItem(STORAGE_KEYS.AUTH_TOKEN);
  }

  /**
   * Eliminar token de autenticación
   */
  removeAuthToken() {
    return this.removeItem(STORAGE_KEYS.AUTH_TOKEN);
  }

  /**
   * Guardar refresh token
   */
  setRefreshToken(token) {
    return this.setItem(STORAGE_KEYS.REFRESH_TOKEN, token);
  }

  /**
   * Obtener refresh token
   */
  getRefreshToken() {
    return this.getItem(STORAGE_KEYS.REFRESH_TOKEN);
  }

  /**
   * Eliminar refresh token
   */
  removeRefreshToken() {
    return this.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
  }

  /**
   * Guardar datos de usuario
   */
  setUserData(userData) {
    return this.setItem(STORAGE_KEYS.USER_DATA, userData);
  }

  /**
   * Obtener datos de usuario
   */
  getUserData() {
    return this.getItem(STORAGE_KEYS.USER_DATA);
  }

  /**
   * Eliminar datos de usuario
   */
  removeUserData() {
    return this.removeItem(STORAGE_KEYS.USER_DATA);
  }

  /**
   * Guardar preferencias de usuario
   */
  setPreferences(preferences) {
    return this.setItem(STORAGE_KEYS.PREFERENCES, preferences);
  }

  /**
   * Obtener preferencias de usuario
   */
  getPreferences() {
    return this.getItem(STORAGE_KEYS.PREFERENCES, {});
  }

  /**
   * Actualizar preferencias (merge)
   */
  updatePreferences(updates) {
    const current = this.getPreferences();
    const updated = { ...current, ...updates };
    return this.setPreferences(updated);
  }

  /**
   * Limpiar todos los datos de sesión
   */
  clearSession() {
    this.removeAuthToken();
    this.removeRefreshToken();
    this.removeUserData();
  }

  /**
   * Verificar si hay sesión activa
   */
  hasActiveSession() {
    return !!this.getAuthToken();
  }
}

// Exportar instancia singleton
export default new StorageService();
