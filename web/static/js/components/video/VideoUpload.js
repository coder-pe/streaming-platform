/**
 * @file VideoUpload.js
 * @description Componente de carga de videos
 */

import Component from '../base/Component.js';
import videoService from '../../services/VideoService.js';
import toast from '../base/Toast.js';
import { VIDEO_CONFIG } from '../../config/constants.js';
import { formatFileSize } from '../../utils/helpers.js';

export default class VideoUpload extends Component {
  constructor(container) {
    super(container);
    this.state = {
      file: null,
      title: '',
      description: '',
      category: '',
      uploading: false,
      progress: 0,
      errors: {},
    };
  }

  template() {
    return `
      <div class="video-upload-container">
        <form id="uploadForm" class="upload-form">
          <div class="upload-area ${this.state.file ? 'has-file' : ''}">
            <input
              type="file"
              id="videoFile"
              accept="${VIDEO_CONFIG.ACCEPTED_FORMATS.join(',')}"
              ${this.state.uploading ? 'disabled' : ''}
            />
            <label for="videoFile" class="upload-label">
              ${this.state.file ? `
                <i class="icon-check"></i>
                <p>${this.state.file.name}</p>
                <small>${formatFileSize(this.state.file.size)}</small>
              ` : `
                <i class="icon-upload"></i>
                <p>Arrastra un video aquí o haz clic para seleccionar</p>
                <small>Formatos permitidos: MP4, WebM, OGG (Máx. ${formatFileSize(VIDEO_CONFIG.MAX_FILE_SIZE)})</small>
              `}
            </label>
          </div>

          ${this.state.file ? `
            <div class="form-group">
              <label for="videoTitle">Título *</label>
              <input
                type="text"
                id="videoTitle"
                class="form-control ${this.state.errors.title ? 'error' : ''}"
                placeholder="Título del video"
                value="${this.state.title}"
                ${this.state.uploading ? 'disabled' : ''}
                required
              />
              ${this.state.errors.title ? `<span class="error-message">${this.state.errors.title}</span>` : ''}
            </div>

            <div class="form-group">
              <label for="videoDescription">Descripción</label>
              <textarea
                id="videoDescription"
                class="form-control"
                placeholder="Descripción del video"
                rows="4"
                ${this.state.uploading ? 'disabled' : ''}
              >${this.state.description}</textarea>
            </div>

            <div class="form-group">
              <label for="videoCategory">Categoría</label>
              <select id="videoCategory" class="form-control" ${this.state.uploading ? 'disabled' : ''}>
                <option value="">Seleccionar categoría</option>
                <option value="education">Educación</option>
                <option value="entertainment">Entretenimiento</option>
                <option value="technology">Tecnología</option>
                <option value="sports">Deportes</option>
                <option value="music">Música</option>
                <option value="other">Otro</option>
              </select>
            </div>

            ${this.state.uploading ? `
              <div class="upload-progress">
                <div class="progress-bar">
                  <div class="progress-fill" style="width: ${this.state.progress}%"></div>
                </div>
                <p class="progress-text">${Math.round(this.state.progress)}% completado</p>
              </div>
            ` : ''}

            <div class="form-actions">
              <button
                type="button"
                class="btn btn-secondary"
                id="cancelBtn"
                ${this.state.uploading ? 'disabled' : ''}
              >
                Cancelar
              </button>
              <button
                type="submit"
                class="btn btn-primary"
                ${this.state.uploading ? 'disabled' : ''}
              >
                ${this.state.uploading ? 'Subiendo...' : 'Subir video'}
              </button>
            </div>
          ` : ''}
        </form>
      </div>
    `;
  }

  attachEventListeners() {
    const fileInput = this.$('#videoFile');
    if (fileInput) {
      fileInput.addEventListener('change', (e) => this.handleFileSelect(e));
    }

    const form = this.$('#uploadForm');
    if (form) {
      form.addEventListener('submit', (e) => this.handleSubmit(e));
    }

    const titleInput = this.$('#videoTitle');
    if (titleInput) {
      titleInput.addEventListener('input', (e) => {
        this.setState({ title: e.target.value, errors: { ...this.state.errors, title: '' } });
      });
    }

    const descInput = this.$('#videoDescription');
    if (descInput) {
      descInput.addEventListener('input', (e) => {
        this.setState({ description: e.target.value });
      });
    }

    const categorySelect = this.$('#videoCategory');
    if (categorySelect) {
      categorySelect.addEventListener('change', (e) => {
        this.setState({ category: e.target.value });
      });
    }

    const cancelBtn = this.$('#cancelBtn');
    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.handleCancel());
    }

    // Drag and drop
    const uploadArea = this.$('.upload-area');
    if (uploadArea) {
      uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('drag-over');
      });

      uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('drag-over');
      });

      uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('drag-over');

        const files = e.dataTransfer.files;
        if (files.length > 0) {
          this.validateAndSetFile(files[0]);
        }
      });
    }
  }

  handleFileSelect(e) {
    const file = e.target.files[0];
    if (file) {
      this.validateAndSetFile(file);
    }
  }

  validateAndSetFile(file) {
    // Validar tamaño
    if (file.size > VIDEO_CONFIG.MAX_FILE_SIZE) {
      toast.error(`El archivo es demasiado grande. Máximo ${formatFileSize(VIDEO_CONFIG.MAX_FILE_SIZE)}`);
      return;
    }

    // Validar formato
    if (!VIDEO_CONFIG.ACCEPTED_FORMATS.includes(file.type)) {
      toast.error('Formato de video no soportado');
      return;
    }

    this.setState({ file });
  }

  validate() {
    const errors = {};

    if (!this.state.title) {
      errors.title = 'El título es requerido';
    } else if (this.state.title.length < 3) {
      errors.title = 'El título debe tener al menos 3 caracteres';
    }

    this.setState({ errors });
    return Object.keys(errors).length === 0;
  }

  async handleSubmit(e) {
    e.preventDefault();

    if (!this.state.file) {
      toast.error('Selecciona un archivo de video');
      return;
    }

    if (!this.validate()) {
      return;
    }

    try {
      this.setState({ uploading: true, progress: 0 });

      const metadata = {
        title: this.state.title,
        description: this.state.description,
        category: this.state.category,
      };

      await videoService.uploadVideo(
        this.state.file,
        metadata,
        (progress) => {
          this.setState({ progress });
        }
      );

      toast.success('¡Video subido correctamente!');

      // Resetear formulario
      this.setState({
        file: null,
        title: '',
        description: '',
        category: '',
        uploading: false,
        progress: 0,
        errors: {},
      });

      // Emitir evento de éxito
      this.emit('upload:success');
    } catch (error) {
      toast.error(error.message || 'Error al subir el video');
      this.setState({ uploading: false, progress: 0 });
    }
  }

  handleCancel() {
    this.setState({
      file: null,
      title: '',
      description: '',
      category: '',
      errors: {},
    });
  }

  emit(event, data) {
    this.container.dispatchEvent(new CustomEvent(event, {
      detail: data,
      bubbles: true
    }));
  }
}
