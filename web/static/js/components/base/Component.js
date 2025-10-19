/**
 * @file Component.js
 * @description Clase base para todos los componentes
 */

import eventBus from '../../core/EventBus.js';

export default class Component {
  constructor(container) {
    this.container = typeof container === 'string'
      ? document.querySelector(container)
      : container;

    this.state = {};
    this.listeners = [];
    this.mounted = false;
  }

  /**
   * Establecer estado del componente
   */
  setState(updates) {
    const oldState = { ...this.state };
    this.state = { ...this.state, ...updates };

    // Llamar hook de actualización
    this.onStateChange(oldState, this.state);

    // Re-renderizar si está montado
    if (this.mounted) {
      this.render();
    }
  }

  /**
   * Hook llamado cuando cambia el estado
   */
  onStateChange(oldState, newState) {
    // Override en subclases si es necesario
  }

  /**
   * Renderizar componente
   */
  render() {
    if (!this.container) return;

    const html = this.template();
    this.container.innerHTML = html;

    // Adjuntar event listeners después de renderizar
    this.attachEventListeners();
  }

  /**
   * Template del componente (debe ser implementado en subclases)
   */
  template() {
    return '';
  }

  /**
   * Adjuntar event listeners (debe ser implementado en subclases)
   */
  attachEventListeners() {
    // Override en subclases
  }

  /**
   * Suscribirse a eventos del EventBus
   */
  subscribe(event, callback) {
    const unsubscribe = eventBus.on(event, callback.bind(this));
    this.listeners.push(unsubscribe);
    return unsubscribe;
  }

  /**
   * Montar componente
   */
  mount() {
    this.mounted = true;
    this.onMount();
    this.render();
  }

  /**
   * Hook llamado cuando el componente se monta
   */
  onMount() {
    // Override en subclases
  }

  /**
   * Desmontar componente
   */
  unmount() {
    this.mounted = false;

    // Limpiar event listeners del EventBus
    this.listeners.forEach(unsubscribe => unsubscribe());
    this.listeners = [];

    // Limpiar contenedor
    if (this.container) {
      this.container.innerHTML = '';
    }

    // Llamar hook de desmontaje
    this.onUnmount();
  }

  /**
   * Hook llamado cuando el componente se desmonta
   */
  onUnmount() {
    // Override en subclases
  }

  /**
   * Mostrar componente
   */
  show() {
    if (this.container) {
      this.container.style.display = '';
      this.container.classList.remove('hidden');
    }
  }

  /**
   * Ocultar componente
   */
  hide() {
    if (this.container) {
      this.container.style.display = 'none';
      this.container.classList.add('hidden');
    }
  }

  /**
   * Alternar visibilidad
   */
  toggle() {
    if (this.container) {
      if (this.container.style.display === 'none') {
        this.show();
      } else {
        this.hide();
      }
    }
  }

  /**
   * Buscar elemento dentro del contenedor
   */
  $(selector) {
    return this.container ? this.container.querySelector(selector) : null;
  }

  /**
   * Buscar todos los elementos dentro del contenedor
   */
  $$(selector) {
    return this.container ? this.container.querySelectorAll(selector) : [];
  }
}
