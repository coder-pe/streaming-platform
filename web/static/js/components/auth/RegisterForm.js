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
      firstName: '',
      lastName: '',
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
          <label for="registerFirstName">Nombre</label>
          <input
            type="text"
            id="registerFirstName"
            name="firstName"
            class="form-control ${this.state.errors.firstName? 'error' : ''}"
            placeholder="John"
            value="${this.state.firstName}"
            required
          />
          ${this.state.errors.firstName? `<span class="error-message">${this.state.errors.firstName}</span>` : ''}
        </div>

        <div class="form-group">
          <label for="registerLastName">Apellido</label>
          <input
            type="text"
            id="registerLastName"
            name="firstName"
            class="form-control ${this.state.errors.lastName? 'error' : ''}"
            placeholder="Smith"
            value="${this.state.lastName}"
            required
          />
          ${this.state.errors.lastName? `<span class="error-message">${this.state.errors.lastName}</span>` : ''}
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

    // Input change listeners - solo actualizar estado sin re-render
    ['firstName', 'lastName', 'email', 'password', 'confirmPassword'].forEach(field => {
      const input = this.$(`#register${field.charAt(0).toUpperCase() + field.slice(1)}`);
      if (input) {
        input.addEventListener('input', (e) => {
          this.state[field] = e.target.value;
          this.state.errors[field] = '';
        });
      }
    });
  }

  validate() {
    const errors = {};

    if (!this.state.firstName) {
      errors.firstName = 'El nombre es requerido';
    } else if (this.state.firstName.length < 3) {
      errors.firstName = 'El nombre debe tener al menos 3 caracteres';
    }

    if (!this.state.lastName) {
      errors.lastName = 'El apellido es requerido';
    } else if (this.state.lastName.length < 3) {
      errors.lastName = 'El apellido debe tener al menos 3 caracteres';
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
      loading.show('Creando cuenta...');

      await authService.register(
        this.state.firstName,
        this.state.lastName,
        this.state.email,
        this.state.password
      );

      toast.success('¡Cuenta creada correctamente!');

      // Limpiar formulario
      this.setState({
        firstName: '',
        lastName: '',
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
