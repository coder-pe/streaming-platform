/**
 * @file UserService.js
 * @description Servicio para operaciones con usuarios
 */

import apiService from './ApiService.js';
import storageService from './StorageService.js';
import eventBus from '../core/EventBus.js';
import state from '../core/State.js';

class UserService {
  /**
   * Obtener perfil de usuario
   */
  async getProfile(userId) {
    try {
      const response = await apiService.get(`/users/${userId}`);
      return response.user || response;
    } catch (error) {
      console.error('Error fetching profile:', error);
      throw error;
    }
  }

  /**
   * Actualizar perfil del usuario actual
   */
  async updateProfile(updates) {
    try {
      const user = state.get('user');
      if (!user) throw new Error('No authenticated user');

      const response = await apiService.put(`/users/${user.id}`, updates);

      if (response.user) {
        storageService.setUserData(response.user);
        state.setUser(response.user);
        eventBus.emit('user:updated', response.user);
      }

      return response;
    } catch (error) {
      console.error('Error updating profile:', error);
      throw error;
    }
  }

  /**
   * Actualizar avatar
   */
  async updateAvatar(file, onProgress) {
    try {
      const user = state.get('user');
      if (!user) throw new Error('No authenticated user');

      const formData = new FormData();
      formData.append('avatar', file);

      const response = await apiService.upload(
        `/users/${user.id}/avatar`,
        formData,
        onProgress
      );

      if (response.user) {
        storageService.setUserData(response.user);
        state.setUser(response.user);
        eventBus.emit('user:avatar-updated', response.user);
      }

      return response;
    } catch (error) {
      console.error('Error updating avatar:', error);
      throw error;
    }
  }

  /**
   * Cambiar contraseña
   */
  async changePassword(currentPassword, newPassword) {
    try {
      const user = state.get('user');
      if (!user) throw new Error('No authenticated user');

      const response = await apiService.put(`/users/${user.id}/password`, {
        current_password: currentPassword,
        new_password: newPassword,
      });

      eventBus.emit('user:password-changed');
      return response;
    } catch (error) {
      console.error('Error changing password:', error);
      throw error;
    }
  }

  /**
   * Obtener estadísticas del usuario
   */
  async getStats(userId) {
    try {
      const response = await apiService.get(`/users/${userId}/stats`);
      return response;
    } catch (error) {
      console.error('Error fetching stats:', error);
      throw error;
    }
  }

  /**
   * Seguir usuario
   */
  async followUser(userId) {
    try {
      const response = await apiService.post(`/users/${userId}/follow`);
      eventBus.emit('user:followed', userId);
      return response;
    } catch (error) {
      console.error('Error following user:', error);
      throw error;
    }
  }

  /**
   * Dejar de seguir usuario
   */
  async unfollowUser(userId) {
    try {
      const response = await apiService.delete(`/users/${userId}/follow`);
      eventBus.emit('user:unfollowed', userId);
      return response;
    } catch (error) {
      console.error('Error unfollowing user:', error);
      throw error;
    }
  }

  /**
   * Obtener seguidores
   */
  async getFollowers(userId) {
    try {
      const response = await apiService.get(`/users/${userId}/followers`);
      return response.followers || response;
    } catch (error) {
      console.error('Error fetching followers:', error);
      throw error;
    }
  }

  /**
   * Obtener siguiendo
   */
  async getFollowing(userId) {
    try {
      const response = await apiService.get(`/users/${userId}/following`);
      return response.following || response;
    } catch (error) {
      console.error('Error fetching following:', error);
      throw error;
    }
  }

  /**
   * Obtener usuario actual
   */
  getUser() {
    return state.get('user') || storageService.getUserData();
  }

  /**
   * Obtener URL de avatar
   */
  getAvatarUrl(avatarPath) {
    if (!avatarPath) {
      return '/static/images/default-avatar.png';
    }
    return avatarPath.startsWith('http')
      ? avatarPath
      : `${apiService.baseURL}/static/${avatarPath}`;
  }
}

// Exportar instancia singleton
export default new UserService();
