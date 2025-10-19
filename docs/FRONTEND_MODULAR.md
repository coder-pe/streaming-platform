# 🎨 Frontend Modular - JavaScript Moderno (ES6+)

## 📋 Resumen

Este documento describe la refactorización completa del frontend de la plataforma de streaming a una arquitectura modular moderna usando JavaScript ES6+ con módulos, componentes reutilizables y separación de responsabilidades.

---

## 🏗️ Arquitectura del Frontend

### Principios de Diseño

1. **Separación de Responsabilidades** - Cada módulo tiene una responsabilidad única
2. **Modularidad** - Código organizado en módulos ES6+ reutilizables
3. **Component-Based** - UI construida con componentes reutilizables
4. **State Management** - Estado centralizado y predecible
5. **Service Layer** - Lógica de negocio separada de la UI

---

## 📁 Estructura de Archivos

```
web/static/js/
├── app.js                      ← Punto de entrada (nuevo)
├── app.legacy.js               ← Backup del código anterior
│
├── config/                     ← Configuración
│   └── constants.js           ✅ Constantes y configuración global
│
├── core/                       ← Núcleo de la aplicación
│   ├── Router.js              □ Sistema de enrutamiento
│   ├── State.js               □ Gestión de estado global
│   └── EventBus.js            □ Sistema de eventos
│
├── services/                   ← Capa de servicios (API)
│   ├── ApiService.js          □ Cliente HTTP base
│   ├── AuthService.js         □ Autenticación y autorización
│   ├── VideoService.js        □ Operaciones con videos
│   ├── UserService.js         □ Operaciones con usuarios
│   └── StorageService.js      □ LocalStorage/SessionStorage
│
├── components/                 ← Componentes UI reutilizables
│   ├── base/                  □ Componentes base
│   │   ├── Component.js       □ Clase base para componentes
│   │   ├── Modal.js           □ Sistema de modales
│   │   ├── Toast.js           □ Notificaciones
│   │   └── Loading.js         □ Estados de carga
│   │
│   ├── auth/                  □ Componentes de autenticación
│   │   ├── LoginForm.js       □ Formulario de login
│   │   └── RegisterForm.js    □ Formulario de registro
│   │
│   ├── video/                 □ Componentes de video
│   │   ├── VideoCard.js       □ Tarjeta de video
│   │   ├── VideoGrid.js       □ Grilla de videos
│   │   ├── VideoPlayer.js     □ Reproductor de video
│   │   └── VideoUpload.js     □ Subida de videos
│   │
│   └── user/                  □ Componentes de usuario
│       ├── UserProfile.js     □ Perfil de usuario
│       └── UserMenu.js        □ Menú de usuario
│
└── utils/                      ← Utilidades
    ├── helpers.js             ✅ Funciones auxiliares
    ├── validators.js          □ Validadores
    └── dom.js                 □ Helpers del DOM

✅ = Completado
□ = Por implementar en producción
```

---

## 🎯 Características Principales

### 1. Módulos ES6+

Todos los archivos usan módulos ES6 con `import`/`export`:

```javascript
// ✅ Antes (todo en un archivo)
class StreamLearnApp {
  // 867 líneas de código...
}

// ✅ Después (modular)
import { ApiService } from './services/ApiService.js';
import { VideoCard } from './components/video/VideoCard.js';
import { formatDuration } from './utils/helpers.js';
```

### 2. Componentes Reutilizables

**Componente Base:**
```javascript
// components/base/Component.js
export class Component {
  constructor(containerId) {
    this.container = document.getElementById(containerId);
    this.state = {};
  }

  render() {
    // Método abstracto
    throw new Error('render() must be implemented');
  }

  setState(newState) {
    this.state = { ...this.state, ...newState };
    this.render();
  }

  mount() {
    this.render();
  }

  unmount() {
    if (this.container) {
      this.container.innerHTML = '';
    }
  }
}
```

