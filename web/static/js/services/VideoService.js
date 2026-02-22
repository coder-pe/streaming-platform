/**
 * @file VideoService.js
 * @description Servicio para operaciones con videos
 */

import apiService from './ApiService.js';
import eventBus from '../core/EventBus.js';
import state from '../core/State.js';

class VideoService {
  /**
   * Obtener todos los videos
   */
  async getVideos(params = {}) {
    try {
      const response = await apiService.get('/videos', params);

      if (response.videos) {
        state.setVideos(response.videos);
      }

      return response;
    } catch (error) {
      console.error('Error fetching videos:', error);
      throw error;
    }
  }

  /**
   * Obtener video por ID
   */
  async getVideo(videoId) {
    try {
      const response = await apiService.get(`/videos/${videoId}`);
      return response.video || response;
    } catch (error) {
      console.error('Error fetching video:', error);
      throw error;
    }
  }

  /**
   * Buscar videos
   */
  async searchVideos(query, params = {}) {
    try {
      const response = await apiService.get('/videos', {
        query,
        ...params,
      });

      if (response.videos) {
        state.setVideos(response.videos);
      }

      return response;
    } catch (error) {
      console.error('Error searching videos:', error);
      throw error;
    }
  }

  /**
   * Subir video
   */
  async uploadVideo(file, metadata, onProgress) {
    try {
      const formData = new FormData();
      formData.append('video', file);
      formData.append('metadata', JSON.stringify({
        title: metadata.title || '',
        description: metadata.description || '',
        category: metadata.category || 'other',
        tags: metadata.tags || [],
        is_public: metadata.is_public !== false,
      }));

      const response = await apiService.upload('/videos', formData, onProgress);

      if (response.video) {
        state.addVideo(response.video);
        eventBus.emit('video:uploaded', response.video);
      }

      return response;
    } catch (error) {
      console.error('Error uploading video:', error);
      throw error;
    }
  }

  /**
   * Actualizar video
   */
  async updateVideo(videoId, updates) {
    try {
      const response = await apiService.put(`/videos/${videoId}`, updates);

      if (response.video) {
        state.updateVideo(videoId, response.video);
        eventBus.emit('video:updated', response.video);
      }

      return response;
    } catch (error) {
      console.error('Error updating video:', error);
      throw error;
    }
  }

  /**
   * Eliminar video
   */
  async deleteVideo(videoId) {
    try {
      await apiService.delete(`/videos/${videoId}`);

      state.removeVideo(videoId);
      eventBus.emit('video:deleted', videoId);

      return true;
    } catch (error) {
      console.error('Error deleting video:', error);
      throw error;
    }
  }

  /**
   * Obtener videos del usuario
   */
  async getUserVideos(userId) {
    try {
      const response = await apiService.get(`/users/${userId}/videos`);
      return response.videos || response;
    } catch (error) {
      console.error('Error fetching user videos:', error);
      throw error;
    }
  }

  /**
   * Dar like a un video
   */
  async likeVideo(videoId) {
    try {
      const response = await apiService.post(`/favorites/${videoId}`);

      if (response.video) {
        state.updateVideo(videoId, response.video);
      }

      eventBus.emit('video:liked', videoId);
      return response;
    } catch (error) {
      console.error('Error liking video:', error);
      throw error;
    }
  }

  /**
   * Quitar like de un video
   */
  async unlikeVideo(videoId) {
    try {
      const response = await apiService.delete(`/favorites/${videoId}`);

      if (response.video) {
        state.updateVideo(videoId, response.video);
      }

      eventBus.emit('video:unliked', videoId);
      return response;
    } catch (error) {
      console.error('Error unliking video:', error);
      throw error;
    }
  }

  /**
   * Incrementar vistas
   */
  async incrementViews(videoId) {
    try {
      await apiService.post(`/videos/${videoId}/view`);
      eventBus.emit('video:viewed', videoId);
    } catch (error) {
      console.error('Error incrementing views:', error);
    }
  }

  /**
   * Obtener URL de streaming
   */
  getStreamUrl(videoId, quality = '720p') {
    return `${apiService.baseURL}/stream/${videoId}/${quality}/playlist.m3u8`;
  }

  /**
   * Obtener URL de thumbnail
   */
  getThumbnailUrl(thumbnailPath) {
    if (!thumbnailPath) return null;
    return thumbnailPath.startsWith('http')
      ? thumbnailPath
      : `${apiService.baseURL}/static/${thumbnailPath}`;
  }
}

// Exportar instancia singleton
export default new VideoService();
