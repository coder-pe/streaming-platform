// app.js - Aplicación Principal
class StreamLearnApp {
    constructor() {
        this.apiBase = '/api';
        this.currentUser = null;
        this.currentVideo = null;
        this.videoPlayer = null;
        this.uploadProgress = 0;
        
        this.init();
    }

    async init() {
        this.setupEventListeners();
        this.setupServiceWorker();
        await this.checkAuthStatus();
        this.loadFeaturedVideos();
        this.showSection('home');
    }

    // ===== SETUP METHODS =====
    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const section = link.getAttribute('href').substring(1);
                this.showSection(section);
            });
        });

        // Auth buttons
        document.getElementById('loginBtn')?.addEventListener('click', () => this.showModal('loginModal'));
        document.getElementById('registerBtn')?.addEventListener('click', () => this.showModal('registerModal'));
        document.getElementById('logoutBtn')?.addEventListener('click', () => this.logout());

        // Modal controls
        document.querySelectorAll('.modal-close').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.hideModal(e.target.closest('.modal').id);
            });
        });

        // Auth form switches
        document.getElementById('switchToRegister')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.hideModal('loginModal');
            this.showModal('registerModal');
        });

        document.getElementById('switchToLogin')?.addEventListener('click', (e) => {
            e.preventDefault();
            this.hideModal('registerModal');
            this.showModal('loginModal');
        });

        // Forms
        document.getElementById('loginForm')?.addEventListener('submit', (e) => this.handleLogin(e));
        document.getElementById('registerForm')?.addEventListener('submit', (e) => this.handleRegister(e));
        document.getElementById('uploadForm')?.addEventListener('submit', (e) => this.handleVideoUpload(e));

        // Search and filters
        document.getElementById('searchBtn')?.addEventListener('click', () => this.searchVideos());
        document.getElementById('searchInput')?.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.searchVideos();
        });

        document.getElementById('categoryFilter')?.addEventListener('change', () => this.searchVideos());
        document.getElementById('sortFilter')?.addEventListener('change', () => this.searchVideos());

        // Video upload
        document.getElementById('selectFileBtn')?.addEventListener('click', () => {
            document.getElementById('videoFile').click();
        });

        document.getElementById('videoFile')?.addEventListener('change', (e) => this.handleFileSelect(e));

        // Drag and drop
        const uploadArea = document.getElementById('uploadArea');
        if (uploadArea) {
            uploadArea.addEventListener('dragover', (e) => {
                e.preventDefault();
                uploadArea.classList.add('dragover');
            });

            uploadArea.addEventListener('dragleave', () => {
                uploadArea.classList.remove('dragover');
            });

            uploadArea.addEventListener('drop', (e) => {
                e.preventDefault();
                uploadArea.classList.remove('dragover');
                const files = e.dataTransfer.files;
                if (files.length > 0) {
                    this.handleFileSelect({ target: { files } });
                }
            });
        }

        // Category cards
        document.querySelectorAll('.category-card').forEach(card => {
            card.addEventListener('click', () => {
                const category = card.dataset.category;
                this.showSection('videos');
                document.getElementById('categoryFilter').value = category;
                this.searchVideos();
            });
        });

        // Profile tabs
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const tabName = btn.dataset.tab;
                this.showTab(tabName);
            });
        });

        // Close modals on outside click
        document.querySelectorAll('.modal').forEach(modal => {
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.hideModal(modal.id);
                }
            });
        });

        // Explore button
        document.getElementById('exploreBtn')?.addEventListener('click', () => {
            this.showSection('videos');
        });
    }

    setupServiceWorker() {
        if ('serviceWorker' in navigator) {
            navigator.serviceWorker.register('/static/sw.js')
                .then(registration => {
                    console.log('SW registered:', registration);
                })
                .catch(error => {
                    console.log('SW registration failed:', error);
                });
        }
    }

    // ===== NAVIGATION =====
    showSection(sectionName) {
        // Update navigation
        document.querySelectorAll('.nav-link').forEach(link => {
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${sectionName}`) {
                link.classList.add('active');
            }
        });

        // Show section
        document.querySelectorAll('.section').forEach(section => {
            section.classList.remove('active');
        });
        
        const targetSection = document.getElementById(sectionName);
        if (targetSection) {
            targetSection.classList.add('active');
            
            // Load content based on section
            switch (sectionName) {
                case 'videos':
                    this.loadVideos();
                    break;
                case 'profile':
                    if (this.currentUser) {
                        this.loadProfile();
                    } else {
                        this.showModal('loginModal');
                    }
                    break;
                case 'upload':
                    if (this.currentUser) {
                        this.resetUploadForm();
                    } else {
                        this.showModal('loginModal');
                    }
                    break;
            }
        }
    }

    showTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.tab === tabName) {
                btn.classList.add('active');
            }
        });

        // Show tab content
        document.querySelectorAll('.tab-panel').forEach(panel => {
            panel.classList.remove('active');
        });
        
        const targetPanel = document.getElementById(tabName);
        if (targetPanel) {
            targetPanel.classList.add('active');
            
            // Load tab-specific content
            switch (tabName) {
                case 'my-videos':
                    this.loadUserVideos();
                    break;
                case 'favorites':
                    this.loadFavorites();
                    break;
                case 'history':
                    this.loadWatchHistory();
                    break;
            }
        }
    }

    // ===== MODAL MANAGEMENT =====
    showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('active');
            document.body.style.overflow = 'hidden';
        }
    }

    hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('active');
            document.body.style.overflow = '';
            
            // Reset forms
            const form = modal.querySelector('form');
            if (form) form.reset();
        }
    }

    // ===== AUTHENTICATION =====
    async checkAuthStatus() {
        const token = localStorage.getItem('authToken');
        if (token) {
            try {
                const response = await this.apiCall('/user/profile', 'GET');
                if (response.ok) {
                    this.currentUser = await response.json();
                    this.updateUIForAuthenticatedUser();
                } else {
                    localStorage.removeItem('authToken');
                }
            } catch (error) {
                console.error('Auth check failed:', error);
                localStorage.removeItem('authToken');
            }
        }
    }

    async handleLogin(e) {
        e.preventDefault();
        
        const email = document.getElementById('loginEmail').value;
        const password = document.getElementById('loginPassword').value;
        
        this.showLoading(true);
        
        try {
            const response = await this.apiCall('/auth/login', 'POST', {
                email,
                password
            });
            
            if (response.ok) {
                const data = await response.json();
                localStorage.setItem('authToken', data.token);
                this.currentUser = data.user;
                this.updateUIForAuthenticatedUser();
                this.hideModal('loginModal');
                this.showToast('Inicio de sesión exitoso', 'success');
            } else {
                const error = await response.json();
                this.showToast(error.message || 'Error en el inicio de sesión', 'error');
            }
        } catch (error) {
            console.error('Login error:', error);
            this.showToast('Error de conexión', 'error');
        } finally {
            this.showLoading(false);
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        
        const firstName = document.getElementById('registerFirstName').value;
        const lastName = document.getElementById('registerLastName').value;
        const email = document.getElementById('registerEmail').value;
        const password = document.getElementById('registerPassword').value;
        const confirmPassword = document.getElementById('confirmPassword').value;
        
        if (password !== confirmPassword) {
            this.showToast('Las contraseñas no coinciden', 'error');
            return;
        }
        
        this.showLoading(true);
        
        try {
            const response = await this.apiCall('/auth/register', 'POST', {
                first_name: firstName,
                last_name: lastName,
                email,
                password
            });
            
            if (response.ok) {
                const data = await response.json();
                localStorage.setItem('authToken', data.token);
                this.currentUser = data.user;
                this.updateUIForAuthenticatedUser();
                this.hideModal('registerModal');
                this.showToast('Registro exitoso', 'success');
            } else {
                const error = await response.json();
                this.showToast(error.message || 'Error en el registro', 'error');
            }
        } catch (error) {
            console.error('Register error:', error);
            this.showToast('Error de conexión', 'error');
        } finally {
            this.showLoading(false);
        }
    }

    logout() {
        localStorage.removeItem('authToken');
        this.currentUser = null;
        this.updateUIForUnauthenticatedUser();
        this.showSection('home');
        this.showToast('Sesión cerrada', 'success');
    }

    updateUIForAuthenticatedUser() {
        // Hide auth buttons, show user menu
        document.getElementById('loginBtn').style.display = 'none';
        document.getElementById('registerBtn').style.display = 'none';
        document.getElementById('userMenu').classList.remove('hidden');
        
        // Show auth-required elements
        document.querySelectorAll('.auth-required').forEach(el => {
            el.style.display = 'block';
        });
        
        // Update user info
        if (this.currentUser) {
            document.getElementById('avatarImg').src = this.currentUser.avatar || '/static/assets/default-avatar.png';
            document.getElementById('profileName').textContent = `${this.currentUser.first_name} ${this.currentUser.last_name}`;
            document.getElementById('profileEmail').textContent = this.currentUser.email;
            document.getElementById('profileRole').textContent = this.currentUser.role;
        }
    }

    updateUIForUnauthenticatedUser() {
        // Show auth buttons, hide user menu
        document.getElementById('loginBtn').style.display = 'inline-flex';
        document.getElementById('registerBtn').style.display = 'inline-flex';
        document.getElementById('userMenu').classList.add('hidden');
        
        // Hide auth-required elements
        document.querySelectorAll('.auth-required').forEach(el => {
            el.style.display = 'none';
        });
    }

    // ===== VIDEO MANAGEMENT =====
    async loadFeaturedVideos() {
        try {
            const response = await this.apiCall('/videos?limit=6', 'GET');
            if (response.ok) {
                const data = await response.json();
                this.renderVideoGrid(data.videos, 'featuredVideos');
            }
        } catch (error) {
            console.error('Error loading featured videos:', error);
        }
    }

    async loadVideos(page = 1) {
        this.showLoading(true, 'videosLoading');
        
        try {
            const searchInput = document.getElementById('searchInput')?.value || '';
            const category = document.getElementById('categoryFilter')?.value || '';
            const sort = document.getElementById('sortFilter')?.value || 'newest';
            
            const params = new URLSearchParams({
                page: page.toString(),
                limit: '12'
            });
            
            if (searchInput) params.append('query', searchInput);
            if (category) params.append('category', category);
            if (sort) params.append('sort', sort);
            
            const response = await this.apiCall(`/videos?${params}`, 'GET');
            if (response.ok) {
                const data = await response.json();
                this.renderVideoGrid(data.videos, 'videosGrid');
                this.renderPagination(data, 'videosPagination');
            }
        } catch (error) {
            console.error('Error loading videos:', error);
            this.showToast('Error cargando videos', 'error');
        } finally {
            this.showLoading(false, 'videosLoading');
        }
    }

    async searchVideos() {
        await this.loadVideos(1);
    }

    renderVideoGrid(videos, containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;
        
        if (videos.length === 0) {
            container.innerHTML = '<p class="text-center">No se encontraron videos.</p>';
            return;
        }
        
        container.innerHTML = videos.map(video => `
            <div class="video-card" onclick="app.playVideo('${video.id}')">
                <div class="video-thumbnail">
                    <img src="${video.thumbnail || '/static/assets/video-placeholder.jpg'}" 
                         alt="${video.title}" 
                         onerror="this.src='/static/assets/video-placeholder.jpg'">
                    <div class="video-duration">${this.formatDuration(video.duration)}</div>
                </div>
                <div class="video-info">
                    <h4 class="video-title">${video.title}</h4>
                    <p class="video-instructor">${video.instructor ? video.instructor.first_name + ' ' + video.instructor.last_name : 'Instructor'}</p>
                    <div class="video-meta">
                        <span>${video.view_count || 0} vistas</span>
                        <div class="video-rating">
                            <span>⭐ ${video.rating ? video.rating.toFixed(1) : 'N/A'}</span>
                        </div>
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderPagination(data, containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;
        
        const totalPages = Math.ceil(data.total / data.limit);
        const currentPage = data.page;
        
        if (totalPages <= 1) {
            container.innerHTML = '';
            return;
        }
        
        let paginationHTML = '';
        
        // Previous button
        paginationHTML += `
            <button ${currentPage <= 1 ? 'disabled' : ''} 
                    onclick="app.loadVideos(${currentPage - 1})">
                Anterior
            </button>
        `;
        
        // Page numbers
        for (let i = Math.max(1, currentPage - 2); i <= Math.min(totalPages, currentPage + 2); i++) {
            paginationHTML += `
                <button class="${i === currentPage ? 'active' : ''}" 
                        onclick="app.loadVideos(${i})">
                    ${i}
                </button>
            `;
        }
        
        // Next button
        paginationHTML += `
            <button ${currentPage >= totalPages ? 'disabled' : ''} 
                    onclick="app.loadVideos(${currentPage + 1})">
                Siguiente
            </button>
        `;
        
        container.innerHTML = paginationHTML;
    }

    async playVideo(videoId) {
        this.showLoading(true);
        
        try {
            const response = await this.apiCall(`/videos/${videoId}`, 'GET');
            if (response.ok) {
                const video = await response.json();
                this.currentVideo = video;
                this.showVideoModal(video);
            } else {
                this.showToast('Video no encontrado', 'error');
            }
        } catch (error) {
            console.error('Error loading video:', error);
            this.showToast('Error cargando video', 'error');
        } finally {
            this.showLoading(false);
        }
    }

    showVideoModal(video) {
        // Update modal content
        document.getElementById('modalVideoTitle').textContent = video.title;
        document.getElementById('modalVideoDescription').textContent = video.description;
        document.getElementById('instructorName').textContent = 
            video.instructor ? `${video.instructor.first_name} ${video.instructor.last_name}` : 'Instructor';
        document.getElementById('videoDate').textContent = 
            `Publicado ${this.formatDate(video.created_at)}`;
        
        // Render tags
        const tagsContainer = document.getElementById('modalVideoTags');
        if (video.tags && video.tags.length > 0) {
            tagsContainer.innerHTML = video.tags.map(tag => 
                `<span class="tag">${tag}</span>`
            ).join('');
        }
        
        // Initialize video player
        this.initVideoPlayer(video);
        
        // Show modal
        this.showModal('videoModal');
    }

    initVideoPlayer(video) {
        // Destroy existing player if any
        if (this.videoPlayer) {
            this.videoPlayer.dispose();
            this.videoPlayer = null;
        }
        
        // Create new player
        this.videoPlayer = videojs('videoPlayer', {
            fluid: true,
            responsive: true,
            playbackRates: [0.5, 1, 1.25, 1.5, 2],
            controls: true,
            preload: 'auto'
        });
        
        // Set HLS source
        const streamUrl = `/api/stream/${video.id}/playlist`;
        this.videoPlayer.src({
            src: streamUrl,
            type: 'application/x-mpegURL'
        });
        
        // Track watch progress
        this.videoPlayer.on('timeupdate', () => {
            if (this.currentUser) {
                const currentTime = Math.floor(this.videoPlayer.currentTime());
                this.updateWatchProgress(video.id, currentTime);
            }
        });
        
        // Handle player errors
        this.videoPlayer.on('error', () => {
            console.error('Video player error');
            this.showToast('Error reproduciendo video', 'error');
        });
    }

    async updateWatchProgress(videoId, position) {
        if (!this.currentUser) return;
        
        try {
            await this.apiCall(`/videos/${videoId}/progress`, 'PUT', {
                position,
                quality: '720p' // Default quality
            });
        } catch (error) {
            console.error('Error updating watch progress:', error);
        }
    }

    // ===== FILE UPLOAD =====
    handleFileSelect(e) {
        const file = e.target.files[0];
        if (!file) return;
        
        // Validate file type
        if (!file.type.startsWith('video/')) {
            this.showToast('Por favor selecciona un archivo de video válido', 'error');
            return;
        }
        
        // Validate file size (100MB max)
        const maxSize = 100 * 1024 * 1024;
        if (file.size > maxSize) {
            this.showToast('El archivo es demasiado grande. Máximo 100MB.', 'error');
            return;
        }
        
        // Show file info
        const fileInfo = document.getElementById('fileInfo');
        fileInfo.innerHTML = `
            <h4>📹 ${file.name}</h4>
            <p>Tamaño: ${this.formatFileSize(file.size)}</p>
            <p>Tipo: ${file.type}</p>
        `;
        
        // Show upload details
        document.getElementById('uploadDetails').classList.remove('hidden');
        
        // Store file for upload
        this.selectedFile = file;
    }

    async handleVideoUpload(e) {
        e.preventDefault();
        
        if (!this.selectedFile) {
            this.showToast('Por favor selecciona un archivo de video', 'error');
            return;
        }
        
        const title = document.getElementById('videoTitle').value;
        const description = document.getElementById('videoDescription').value;
        const category = document.getElementById('videoCategory').value;
        const tags = document.getElementById('videoTags').value.split(',').map(tag => tag.trim()).filter(tag => tag);
        const isPublic = document.getElementById('videoPublic').checked;
        
        if (!title || !category) {
            this.showToast('Por favor completa los campos requeridos', 'error');
            return;
        }
        
        const formData = new FormData();
        formData.append('video', this.selectedFile);
        formData.append('metadata', JSON.stringify({
            title,
            description,
            category,
            tags,
            is_public: isPublic
        }));
        
        // Show progress
        document.getElementById('uploadProgress').classList.remove('hidden');
        document.getElementById('uploadSubmit').disabled = true;
        
        try {
            const xhr = new XMLHttpRequest();
            
            // Track upload progress
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentComplete = Math.round((e.loaded / e.total) * 100);
                    this.updateUploadProgress(percentComplete);
                }
            });
            
            // Handle completion
            xhr.addEventListener('load', () => {
                if (xhr.status === 201) {
                    this.showToast('Video subido exitosamente. Procesando...', 'success');
                    this.resetUploadForm();
                    this.showSection('profile');
                    this.showTab('my-videos');
                } else {
                    const error = JSON.parse(xhr.responseText);
                    this.showToast(error.message || 'Error subiendo video', 'error');
                }
                document.getElementById('uploadSubmit').disabled = false;
            });
            
            // Handle errors
            xhr.addEventListener('error', () => {
                this.showToast('Error de conexión durante la subida', 'error');
                document.getElementById('uploadSubmit').disabled = false;
            });
            
            // Send request
            xhr.open('POST', `${this.apiBase}/videos/upload`);
            xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('authToken')}`);
            xhr.send(formData);
            
        } catch (error) {
            console.error('Upload error:', error);
            this.showToast('Error subiendo video', 'error');
            document.getElementById('uploadSubmit').disabled = false;
        }
    }

    updateUploadProgress(percent) {
        document.getElementById('progressFill').style.width = `${percent}%`;
        document.getElementById('progressText').textContent = `Subiendo: ${percent}%`;
    }

    resetUploadForm() {
        document.getElementById('uploadForm').reset();
        document.getElementById('uploadDetails').classList.add('hidden');
        document.getElementById('uploadProgress').classList.add('hidden');
        this.selectedFile = null;
        this.updateUploadProgress(0);
    }

    // ===== PROFILE MANAGEMENT =====
    async loadProfile() {
        if (!this.currentUser) return;
        
        // Update profile display
        document.getElementById('profileAvatar').src = this.currentUser.avatar || '/static/assets/default-avatar.png';
        document.getElementById('profileName').textContent = `${this.currentUser.first_name} ${this.currentUser.last_name}`;
        document.getElementById('profileEmail').textContent = this.currentUser.email;
        document.getElementById('profileRole').textContent = this.currentUser.role;
        
        // Load default tab
        this.showTab('my-videos');
    }

    async loadUserVideos() {
        if (!this.currentUser) return;
        
        try {
            const response = await this.apiCall(`/videos?instructor=${this.currentUser.id}`, 'GET');
            if (response.ok) {
                const data = await response.json();
                this.renderVideoGrid(data.videos, 'myVideosGrid');
            }
        } catch (error) {
            console.error('Error loading user videos:', error);
        }
    }

    async loadFavorites() {
        // Implementation for loading favorite videos
        console.log('Loading favorites...');
    }

    async loadWatchHistory() {
        // Implementation for loading watch history
        console.log('Loading watch history...');
    }

    // ===== UTILITY METHODS =====
    async apiCall(endpoint, method = 'GET', data = null) {
        const url = `${this.apiBase}${endpoint}`;
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json'
            }
        };
        
        // Add auth token if available
        const token = localStorage.getItem('authToken');
        if (token) {
            options.headers.Authorization = `Bearer ${token}`;
        }
        
        // Add body for POST/PUT requests
        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }
        
        return fetch(url, options);
    }

    showLoading(show, elementId = 'loadingOverlay') {
        const element = document.getElementById(elementId);
        if (element) {
            if (show) {
                element.classList.remove('hidden');
            } else {
                element.classList.add('hidden');
            }
        }
    }

    showToast(message, type = 'info') {
        const container = document.getElementById('toastContainer');
        if (!container) return;
        
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <div class="toast-content">
                <p>${message}</p>
            </div>
        `;
        
        container.appendChild(toast);
        
        // Auto remove after 5 seconds
        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        }, 5000);
        
        // Allow manual close
        toast.addEventListener('click', () => {
            toast.style.opacity = '0';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        });
    }

    formatDuration(seconds) {
        if (!seconds) return '0:00';
        
        const hours = Math.floor(seconds / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        const secs = seconds % 60;
        
        if (hours > 0) {
            return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
        } else {
            return `${minutes}:${secs.toString().padStart(2, '0')}`;
        }
    }

    formatFileSize(bytes) {
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        if (bytes === 0) return '0 Bytes';
        
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffTime = Math.abs(now - date);
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        
        if (diffDays === 1) return 'hace 1 día';
        if (diffDays < 7) return `hace ${diffDays} días`;
        if (diffDays < 30) return `hace ${Math.floor(diffDays / 7)} semanas`;
        if (diffDays < 365) return `hace ${Math.floor(diffDays / 30)} meses`;
        return `hace ${Math.floor(diffDays / 365)} años`;
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.app = new StreamLearnApp();
});

// Export for module usage
export default StreamLearnApp;