**Ejemplo de Componente:**
```javascript
// components/video/VideoCard.js
import { Component } from '../base/Component.js';
import { formatDuration, formatRelativeDate } from '../../utils/helpers.js';

export class VideoCard extends Component {
  constructor(video) {
    super();
    this.video = video;
  }

  render() {
    return `
      <div class="video-card" data-video-id="${this.video.id}">
        <div class="video-thumbnail">
          <img src="${this.video.thumbnail || '/static/assets/video-placeholder.jpg'}"
               alt="${this.video.title}">
          <div class="video-duration">${formatDuration(this.video.duration)}</div>
        </div>
        <div class="video-info">
          <h4 class="video-title">${this.video.title}</h4>
          <p class="video-instructor">${this.getInstructorName()}</p>
          <div class="video-meta">
            <span>${this.video.view_count || 0} vistas</span>
            <div class="video-rating">
              <span>⭐ ${this.video.rating?.toFixed(1) || 'N/A'}</span>
            </div>
          </div>
        </div>
      </div>
    `;
  }

  getInstructorName() {
    if (!this.video.instructor) return 'Instructor';
    return `${this.video.instructor.first_name} ${this.video.instructor.last_name}`;
  }

  attachEvents() {
    this.element.addEventListener('click', () => {
      this.emit('video:select', this.video);
    });
  }
}
```

### 3. Servicios API

**Servicio Base:**
```javascript
// services/ApiService.js
import { API_CONFIG, STORAGE_KEYS } from '../config/constants.js';

export class ApiService {
  constructor(baseURL = API_CONFIG.BASE_URL) {
    this.baseURL = baseURL;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;

    const config = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...this.getAuthHeaders(),
        ...options.headers,
      },
    };

    if (options.body && typeof options.body !== 'string') {
      config.body = JSON.stringify(options.body);
    }

    try {
      const response = await fetch(url, config);
      return await this.handleResponse(response);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  getAuthHeaders() {
    const token = localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  async handleResponse(response) {
    const data = await response.json().catch(() => null);

    if (!response.ok) {
      throw {
        status: response.status,
        message: data?.message || response.statusText,
        data,
      };
    }

    return data;
  }

  handleError(error) {
    console.error('API Error:', error);
    return error;
  }

  get(endpoint, options) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  post(endpoint, body, options) {
    return this.request(endpoint, { ...options, method: 'POST', body });
  }

  put(endpoint, body, options) {
    return this.request(endpoint, { ...options, method: 'PUT', body });
  }

  delete(endpoint, options) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }
}
```

**Servicio de Videos:**
```javascript
// services/VideoService.js
import { ApiService } from './ApiService.js';

export class VideoService extends ApiService {
  async getVideos({ page = 1, limit = 12, query, category, sort } = {}) {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });

    if (query) params.append('query', query);
    if (category) params.append('category', category);
    if (sort) params.append('sort', sort);

    return this.get(`/videos?${params}`);
  }

  async getVideo(videoId) {
    return this.get(`/videos/${videoId}`);
  }

  async uploadVideo(formData) {
    return this.request('/videos/upload', {
      method: 'POST',
      body: formData,
      headers: {}, // Let browser set Content-Type for FormData
    });
  }

  async updateVideo(videoId, data) {
    return this.put(`/videos/${videoId}`, data);
  }

  async deleteVideo(videoId) {
    return this.delete(`/videos/${videoId}`);
  }

  async updateWatchProgress(videoId, position, quality) {
    return this.put(`/videos/${videoId}/progress`, { position, quality });
  }
}
```

### 4. Gestión de Estado

```javascript
// core/State.js
export class State {
  constructor(initialState = {}) {
    this.state = initialState;
    this.listeners = new Map();
  }

  getState() {
    return { ...this.state };
  }

  setState(updates) {
    const prevState = this.state;
    this.state = { ...this.state, ...updates };
    this.notify(prevState, this.state);
  }

  subscribe(key, callback) {
    if (!this.listeners.has(key)) {
      this.listeners.set(key, []);
    }
    this.listeners.get(key).push(callback);

    // Retorna función para desuscribirse
    return () => {
      const callbacks = this.listeners.get(key);
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    };
  }

  notify(prevState, newState) {
    this.listeners.forEach((callbacks, key) => {
      if (prevState[key] !== newState[key]) {
        callbacks.forEach(callback => callback(newState[key], prevState[key]));
      }
    });
  }
}

// Uso:
const appState = new State({
  user: null,
  videos: [],
  currentVideo: null,
  isLoading: false,
});

// Suscribirse a cambios
appState.subscribe('user', (newUser, oldUser) => {
  console.log('User changed:', newUser);
  updateUI(newUser);
});

// Actualizar estado
appState.setState({ user: userData });
```

### 5. Sistema de Enrutamiento

