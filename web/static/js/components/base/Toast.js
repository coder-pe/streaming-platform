/**
 * @file Toast.js
 * @description Sistema de notificaciones toast
 */

import { UI_CONFIG } from '../../config/constants.js';

class Toast {
  constructor() {
    this.container = null;
    this.init();
  }

  /**
   * Inicializar contenedor de toasts
   */
  init() {
    this.container = document.createElement('div');
    this.container.className = 'toast-container';
    this.container.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      z-index: 10000;
      pointer-events: none;
    `;
    document.body.appendChild(this.container);
  }

  /**
   * Mostrar toast
   */
  show(message, type = 'info', duration = UI_CONFIG.TOAST_DURATION) {
    const toast = this.createToast(message, type);
    this.container.appendChild(toast);

    // Animación de entrada
    setTimeout(() => {
      toast.classList.add('show');
    }, 10);

    // Auto-cerrar
    setTimeout(() => {
      this.close(toast);
    }, duration);

    return toast;
  }

  /**
   * Crear elemento toast
   */
  createToast(message, type) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.style.cssText = `
      background: var(--surface);
      color: var(--text-primary);
      padding: 16px 20px;
      margin-bottom: 10px;
      border-radius: 8px;
      box-shadow: var(--shadow-lg);
      min-width: 300px;
      max-width: 500px;
      pointer-events: auto;
      opacity: 0;
      transform: translateX(100%);
      transition: all 0.3s ease;
      display: flex;
      align-items: center;
      gap: 12px;
      border-left: 4px solid ${this.getTypeColor(type)};
    `;

    // Icono
    const icon = document.createElement('span');
    icon.className = 'toast-icon';
    icon.textContent = this.getTypeIcon(type);
    icon.style.cssText = `
      font-size: 20px;
      flex-shrink: 0;
    `;

    // Mensaje
    const messageEl = document.createElement('span');
    messageEl.className = 'toast-message';
    messageEl.textContent = message;
    messageEl.style.cssText = `
      flex: 1;
      word-break: break-word;
    `;

    // Botón de cierre
    const closeBtn = document.createElement('button');
    closeBtn.className = 'toast-close';
    closeBtn.textContent = '×';
    closeBtn.style.cssText = `
      background: none;
      border: none;
      color: var(--text-secondary);
      font-size: 24px;
      cursor: pointer;
      padding: 0;
      width: 24px;
      height: 24px;
      flex-shrink: 0;
      line-height: 1;
    `;
    closeBtn.onclick = () => this.close(toast);

    toast.appendChild(icon);
    toast.appendChild(messageEl);
    toast.appendChild(closeBtn);

    return toast;
  }

  /**
   * Cerrar toast
   */
  close(toast) {
    toast.classList.remove('show');
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';

    setTimeout(() => {
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
    }, 300);
  }

  /**
   * Obtener color según tipo
   */
  getTypeColor(type) {
    const colors = {
      success: '#10b981',
      error: '#ef4444',
      warning: '#f59e0b',
      info: '#3b82f6',
    };
    return colors[type] || colors.info;
  }

  /**
   * Obtener icono según tipo
   */
  getTypeIcon(type) {
    const icons = {
      success: '✓',
      error: '✕',
      warning: '⚠',
      info: 'ℹ',
    };
    return icons[type] || icons.info;
  }

  /**
   * Métodos helper
   */
  success(message, duration) {
    return this.show(message, 'success', duration);
  }

  error(message, duration) {
    return this.show(message, 'error', duration);
  }

  warning(message, duration) {
    return this.show(message, 'warning', duration);
  }

  info(message, duration) {
    return this.show(message, 'info', duration);
  }
}

// Agregar estilos para animación show
const style = document.createElement('style');
style.textContent = `
  .toast.show {
    opacity: 1 !important;
    transform: translateX(0) !important;
  }
`;
document.head.appendChild(style);

// Exportar instancia singleton
export default new Toast();
