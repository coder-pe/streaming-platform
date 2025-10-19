/**
 * @file AuthService.js
 * @description Servicio de autenticación y autorización
 */

import apiService from './ApiService.js';
import storageService from './StorageService.js';
import eventBus from '../core/EventBus.js';
import state from '../core/State.js';

class AuthService {
  /**
   * Registrar nuevo usuario
   */
  async register(username, email, password) {
    try {
      const response = await apiService.post('/auth/register', {
        username,
        email,
        password,
      });

      // Guardar tokens y datos de usuario
      if (response.token) {
        this.saveAuthData(response);
      }

      return response;
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    }
  }

  /**
   * Iniciar sesión
   */
  async login(email, password) {
    try {
      const response = await apiService.post('/auth/login', {
        email,
        password,
      });

      // Guardar tokens y datos de usuario
      if (response.token) {
        this.saveAuthData(response);
      }

      return response;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  }

  /**
   * Cerrar sesión
   */
  async logout() {
    try {
      // Intentar cerrar sesión en el servidor
      await apiService.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Limpiar datos locales siempre
      this.clearAuthData();
    }
  }

  /**
   * Refrescar token
   */
  async refreshToken() {
    try {
      const refreshToken = storageService.getRefreshToken();
      if (!refreshToken) {
        throw new Error('No refresh token available');
      }

      const response = await apiService.post('/auth/refresh', {
        refresh_token: refreshToken,
      });

      if (response.token) {
        storageService.setAuthToken(response.token);
        if (response.refresh_token) {
          storageService.setRefreshToken(response.refresh_token);
        }
      }

      return response;
    } catch (error) {
      console.error('Token refresh error:', error);
      this.clearAuthData();
      throw error;
    }
  }

  /**
   * Obtener perfil del usuario actual
   */
  async getCurrentUser() {
    try {
      const response = await apiService.get('/auth/me');

      if (response.user) {
        storageService.setUserData(response.user);
        state.setUser(response.user);
      }

      return response.user;
    } catch (error) {
      console.error('Get current user error:', error);
      throw error;
    }
  }

  /**
   * Verificar si el usuario está autenticado
   */
  isAuthenticated() {
    return storageService.hasActiveSession();
  }

  /**
   * Obtener usuario actual del estado
   */
  getUser() {
    return state.get('user') || storageService.getUserData();
  }

  /**
   * Guardar datos de autenticación
   */
  saveAuthData(authData) {
    if (authData.token) {
      storageService.setAuthToken(authData.token);
    }
    if (authData.refresh_token) {
      storageService.setRefreshToken(authData.refresh_token);
    }
    if (authData.user) {
      storageService.setUserData(authData.user);
      state.setUser(authData.user);
    }

    // Emitir evento de autenticación exitosa
    eventBus.emit('auth:login', authData.user);
  }

  /**
   * Limpiar datos de autenticación
   */
  clearAuthData() {
    storageService.clearSession();
    state.clearUser();

    // Emitir evento de cierre de sesión
    eventBus.emit('auth:logout');
  }

  /**
   * Inicializar autenticación al cargar la app
   */
  async initialize() {
    if (this.isAuthenticated()) {
      try {
        await this.getCurrentUser();
      } catch (error) {
        // Si falla, intentar refrescar token
        try {
          await this.refreshToken();
          await this.getCurrentUser();
        } catch (refreshError) {
          // Si falla el refresh, limpiar sesión
          this.clearAuthData();
        }
      }
    }
  }
}

// Exportar instancia singleton
export default new AuthService();