```javascript
// core/Router.js
export class Router {
  constructor(routes) {
    this.routes = routes;
    this.currentRoute = null;
    this.init();
  }

  init() {
    window.addEventListener('hashchange', () => this.handleRoute());
    window.addEventListener('load', () => this.handleRoute());
  }

  handleRoute() {
    const hash = window.location.hash.slice(1) || 'home';
    const [route, ...params] = hash.split('/');

    if (this.routes[route]) {
      this.currentRoute = route;
      this.routes[route](params);
    } else {
      this.navigate('home');
    }
  }

  navigate(route, ...params) {
    const path = params.length ? `${route}/${params.join('/')}` : route;
    window.location.hash = path;
  }
}

// Uso:
const router = new Router({
  home: () => renderHomePage(),
  videos: (params) => renderVideosPage(params),
  video: (params) => renderVideoPlayer(params[0]),
  profile: () => renderProfilePage(),
  upload: () => renderUploadPage(),
});
```

---

## 🎨 Sistema de Componentes

### Toast Notifications

```javascript
// components/base/Toast.js
export class Toast {
  static show(message, type = 'info', duration = 5000) {
    const container = document.getElementById('toastContainer');
    if (!container) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
      <div class="toast-content">
        <p>${message}</p>
      </div>
    `;

    container.appendChild(toast);

    // Auto remove
    setTimeout(() => {
      toast.style.opacity = '0';
      setTimeout(() => toast.remove(), 300);
    }, duration);

    // Manual close
    toast.addEventListener('click', () => {
      toast.style.opacity = '0';
      setTimeout(() => toast.remove(), 300);
    });
  }

  static success(message) {
    return Toast.show(message, 'success');
  }

  static error(message) {
    return Toast.show(message, 'error');
  }

  static warning(message) {
    return Toast.show(message, 'warning');
  }

  static info(message) {
    return Toast.show(message, 'info');
  }
}
```

### Modal System

```javascript
// components/base/Modal.js
export class Modal {
  constructor(id, options = {}) {
    this.id = id;
    this.options = {
      closeOnOutsideClick: true,
      closeOnEscape: true,
      ...options,
    };
    this.element = document.getElementById(id);
    this.attachEvents();
  }

  show() {
    if (!this.element) return;

    this.element.classList.add('active');
    document.body.style.overflow = 'hidden';

    if (this.options.onShow) {
      this.options.onShow();
    }
  }

  hide() {
    if (!this.element) return;

    this.element.classList.remove('active');
    document.body.style.overflow = '';

    if (this.options.onHide) {
      this.options.onHide();
    }
  }

  attachEvents() {
    if (!this.element) return;

    // Close button
    const closeBtn = this.element.querySelector('.modal-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.hide());
    }

    // Click outside
    if (this.options.closeOnOutsideClick) {
      this.element.addEventListener('click', (e) => {
        if (e.target === this.element) {
          this.hide();
        }
      });
    }

    // Escape key
    if (this.options.closeOnEscape) {
      document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && this.element.classList.contains('active')) {
          this.hide();
        }
      });
    }
  }
}
```

---

## 📦 Nuevo app.js (Punto de Entrada)

```javascript
// app.js - Aplicación principal refactorizada
import { State } from './core/State.js';
import { Router } from './core/Router.js';
import { EventBus } from './core/EventBus.js';

import { AuthService } from './services/AuthService.js';
import { VideoService } from './services/VideoService.js';
import { UserService } from './services/UserService.js';

import { Toast } from './components/base/Toast.js';
import { Modal } from './components/base/Modal.js';

class StreamLearnApp {
  constructor() {
    // Inicializar servicios
    this.authService = new AuthService();
    this.videoService = new VideoService();
    this.userService = new UserService();

    // Inicializar estado global
    this.state = new State({
      user: null,
      videos: [],
      currentVideo: null,
      isLoading: false,
      filters: {
        query: '',
        category: '',
        sort: 'newest',
      },
    });

    // Inicializar event bus
    this.eventBus = new EventBus();

    // Inicializar router
    this.router = new Router({
      home: () => this.renderHome(),
      videos: () => this.renderVideos(),
      video: (params) => this.renderVideo(params[0]),
      profile: () => this.renderProfile(),
      upload: () => this.renderUpload(),
    });

    this.init();
  }

  async init() {
    await this.checkAuth();
    this.setupGlobalEvents();
    this.setupStateSubscriptions();
  }

  async checkAuth() {
    try {
      const user = await this.authService.getCurrentUser();
      this.state.setState({ user });
    } catch (error) {
      console.log('Not authenticated');
    }
  }

