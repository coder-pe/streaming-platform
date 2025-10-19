/**
 * @file RegisterForm.js
 * @description Formulario de registro
 */

import Component from '../base/Component.js';
import authService from '../../services/AuthService.js';
import toast from '../base/Toast.js';
import loading from '../base/Loading.js';
import { isValidEmail, isStrongPassword } from '../../utils/helpers.js';

export default class RegisterForm extends Component {
  constructor(container) {
    super(container);
    this.state = {
      username: '',
      email: '',
      password: '',
      confirmPassword: '',
      errors: {},
    };
  }

  template() {
    return `
      <form class="auth-form" id="registerForm">
        <div class="form-group">
          <label for="registerUsername">Nombre de usuario</label>
          <input
            type="text"
            id="registerUsername"
            name="username"
            class="form-control ${this.state.errors.username ? 'error' : ''}"
            placeholder="usuario123"
            value="${this.state.username}"
            required
          />
          ${this.state.errors.username ? `<span class="error-message">${this.state.errors.username}</span>` : ''}
        </div>

        <div class="form-group">
          <label for="registerEmail">Correo electrónico</label>
          <input
            type="email"
            id="registerEmail"
            name="email"
            class="form-control ${this.state.errors.email ? 'error' : ''}"
            placeholder="correo@ejemplo.com"
            value="${this.state.email}"
            required
          />
          ${this.state.errors.email ? `<span class="error-message">${this.state.errors.email}</span>` : ''}
        </div>

        <div class="form-group">
          <label for="registerPassword">Contraseña</label>
          <input
            type="password"
            id="registerPassword"
            name="password"
            class="form-control ${this.state.errors.password ? 'error' : ''}"
            placeholder="********"
            value="${this.state.password}"
            required
          />
          ${this.state.errors.password ? `<span class="error-message">${this.state.errors.password}</span>` : ''}
        </div>

        <div class="form-group">
          <label for="registerConfirmPassword">Confirmar contraseña</label>
          <input
            type="password"
            id="registerConfirmPassword"
            name="confirmPassword"
            class="form-control ${this.state.errors.confirmPassword ? 'error' : ''}"
            placeholder="********"
            value="${this.state.confirmPassword}"
            required
          />
          ${this.state.errors.confirmPassword ? `<span class="error-message">${this.state.errors.confirmPassword}</span>` : ''}
        </div>

        <button type="submit" class="btn btn-primary btn-block">
          Registrarse
        </button>
      </form>
    `;
  }

  attachEventListeners() {
    const form = this.$('#registerForm');
    if (form) {
      form.addEventListener('submit', (e) => this.handleSubmit(e));
    }

    // Input change listeners
    ['username', 'email', 'password', 'confirmPassword'].forEach(field => {
      const input = this.$(`#register${field.charAt(0).toUpperCase() + field.slice(1)}`);
      if (input) {
        input.addEventListener('input', (e) => {
          this.setState({
            [field]: e.target.value,
            errors: { ...this.state.errors, [field]: '' }
          });
        });
      }
    });
  }

  validate() {
    const errors = {};

    if (!this.state.username) {
      errors.username = 'El nombre de usuario es requerido';
    } else if (this.state.username.length < 3) {
      errors.username = 'El nombre de usuario debe tener al menos 3 caracteres';
    }

    if (!this.state.email) {
      errors.email = 'El correo es requerido';
    } else if (!isValidEmail(this.state.email)) {
      errors.email = 'Correo inválido';
    }

    if (!this.state.password) {
      errors.password = 'La contraseña es requerida';
    } else if (!isStrongPassword(this.state.password)) {
      errors.password = 'La contraseña debe tener al menos 8 caracteres, una mayúscula, una minúscula, un número y un carácter especial';
    }

    if (!this.state.confirmPassword) {
      errors.confirmPassword = 'Confirma tu contraseña';
    } else if (this.state.password !== this.state.confirmPassword) {
      errors.confirmPassword = 'Las contraseñas no coinciden';
    }

    this.setState({ errors });
    return Object.keys(errors).length === 0;
  }

  async handleSubmit(e) {
    e.preventDefault();

    if (!this.validate()) {
      return;
    }

    try {
      loading.show('Creando cuenta...');

      await authService.register(
        this.state.username,
        this.state.email,
        this.state.password
      );

      toast.success('¡Cuenta creada correctamente!');

      // Limpiar formulario
      this.setState({
        username: '',
        email: '',
        password: '',
        confirmPassword: '',
        errors: {}
      });

      // Emitir evento de registro exitoso
      this.emit('register:success');
    } catch (error) {
      toast.error(error.message || 'Error al crear cuenta');
    } finally {
      loading.hide();
    }
  }

  emit(event, data) {
    this.container.dispatchEvent(new CustomEvent(event, { detail: data }));
  }
}
