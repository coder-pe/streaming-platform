/**
 * @file app.js
 * @description Punto de entrada principal de la aplicación modular
 */

// Core modules
import router from './core/Router.js';
import state from './core/State.js';
import eventBus from './core/EventBus.js';

// Services
import authService from './services/AuthService.js';
import videoService from './services/VideoService.js';

// Components
import toast from './components/base/Toast.js';
import loading from './components/base/Loading.js';
import Modal from './components/base/Modal.js';
import LoginForm from './components/auth/LoginForm.js';
import RegisterForm from './components/auth/RegisterForm.js';
import VideoGrid from './components/video/VideoGrid.js';
import VideoPlayer from './components/video/VideoPlayer.js';
import VideoUpload from './components/video/VideoUpload.js';
import UserProfile from './components/user/UserProfile.js';
import UserMenu from './components/user/UserMenu.js';

/**
 * Clase principal de la aplicación
 */
class App {
  constructor() {
    this.components = {};
    this.modals = {};
  }

  /**
   * Inicializar aplicación
   */
  async init() {
    try {
      // Mostrar loading inicial
      loading.show('Iniciando aplicación...');

      // Inicializar autenticación
      await authService.initialize();

      // Inicializar componentes persistentes
      this.initPersistentComponents();

      // Configurar rutas
      this.setupRoutes();

      // Configurar event listeners globales
      this.setupEventListeners();

      // Navegar a ruta inicial
      router.handleRouteChange();

      // Ocultar loading
      loading.hide();

      console.log('✅ Aplicación inicializada correctamente');
    } catch (error) {
      console.error('❌ Error inicializando aplicación:', error);
      loading.hide();
      toast.error('Error al iniciar la aplicación');
    }
  }

  /**
   * Inicializar componentes persistentes (siempre visibles)
   */
  initPersistentComponents() {
    // Menú de usuario
    const userMenuContainer = document.querySelector('#userMenu');
    if (userMenuContainer) {
      this.components.userMenu = new UserMenu(userMenuContainer);
      this.components.userMenu.mount();
    }

    // Reproductor de video (oculto inicialmente)
    const playerContainer = document.querySelector('#videoPlayerContainer');
    if (playerContainer) {
      this.components.videoPlayer = new VideoPlayer(playerContainer);
      this.components.videoPlayer.mount();
      this.components.videoPlayer.hide();
    }
  }

  /**
   * Configurar rutas de la aplicación
   */
  setupRoutes() {
    // Ruta: Home (lista de videos)
    router.register('home', () => this.showHome());

    // Ruta: Videos (explorar)
    router.register('videos', () => this.showVideos());

    // Ruta: Upload
    router.register('upload', () => this.showUpload());

    // Ruta: Profile
    router.register('profile', () => this.showProfile());

    // Establecer ruta por defecto
    router.setDefault('home');
  }

