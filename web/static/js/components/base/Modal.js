/**
 * @file Modal.js
 * @description Sistema de modales reutilizable
 */

import Component from './Component.js';
import { UI_CONFIG } from '../../config/constants.js';

export default class Modal extends Component {
  constructor(options = {}) {
    super(null);

    this.options = {
      title: options.title || '',
      content: options.content || '',
      footer: options.footer || '',
      closeOnBackdrop: options.closeOnBackdrop !== false,
      closeOnEscape: options.closeOnEscape !== false,
      size: options.size || 'medium', // small, medium, large
      onClose: options.onClose || null,
      ...options,
    };

    this.isOpen = false;
    this.createModal();
  }

  /**
   * Crear elemento modal
   */
  createModal() {
    // Crear contenedor del modal
    this.container = document.createElement('div');
    this.container.className = 'modal';
    this.container.innerHTML = this.template();

    // Añadir al DOM pero oculto
    document.body.appendChild(this.container);
    this.hide();

    // Adjuntar event listeners
    this.attachEventListeners();
  }

  /**
   * Template del modal
   */
  template() {
    return `
      <div class="modal-backdrop"></div>
      <div class="modal-dialog modal-${this.options.size}">
        <div class="modal-content">
          <div class="modal-header">
            ${this.options.title ? `<h3 class="modal-title">${this.options.title}</h3>` : '<h3 class="modal-title"></h3>'}
            <button class="modal-close" aria-label="Cerrar">&times;</button>
          </div>
          <div class="modal-body">
            ${this.options.content}
          </div>
          ${this.options.footer ? `
            <div class="modal-footer">
              ${this.options.footer}
            </div>
          ` : ''}
        </div>
      </div>
    `;
  }

  /**
   * Adjuntar event listeners
   */
  attachEventListeners() {
    // Cerrar con botón X
    const closeBtn = this.$('.modal-close');
    if (closeBtn) {
      closeBtn.addEventListener('click', () => this.close());
    }

    // Cerrar con backdrop
    if (this.options.closeOnBackdrop) {
      const backdrop = this.$('.modal-backdrop');
      if (backdrop) {
        backdrop.addEventListener('click', () => this.close());
      }
    }

    // Cerrar con tecla Escape
    if (this.options.closeOnEscape) {
      this.handleEscape = (e) => {
        if (e.key === 'Escape' && this.isOpen) {
          this.close();
        }
      };
      document.addEventListener('keydown', this.handleEscape);
    }
  }

  /**
   * Abrir modal
   */
  open() {
    if (this.isOpen) return;

    this.isOpen = true;
    this.show();

    // Agregar clase al body para prevenir scroll
    document.body.style.overflow = 'hidden';

    // Animación
    setTimeout(() => {
      this.container.classList.add('active');
    }, 10);

    return this;
  }

  /**
   * Cerrar modal
   */
  close() {
    if (!this.isOpen) return;

    this.isOpen = false;
    this.container.classList.remove('active');

    // Esperar animación antes de ocultar
    setTimeout(() => {
      this.hide();
      document.body.style.overflow = '';

      // Callback de cierre
      if (this.options.onClose) {
        this.options.onClose();
      }
    }, UI_CONFIG.MODAL_ANIMATION_DURATION);

    return this;
  }

  /**
   * Actualizar contenido
   */
  setContent(content) {
    const body = this.$('.modal-body');
    if (body) {
      body.innerHTML = content;
    }
    return this;
  }

  /**
   * Actualizar título
   */
  setTitle(title) {
    const titleEl = this.$('.modal-title');
    if (titleEl) {
      titleEl.textContent = title;
    }
    return this;
  }

  /**
   * Actualizar footer
   */
  setFooter(footer) {
    let footerEl = this.$('.modal-footer');
    if (!footerEl && footer) {
      const content = this.$('.modal-content');
      footerEl = document.createElement('div');
      footerEl.className = 'modal-footer';
      content.appendChild(footerEl);
    }
    if (footerEl) {
      footerEl.innerHTML = footer;
    }
    return this;
  }

  /**
   * Destruir modal
   */
  destroy() {
    // Remover event listener de escape
    if (this.handleEscape) {
      document.removeEventListener('keydown', this.handleEscape);
    }

    // Remover del DOM
    if (this.container && this.container.parentNode) {
      this.container.parentNode.removeChild(this.container);
    }

    // Restaurar scroll del body
    document.body.style.overflow = '';
  }

  /**
   * Mostrar componente (override)
   */
  show() {
    if (this.container) {
      this.container.style.display = 'block';
    }
  }

  /**
   * Ocultar componente (override)
   */
  hide() {
    if (this.container) {
      this.container.style.display = 'none';
    }
  }
}
