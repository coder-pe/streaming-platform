/**
 * @file LoginForm.js
 * @description Formulario de inicio de sesión
 */

import Component from '../base/Component.js';
import authService from '../../services/AuthService.js';
import toast from '../base/Toast.js';
import loading from '../base/Loading.js';
import { isValidEmail } from '../../utils/helpers.js';

export default class LoginForm extends Component {
  constructor(container) {
    super(container);
    this.state = {
      email: '',
      password: '',
      errors: {},
    };
  }

  template() {
    return `
      <form class="auth-form" id="loginForm">
        <div class="form-group">
          <label for="loginEmail">Correo electrónico</label>
          <input
            type="email"
            id="loginEmail"
            name="email"
            class="form-control ${this.state.errors.email ? 'error' : ''}"
            placeholder="correo@ejemplo.com"
            value="${this.state.email}"
            required
          />
          ${this.state.errors.email ? `<span class="error-message">${this.state.errors.email}</span>` : ''}
        </div>

        <div class="form-group">
          <label for="loginPassword">Contraseña</label>
          <input
            type="password"
            id="loginPassword"
            name="password"
            class="form-control ${this.state.errors.password ? 'error' : ''}"
            placeholder="********"
            value="${this.state.password}"
            required
          />
          ${this.state.errors.password ? `<span class="error-message">${this.state.errors.password}</span>` : ''}
        </div>

        <button type="submit" class="btn btn-primary btn-block">
          Iniciar Sesión
        </button>

        <div class="form-footer">
          <a href="#" class="link" id="forgotPasswordLink">¿Olvidaste tu contraseña?</a>
        </div>
      </form>
    `;
  }

  attachEventListeners() {
    const form = this.$('#loginForm');
    if (form) {
      form.addEventListener('submit', (e) => this.handleSubmit(e));
    }

    // Input change listeners - solo actualizar estado sin re-render
    const emailInput = this.$('#loginEmail');
    const passwordInput = this.$('#loginPassword');

    if (emailInput) {
      emailInput.addEventListener('input', (e) => {
        this.state.email = e.target.value;
        this.state.errors.email = '';
      });
    }

    if (passwordInput) {
      passwordInput.addEventListener('input', (e) => {
        this.state.password = e.target.value;
        this.state.errors.password = '';
      });
    }
  }

  validate() {
    const errors = {};

    if (!this.state.email) {
      errors.email = 'El correo es requerido';
    } else if (!isValidEmail(this.state.email)) {
      errors.email = 'Correo inválido';
    }

    if (!this.state.password) {
      errors.password = 'La contraseña es requerida';
    } else if (this.state.password.length < 6) {
      errors.password = 'La contraseña debe tener al menos 6 caracteres';
    }

    // Solo re-renderizar si hay errores
    if (Object.keys(errors).length > 0) {
      this.setState({ errors });
    }

    return Object.keys(errors).length === 0;
  }

  async handleSubmit(e) {
    e.preventDefault();

    if (!this.validate()) {
      return;
    }

    try {
      loading.show('Iniciando sesión...');

      await authService.login(this.state.email, this.state.password);

      // toast.success('¡Sesión iniciada correctamente!');

      // Limpiar formulario
      this.setState({ email: '', password: '', errors: {} });

      // Emitir evento de login exitoso
      this.emit('login:success');
    } catch (error) {
      toast.error(error.message || 'Error al iniciar sesión');
    } finally {
      loading.hide();
    }
  }

  emit(event, data) {
    this.container.dispatchEvent(new CustomEvent(event, { detail: data }));
  }

  /**
   * Hook llamado cuando el componente se desmonta
   */
  onUnmount() {
    // Limpiar estado del formulario
    this.state = {
      email: '',
      password: '',
      errors: {},
    };
  }
}