  setupGlobalEvents() {
    // Suscribirse a eventos globales
    this.eventBus.on('auth:login', (user) => {
      this.state.setState({ user });
      Toast.success('Sesión iniciada correctamente');
    });

    this.eventBus.on('auth:logout', () => {
      this.state.setState({ user: null });
      this.router.navigate('home');
      Toast.success('Sesión cerrada');
    });

    this.eventBus.on('video:upload:success', () => {
      Toast.success('Video subido exitosamente');
      this.router.navigate('profile');
    });
  }

  setupStateSubscriptions() {
    // Reaccionar a cambios de estado
    this.state.subscribe('user', (newUser) => {
      this.updateUIForAuth(newUser);
    });

    this.state.subscribe('isLoading', (isLoading) => {
      const overlay = document.getElementById('loadingOverlay');
      if (overlay) {
        overlay.classList.toggle('hidden', !isLoading);
      }
    });
  }

  updateUIForAuth(user) {
    const loginBtn = document.getElementById('loginBtn');
    const registerBtn = document.getElementById('registerBtn');
    const userMenu = document.getElementById('userMenu');
    const authRequired = document.querySelectorAll('.auth-required');

    if (user) {
      // Usuario autenticado
      if (loginBtn) loginBtn.style.display = 'none';
      if (registerBtn) registerBtn.style.display = 'none';
      if (userMenu) userMenu.classList.remove('hidden');
      authRequired.forEach(el => el.style.display = 'block');
    } else {
      // Usuario no autenticado
      if (loginBtn) loginBtn.style.display = 'inline-flex';
      if (registerBtn) registerBtn.style.display = 'inline-flex';
      if (userMenu) userMenu.classList.add('hidden');
      authRequired.forEach(el => el.style.display = 'none');
    }
  }

  async renderHome() {
    // Implementar renderizado de home
  }

  async renderVideos() {
    // Implementar renderizado de videos
  }

  async renderVideo(videoId) {
    // Implementar renderizado de video player
  }

  async renderProfile() {
    // Implementar renderizado de perfil
  }

  async renderUpload() {
    // Implementar renderizado de upload
  }
}

// Inicializar app cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
  window.app = new StreamLearnApp();
});

export default StreamLearnApp;
```

---

## 🔄 Comparación: Antes vs Después

### Antes (Monolítico)

```
✅ Un solo archivo (app.js)
❌ 867 líneas de código
❌ Difícil de mantener
❌ No reutilizable
❌ Todo acoplado
❌ Difícil de testear
```

### Después (Modular)

```
✅ Múltiples módulos organizados
✅ ~100-200 líneas por archivo
✅ Fácil de mantener
✅ Componentes reutilizables
✅ Separación de responsabilidades
✅ Fácil de testear
✅ Tree-shaking (solo importas lo que usas)
✅ Mejor rendimiento
```

---

## 📊 Beneficios de la Refactorización

### 1. Mantenibilidad ⬆️
- Código organizado por responsabilidad
- Fácil encontrar y modificar código
- Menos bugs por aislamiento

### 2. Reutilización ⬆️
- Componentes reutilizables
- Servicios compartidos
- Helpers globales

### 3. Escalabilidad ⬆️
- Fácil agregar nuevas features
- Estructura clara para crecer
- No hay límite de tamaño

### 4. Testabilidad ⬆️
- Módulos independientes
- Fácil crear mocks
- Unit tests por módulo

### 5. Performance ⬆️
- Code splitting natural
- Tree shaking
- Carga lazy de módulos

---

## 🚀 Próximos Pasos

### Implementaciones Futuras

1. **TypeScript** - Tipado estático para mayor seguridad
2. **Build Process** - Webpack/Vite para bundling
3. **Testing** - Jest/Vitest para unit tests
4. **PWA Completo** - Service Worker mejorado
5. **Virtual DOM** - Para mejor rendimiento en listas grandes
6. **State Persistence** - IndexedDB para offline
7. **i18n** - Internacionalización
8. **A11y** - Accesibilidad mejorada

---

## 📚 Referencias

- [MDN Web Docs - Módulos JavaScript](https://developer.mozilla.org/es/docs/Web/JavaScript/Guide/Modules)
- [ES6 Features](http://es6-features.org/)
- [JavaScript Design Patterns](https://www.patterns.dev/)
- [Clean Code JavaScript](https://github.com/ryanmcdermott/clean-code-javascript)

---

**Estado**: 🟡 En Progreso (Documentación completa, implementación parcial)

**Fecha**: 2024

**Arquitectura**: Modular ES6+ con componentes, servicios y estado centralizado
