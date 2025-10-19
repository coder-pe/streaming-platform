/**
 * @file Router.js
 * @description Sistema de enrutamiento SPA con History API
 */

import eventBus from './EventBus.js';
import state from './State.js';

class Router {
  constructor() {
    this.routes = {};
    this.currentRoute = null;
    this.defaultRoute = 'home';
    this.useHistoryAPI = true; // true = URLs limpias, false = hash URLs

    // Escuchar cambios de navegación
    if (this.useHistoryAPI) {
      // History API - URLs limpias (sin #)
      window.addEventListener('popstate', (e) => this.handleRouteChange());
      window.addEventListener('load', () => this.handleRouteChange());
    } else {
      // Hash-based routing - URLs con #
      window.addEventListener('hashchange', () => this.handleRouteChange());
      window.addEventListener('load', () => this.handleRouteChange());
    }
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

    if (this.useHistoryAPI) {
      // History API - pushState para URLs limpias
      const url = route === this.defaultRoute ? '/' : `/${route}`;
      window.history.pushState({ route, params }, '', url);
      this.executeRoute(route, params);
    } else {
      // Hash-based routing
      window.location.hash = route;
      this.executeRoute(route, params);
    }
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
    let route, queryString;

    if (this.useHistoryAPI) {
      // History API - leer desde pathname
      const path = window.location.pathname;
      const pathParts = path.split('?');
      const routePath = pathParts[0].slice(1) || this.defaultRoute; // Remove leading /
      queryString = pathParts[1];
      route = routePath || this.defaultRoute;
    } else {
      // Hash-based - leer desde hash
      const hash = window.location.hash.slice(1) || this.defaultRoute;
      const hashParts = hash.split('?');
      route = hashParts[0];
      queryString = hashParts[1];
    }

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

  /**
   * Obtener URL para una ruta (útil para links)
   */
  getUrl(route) {
    if (this.useHistoryAPI) {
      return route === this.defaultRoute ? '/' : `/${route}`;
    } else {
      return `#${route}`;
    }
  }

  /**
   * Habilitar/deshabilitar History API
   */
  setUseHistoryAPI(value) {
    this.useHistoryAPI = value;
  }
}

// Exportar instancia singleton
export default new Router();