  /**
   * Configurar event listeners globales
   */
  setupEventListeners() {
    // Navegación
    document.addEventListener('click', (e) => {
      const navLink = e.target.closest('[data-route]');
      if (navLink) {
        e.preventDefault();
        const route = navLink.getAttribute('data-route');
        router.navigate(route);
      }
    });

    // Eventos del menú de usuario
    eventBus.on('menu:login', () => this.showLoginModal());
    eventBus.on('menu:register', () => this.showRegisterModal());

    // Eventos de videos
    eventBus.on('video:select', (video) => this.handleVideoSelect(video));
    eventBus.on('video:play', (video) => this.handleVideoPlay(video));

    // Eventos de autenticación
    eventBus.on('auth:login', () => {
      toast.success('¡Bienvenido!');
      this.closeAllModals();
      router.navigate('home');
    });

    eventBus.on('auth:logout', () => {
      toast.info('Sesión cerrada');
      router.navigate('home');
    });

    eventBus.on('auth:unauthorized', () => {
      toast.warning('Tu sesión ha expirado. Por favor inicia sesión nuevamente.');
      this.showLoginModal();
    });

    // Eventos de cambio de ruta
    eventBus.on('route:before', (route) => {
      this.hideAllSections();
      loading.show();
    });

    eventBus.on('route:after', (route) => {
      loading.hide();
      this.updateNavigation(route);
    });

    // Búsqueda
    const searchInput = document.querySelector('#searchInput');
    if (searchInput) {
      let searchTimeout;
      searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
          this.handleSearch(e.target.value);
        }, 300);
      });
    }
  }

  /**
   * Mostrar sección Home
   */
  async showHome() {
    this.showSection('homeSection');

    try {
      const videosContainer = document.querySelector('#homeVideos');
      if (!videosContainer) return;

      // Crear o reutilizar grid de videos
      if (!this.components.homeVideoGrid) {
        this.components.homeVideoGrid = new VideoGrid(videosContainer);
        this.components.homeVideoGrid.mount();
      }

      this.components.homeVideoGrid.setLoading(true);

      // Cargar videos
      const response = await videoService.getVideos({ limit: 12 });
      this.components.homeVideoGrid.setVideos(response.videos || []);
    } catch (error) {
      console.error('Error loading home videos:', error);
      toast.error('Error al cargar los videos');
      if (this.components.homeVideoGrid) {
        this.components.homeVideoGrid.setError(error.message);
      }
    }
  }

  /**
   * Mostrar sección de Videos
   */
  async showVideos() {
    this.showSection('videosSection');

    try {
      const videosContainer = document.querySelector('#allVideos');
      if (!videosContainer) return;

      if (!this.components.videosGrid) {
        this.components.videosGrid = new VideoGrid(videosContainer);
        this.components.videosGrid.mount();
      }

      this.components.videosGrid.setLoading(true);

      const response = await videoService.getVideos({ limit: 24 });
      this.components.videosGrid.setVideos(response.videos || []);
    } catch (error) {
      console.error('Error loading videos:', error);
      toast.error('Error al cargar los videos');
      if (this.components.videosGrid) {
        this.components.videosGrid.setError(error.message);
      }
    }
  }

  /**
   * Mostrar sección de Upload
   */
  showUpload() {
    // Verificar autenticación
    if (!authService.isAuthenticated()) {
      toast.warning('Debes iniciar sesión para subir videos');
      this.showLoginModal();
      return;
    }

    this.showSection('uploadSection');

    const uploadContainer = document.querySelector('#uploadContainer');
    if (!uploadContainer) return;

    if (!this.components.videoUpload) {
      this.components.videoUpload = new VideoUpload(uploadContainer);
      this.components.videoUpload.mount();

      // Escuchar evento de upload exitoso
      uploadContainer.addEventListener('upload:success', () => {
        router.navigate('videos');
      });
    }
  }

  /**
   * Mostrar sección de Profile
   */
  showProfile() {
    // Verificar autenticación
    if (!authService.isAuthenticated()) {
      toast.warning('Debes iniciar sesión para ver tu perfil');
      this.showLoginModal();
      return;
    }

    this.showSection('profileSection');

    const profileContainer = document.querySelector('#profileContainer');
    if (!profileContainer) return;

    if (!this.components.userProfile) {
      this.components.userProfile = new UserProfile(profileContainer);
      this.components.userProfile.mount();
    } else {
      // Re-cargar perfil si ya existe
      this.components.userProfile.loadProfile();
    }
  }

  /**
   * Manejar selección de video
   */
  handleVideoSelect(video) {
    console.log('Video selected:', video);
    // Aquí se podría abrir un modal con detalles del video
    // o navegar a una página de detalles
  }

  /**
   * Manejar reproducción de video
   */
  handleVideoPlay(video) {
    if (this.components.videoPlayer) {
      this.components.videoPlayer.loadVideo(video);
      this.components.videoPlayer.show();

      // Scroll al reproductor
      this.components.videoPlayer.container.scrollIntoView({
        behavior: 'smooth',
        block: 'start'
      });
    }
  }

  /**
   * Manejar búsqueda
   */
  async handleSearch(query) {
    if (!query || query.length < 3) return;

    try {
      loading.show('Buscando...');

      const response = await videoService.searchVideos(query);

      // Actualizar grid actual con resultados
      const currentGrid = this.components.homeVideoGrid || this.components.videosGrid;
      if (currentGrid) {
        currentGrid.setVideos(response.videos || []);
      }

      toast.info(`Se encontraron ${response.videos?.length || 0} resultados`);
    } catch (error) {
      console.error('Search error:', error);
      toast.error('Error al buscar videos');
    } finally {
      loading.hide();
    }
  }

  /**
   * Mostrar modal de login
   */
  showLoginModal() {
    if (this.modals.login) {
      this.modals.login.open();
      return;
    }

    const modalContainer = document.createElement('div');
    const loginFormContainer = document.createElement('div');

    this.modals.login = new Modal({
      title: 'Iniciar Sesión',
      content: loginFormContainer.outerHTML,
      size: 'small',
      onClose: () => {
        if (this.components.loginForm) {
          this.components.loginForm.unmount();
        }
      }
    });

    this.modals.login.open();

    // Montar formulario de login en el modal
    const modalBody = this.modals.login.$('.modal-body');
    if (modalBody) {
      this.components.loginForm = new LoginForm(modalBody);
      this.components.loginForm.mount();

      // Escuchar evento de login exitoso
      modalBody.addEventListener('login:success', () => {
        this.modals.login.close();
      });
    }
  }

  /**
   * Mostrar modal de registro
   */
  showRegisterModal() {
    if (this.modals.register) {
      this.modals.register.open();
      return;
    }

    const modalContainer = document.createElement('div');
    const registerFormContainer = document.createElement('div');

    this.modals.register = new Modal({
      title: 'Crear Cuenta',
      content: registerFormContainer.outerHTML,
      size: 'small',
      onClose: () => {
        if (this.components.registerForm) {
          this.components.registerForm.unmount();
        }
      }
    });

    this.modals.register.open();

    // Montar formulario de registro en el modal
    const modalBody = this.modals.register.$('.modal-body');
    if (modalBody) {
      this.components.registerForm = new RegisterForm(modalBody);
      this.components.registerForm.mount();

      // Escuchar evento de registro exitoso
      modalBody.addEventListener('register:success', () => {
        this.modals.register.close();
      });
    }
  }

  /**
   * Cerrar todos los modales
   */
  closeAllModals() {
    Object.values(this.modals).forEach(modal => {
      if (modal.isOpen) {
        modal.close();
      }
    });
  }

  /**
   * Mostrar sección específica
   */
  showSection(sectionId) {
    const section = document.querySelector(`#${sectionId}`);
    if (section) {
      section.style.display = 'block';
      section.classList.remove('hidden');
    }
  }

  /**
   * Ocultar todas las secciones
   */
  hideAllSections() {
    const sections = document.querySelectorAll('main > section');
    sections.forEach(section => {
      section.style.display = 'none';
      section.classList.add('hidden');
    });

    // Ocultar reproductor también
    if (this.components.videoPlayer) {
      this.components.videoPlayer.hide();
    }
  }

  /**
   * Actualizar navegación activa
   */
  updateNavigation(currentRoute) {
    const navLinks = document.querySelectorAll('[data-route]');
    navLinks.forEach(link => {
      const route = link.getAttribute('data-route');
      if (route === currentRoute) {
        link.classList.add('active');
      } else {
        link.classList.remove('active');
      }
    });
  }
}

// Inicializar aplicación cuando el DOM esté listo
document.addEventListener('DOMContentLoaded', () => {
  const app = new App();
  app.init();

  // Exponer app globalmente para debugging
  window.app = app;
});

// Exportar para uso en módulos
export default App;
