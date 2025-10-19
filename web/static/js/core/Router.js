/**
 * @file Router.js
 * @description Sistema de enrutamiento SPA
 */

import eventBus from './EventBus.js';
import state from './State.js';

class Router {
  constructor() {
    this.routes = {};
    this.currentRoute = null;
    this.defaultRoute = 'home';

    // Escuchar cambios en el hash de la URL
    window.addEventListener('hashchange', () => this.handleRouteChange());
    window.addEventListener('load', () => this.handleRouteChange());
  }

  /**
   * Registrar una ruta
   */
  register(name, handler) {
    this.routes[name] = handler;
  }

  /**
   * Navegar a una ruta
   */
  navigate(route, params = {}) {
    if (!this.routes[route]) {
      console.warn(`Route "${route}" not found, navigating to default`);
      route = this.defaultRoute;
    }

    // Actualizar hash de URL
    window.location.hash = route;

    // Ejecutar handler de la ruta
    this.executeRoute(route, params);
  }

  /**
   * Ejecutar handler de ruta
   */
  executeRoute(route, params = {}) {
    const handler = this.routes[route];
    if (!handler) return;

    // Actualizar estado
    state.setRoute(route);
    this.currentRoute = route;

    // Emitir eventos
    eventBus.emit('route:before', route, params);

    try {
      handler(params);
      eventBus.emit('route:after', route, params);
    } catch (error) {
      console.error(`Error executing route "${route}":`, error);
      eventBus.emit('route:error', route, error);
    }
  }

  /**
   * Manejar cambio de ruta desde URL
   */
  handleRouteChange() {
    const hash = window.location.hash.slice(1) || this.defaultRoute;
    const [route, queryString] = hash.split('?');
    const params = this.parseQueryString(queryString);

    this.executeRoute(route, params);
  }

  /**
   * Parsear query string
   */
  parseQueryString(queryString) {
    if (!queryString) return {};

    const params = {};
    const pairs = queryString.split('&');

    pairs.forEach(pair => {
      const [key, value] = pair.split('=');
      params[decodeURIComponent(key)] = decodeURIComponent(value || '');
    });

    return params;
  }

  /**
   * Obtener ruta actual
   */
  getCurrentRoute() {
    return this.currentRoute;
  }

  /**
   * Volver a la ruta anterior
   */
  back() {
    window.history.back();
  }

  /**
   * Establecer ruta por defecto
   */
  setDefault(route) {
    this.defaultRoute = route;
  }
}

// Exportar instancia singleton
export default new Router();
