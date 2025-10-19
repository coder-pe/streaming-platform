/**
 * @file EventBus.js
 * @description Sistema de eventos para comunicación entre componentes
 */

class EventBus {
  constructor() {
    this.events = {};
  }

  /**
   * Suscribirse a un evento
   */
  on(event, callback) {
    if (!this.events[event]) {
      this.events[event] = [];
    }
    this.events[event].push(callback);

    // Retornar función para desuscribirse
    return () => this.off(event, callback);
  }

  /**
   * Suscribirse a un evento solo una vez
   */
  once(event, callback) {
    const onceCallback = (...args) => {
      callback(...args);
      this.off(event, onceCallback);
    };
    return this.on(event, onceCallback);
  }

  /**
   * Desuscribirse de un evento
   */
  off(event, callback) {
    if (!this.events[event]) return;

    if (!callback) {
      // Si no se proporciona callback, eliminar todos los listeners
      delete this.events[event];
      return;
    }

    this.events[event] = this.events[event].filter(cb => cb !== callback);
  }

  /**
   * Emitir un evento
   */
  emit(event, ...args) {
    if (!this.events[event]) return;

    this.events[event].forEach(callback => {
      try {
        callback(...args);
      } catch (error) {
        console.error(`Error in event handler for "${event}":`, error);
      }
    });
  }

  /**
   * Limpiar todos los eventos
   */
  clear() {
    this.events = {};
  }
}

// Exportar instancia singleton
export default new EventBus();
