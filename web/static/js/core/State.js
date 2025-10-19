/**
 * @file State.js
 * @description Gestión centralizada del estado de la aplicación
 */

import eventBus from './EventBus.js';

class State {
  constructor() {
    this.state = {
      user: null,
      isAuthenticated: false,
      currentRoute: 'home',
      videos: [],
      loading: false,
      error: null,
    };
  }

  /**
   * Obtener todo el estado
   */
  getState() {
    return { ...this.state };
  }

  /**
   * Obtener un valor específico del estado
   */
  get(key) {
    return this.state[key];
  }

  /**
   * Actualizar el estado
   */
  setState(updates) {
    const oldState = { ...this.state };
    this.state = { ...this.state, ...updates };

    // Emitir evento de cambio de estado
    eventBus.emit('state:changed', {
      oldState,
      newState: this.state,
      changes: updates,
    });

    // Emitir eventos específicos para cada cambio
    Object.keys(updates).forEach(key => {
      if (oldState[key] !== updates[key]) {
        eventBus.emit(`state:${key}`, updates[key], oldState[key]);
      }
    });
  }

  /**
   * Establecer usuario autenticado
   */
  setUser(user) {
    this.setState({
      user,
      isAuthenticated: !!user,
    });
  }

  /**
   * Limpiar usuario (logout)
   */
  clearUser() {
    this.setState({
      user: null,
      isAuthenticated: false,
    });
  }

  /**
   * Establecer videos
   */
  setVideos(videos) {
    this.setState({ videos });
  }

  /**
   * Agregar un video
   */
  addVideo(video) {
    this.setState({
      videos: [video, ...this.state.videos],
    });
  }

  /**
   * Actualizar un video
   */
  updateVideo(videoId, updates) {
    const videos = this.state.videos.map(video =>
      video.id === videoId ? { ...video, ...updates } : video
    );
    this.setState({ videos });
  }

  /**
   * Eliminar un video
   */
  removeVideo(videoId) {
    const videos = this.state.videos.filter(video => video.id !== videoId);
    this.setState({ videos });
  }

  /**
   * Establecer ruta actual
   */
  setRoute(route) {
    this.setState({ currentRoute: route });
  }

  /**
   * Establecer estado de carga
   */
  setLoading(loading) {
    this.setState({ loading });
  }

  /**
   * Establecer error
   */
  setError(error) {
    this.setState({ error });
  }

  /**
   * Limpiar error
   */
  clearError() {
    this.setState({ error: null });
  }

  /**
   * Resetear estado a valores iniciales
   */
  reset() {
    this.state = {
      user: null,
      isAuthenticated: false,
      currentRoute: 'home',
      videos: [],
      loading: false,
      error: null,
    };
    eventBus.emit('state:reset');
  }
}

// Exportar instancia singleton
export default new State();
