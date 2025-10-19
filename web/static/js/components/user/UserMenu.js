/**
 * @file UserMenu.js
 * @description Componente de menú de usuario
 */

import Component from '../base/Component.js';
import authService from '../../services/AuthService.js';
import userService from '../../services/UserService.js';
import eventBus from '../../core/EventBus.js';
import state from '../../core/State.js';

export default class UserMenu extends Component {
  constructor(container) {
    super(container);
    this.state = {
      user: null,
      isOpen: false,
    };
  }

  onMount() {
    // Suscribirse a cambios de autenticación
    this.subscribe('auth:login', (user) => {
      this.setState({ user });
    });

    this.subscribe('auth:logout', () => {
      this.setState({ user: null });
    });

    this.subscribe('state:user', (user) => {
      this.setState({ user });
    });

    // Cargar usuario actual
    const user = state.get('user');
    if (user) {
      this.setState({ user });
    }
  }

  template() {
    if (!this.state.user) {
      return `
        <div class="user-menu-guest">
          <button class="btn btn-primary btn-sm" data-action="login">Iniciar Sesión</button>
          <button class="btn btn-secondary btn-sm" data-action="register">Registrarse</button>
        </div>
      `;
    }

    const avatarUrl = userService.getAvatarUrl(this.state.user.avatar);

    return `
      <div class="user-menu">
        <button class="user-menu-trigger" data-action="toggle">
          <img src="${avatarUrl}" alt="${this.state.user.firstName}" class="user-avatar">
          <span class="user-name">${this.state.user.firstName}</span>
          <i class="icon-chevron-down"></i>
        </button>

        <div class="user-menu-dropdown ${this.state.isOpen ? 'open' : ''}">
          <a href="/profile" class="menu-item" data-route="profile">
            <i class="icon-user"></i>
            Mi Perfil
          </a>
          <a href="/videos" class="menu-item" data-route="videos">
            <i class="icon-video"></i>
            Mis Videos
          </a>
          <a href="/upload" class="menu-item" data-route="upload">
            <i class="icon-upload"></i>
            Subir Video
          </a>
          <div class="menu-divider"></div>
          <button class="menu-item" data-action="logout">
            <i class="icon-logout"></i>
            Cerrar Sesión
          </button>
        </div>
      </div>
    `;
  }

  attachEventListeners() {
    // Botones para invitados
    const loginBtn = this.$('[data-action="login"]');
    if (loginBtn) {
      loginBtn.addEventListener('click', () => this.emit('menu:login'));
    }

    const registerBtn = this.$('[data-action="register"]');
    if (registerBtn) {
      registerBtn.addEventListener('click', () => this.emit('menu:register'));
    }

    // Toggle del menú
    const toggleBtn = this.$('[data-action="toggle"]');
    if (toggleBtn) {
      toggleBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.setState({ isOpen: !this.state.isOpen });
      });
    }

    // Cerrar menú al hacer click fuera
    if (this.state.user) {
      document.addEventListener('click', (e) => {
        if (this.state.isOpen && !this.container.contains(e.target)) {
          this.setState({ isOpen: false });
        }
      });
    }

    // Botón de logout
    const logoutBtn = this.$('[data-action="logout"]');
    if (logoutBtn) {
      logoutBtn.addEventListener('click', () => this.handleLogout());
    }
  }

  async handleLogout() {
    try {
      await authService.logout();
      this.setState({ isOpen: false });
    } catch (error) {
      console.error('Logout error:', error);
    }
  }

  emit(event, data) {
    eventBus.emit(event, data);
  }
}
