/**
 * @file UserProfile.js
 * @description Componente de perfil de usuario
 */

import Component from '../base/Component.js';
import userService from '../../services/UserService.js';
import toast from '../base/Toast.js';
import loading from '../base/Loading.js';

export default class UserProfile extends Component {
  constructor(container) {
    super(container);
    this.state = {
      user: null,
      editing: false,
      stats: null,
    };
  }

  async onMount() {
    await this.loadProfile();
  }

  async loadProfile() {
    try {
      loading.show('Cargando perfil...');
      const user = userService.getUser();

      if (user) {
        const stats = await userService.getStats(user.id);
        this.setState({ user, stats });
      }
    } catch (error) {
      toast.error('Error al cargar el perfil');
    } finally {
      loading.hide();
    }
  }

  template() {
    if (!this.state.user) {
      return '<div class="profile-loading">Cargando perfil...</div>';
    }

    const avatarUrl = userService.getAvatarUrl(this.state.user.avatar);

    return `
      <div class="user-profile">
        <div class="profile-header">
          <div class="profile-avatar">
            <img src="${avatarUrl}" alt="${this.state.user.firstName || this.state.user.email}">
            ${this.state.editing ? `
              <button class="btn-change-avatar" data-action="change-avatar">
                <i class="icon-camera"></i>
              </button>
            ` : ''}
          </div>

          <div class="profile-info">
            ${this.state.editing ? `
              <input
                type="text"
                id="firstNameInput"
                class="form-control"
                value="${this.state.user.firstName || ''}"
              />
              <input
                type="text"
                id="lastNameInput"
                class="form-control"
                value="${this.state.user.lastName || ''}"
              />
              <textarea
                id="bioInput"
                class="form-control"
                placeholder="Escribe tu bio..."
                rows="3"
              >${this.state.user.bio || ''}</textarea>
            ` : `
              <h1>${this.state.user.firstName || ''} ${this.state.user.lastName || ''}</h1>
              <p>${this.state.user.email}</p>
              ${this.state.user.bio ? `<p class="profile-bio">${this.state.user.bio}</p>` : ''}
            `}

            <div class="profile-stats">
              <div class="stat">
                <span class="stat-value">${this.state.stats?.totalVideos || 0}</span>
                <span class="stat-label">Videos</span>
              </div>
              <div class="stat">
                <span class="stat-value">${this.state.stats?.favoritesCount || 0}</span>
                <span class="stat-label">Favoritos</span>
              </div>
              <div class="stat">
                <span class="stat-value">${this.state.stats?.videosWatched || 0}</span>
                <span class="stat-label">Vistos</span>
              </div>
            </div>

            <div class="profile-actions">
              ${this.state.editing ? `
                <button class="btn btn-primary" data-action="save">Guardar</button>
                <button class="btn btn-secondary" data-action="cancel">Cancelar</button>
              ` : `
                <button class="btn btn-primary" data-action="edit">Editar perfil</button>
              `}
            </div>
          </div>
        </div>
      </div>
    `;
  }

  attachEventListeners() {
    const editBtn = this.$('[data-action="edit"]');
    if (editBtn) {
      editBtn.addEventListener('click', () => this.toggleEdit(true));
    }

    const saveBtn = this.$('[data-action="save"]');
    if (saveBtn) {
      saveBtn.addEventListener('click', () => this.handleSave());
    }

    const cancelBtn = this.$('[data-action="cancel"]');
    if (cancelBtn) {
      cancelBtn.addEventListener('click', () => this.toggleEdit(false));
    }

    const changeAvatarBtn = this.$('[data-action="change-avatar"]');
    if (changeAvatarBtn) {
      changeAvatarBtn.addEventListener('click', () => this.handleChangeAvatar());
    }
  }

  toggleEdit(editing) {
    this.setState({ editing });
  }

  async handleSave() {
    try {
      loading.show('Guardando cambios...');

      const firstName = this.$('#firstNameInput')?.value;
      const lastName = this.$('#lastNameInput')?.value;
      const bio = this.$('#bioInput')?.value;

      const updates = { firstName, lastName, bio };
      await userService.updateProfile(updates);

      toast.success('Perfil actualizado correctamente');
      this.setState({ editing: false });
      await this.loadProfile();
    } catch (error) {
      toast.error(error.message || 'Error al actualizar el perfil');
    } finally {
      loading.hide();
    }
  }

  handleChangeAvatar() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'image/*';

    input.onchange = async (e) => {
      const file = e.target.files[0];
      if (file) {
        try {
          loading.show('Subiendo avatar...');
          await userService.updateAvatar(file, (progress) => {
            loading.setText(`Subiendo avatar... ${Math.round(progress)}%`);
          });
          toast.success('Avatar actualizado correctamente');
          await this.loadProfile();
        } catch (error) {
          toast.error('Error al actualizar el avatar');
        } finally {
          loading.hide();
        }
      }
    };

    input.click();
  }
}
