/**
 * @file Loading.js
 * @description Componente de indicador de carga
 */

class Loading {
  constructor() {
    this.container = null;
    this.isVisible = false;
    this.init();
  }

  /**
   * Inicializar contenedor de loading
   */
  init() {
    this.container = document.createElement('div');
    this.container.className = 'loading-overlay';
    this.container.innerHTML = `
      <div class="loading-spinner">
        <div class="spinner"></div>
        <p class="loading-text">Cargando...</p>
      </div>
    `;

    this.container.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.7);
      display: none;
      justify-content: center;
      align-items: center;
      z-index: 9999;
    `;

    const spinner = this.container.querySelector('.loading-spinner');
    spinner.style.cssText = `
      text-align: center;
      color: var(--text-primary);
    `;

    // Agregar estilos del spinner
    this.addSpinnerStyles();

    document.body.appendChild(this.container);
  }

  /**
   * Agregar estilos del spinner
   */
  addSpinnerStyles() {
    if (document.getElementById('loading-styles')) return;

    const style = document.createElement('style');
    style.id = 'loading-styles';
    style.textContent = `
      .spinner {
        border: 4px solid rgba(255, 255, 255, 0.1);
        border-left-color: var(--primary);
        border-radius: 50%;
        width: 50px;
        height: 50px;
        animation: spin 1s linear infinite;
        margin: 0 auto 16px;
      }

      @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
      }

      .loading-text {
        margin: 0;
        font-size: 16px;
        color: var(--text-primary);
      }
    `;
    document.head.appendChild(style);
  }

  /**
   * Mostrar loading
   */
  show(text = 'Cargando...') {
    if (this.isVisible) return;

    const textEl = this.container.querySelector('.loading-text');
    if (textEl) {
      textEl.textContent = text;
    }

    this.container.style.display = 'flex';
    this.isVisible = true;
  }

  /**
   * Ocultar loading
   */
  hide() {
    if (!this.isVisible) return;

    this.container.style.display = 'none';
    this.isVisible = false;
  }

  /**
   * Actualizar texto
   */
  setText(text) {
    const textEl = this.container.querySelector('.loading-text');
    if (textEl) {
      textEl.textContent = text;
    }
  }
}

// Exportar instancia singleton
export default new Loading();
