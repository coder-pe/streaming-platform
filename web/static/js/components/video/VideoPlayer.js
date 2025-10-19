/**
 * @file VideoPlayer.js
 * @description Componente de reproductor de video
 */

import Component from '../base/Component.js';
import videoService from '../../services/VideoService.js';
import { VIDEO_CONFIG } from '../../config/constants.js';

export default class VideoPlayer extends Component {
  constructor(container) {
    super(container);
    this.state = {
      video: null,
      quality: VIDEO_CONFIG.DEFAULT_QUALITY,
      playing: false,
    };
    this.player = null;
  }

  template() {
    if (!this.state.video) {
      return '<div class="video-player-empty">Selecciona un video para reproducir</div>';
    }

    return `
      <div class="video-player-container">
        <video
          id="videoPlayer"
          class="video-js vjs-default-skin"
          controls
          preload="auto"
          data-setup='{"fluid": true}'
        >
          <source src="${this.getVideoSource()}" type="application/x-mpegURL">
        </video>

        <div class="video-player-info">
          <h2>${this.state.video.title}</h2>
          <div class="video-player-meta">
            <span>${this.state.video.views || 0} vistas</span>
            <div class="video-player-actions">
              <button class="btn btn-icon" data-action="like">
                <i class="icon-heart"></i>
                <span>${this.state.video.likes || 0}</span>
              </button>
              <button class="btn btn-icon" data-action="share">
                <i class="icon-share"></i>
                Compartir
              </button>
            </div>
          </div>
          ${this.state.video.description ? `
            <div class="video-player-description">
              <p>${this.state.video.description}</p>
            </div>
          ` : ''}
        </div>

        <div class="video-quality-selector">
          <label for="qualitySelect">Calidad:</label>
          <select id="qualitySelect" class="form-control">
            ${VIDEO_CONFIG.PLAYBACK_RATES.map(rate => `
              <option value="${rate}p" ${this.state.quality === `${rate}p` ? 'selected' : ''}>
                ${rate}p
              </option>
            `).join('')}
          </select>
        </div>
      </div>
    `;
  }

  attachEventListeners() {
    // Botón de like
    const likeBtn = this.$('[data-action="like"]');
    if (likeBtn) {
      likeBtn.addEventListener('click', () => this.handleLike());
    }

    // Botón de compartir
    const shareBtn = this.$('[data-action="share"]');
    if (shareBtn) {
      shareBtn.addEventListener('click', () => this.handleShare());
    }

    // Selector de calidad
    const qualitySelect = this.$('#qualitySelect');
    if (qualitySelect) {
      qualitySelect.addEventListener('change', (e) => {
        this.changeQuality(e.target.value);
      });
    }
  }

  onMount() {
    if (this.state.video) {
      this.initPlayer();
    }
  }

  initPlayer() {
    // Inicializar Video.js si está disponible
    if (typeof videojs !== 'undefined') {
      const playerEl = this.$('#videoPlayer');
      if (playerEl) {
        this.player = videojs(playerEl, {
          controls: true,
          autoplay: false,
          preload: 'auto',
          fluid: true,
          playbackRates: VIDEO_CONFIG.PLAYBACK_RATES,
        });

        // Event listeners del player
        this.player.on('play', () => this.onPlay());
        this.player.on('pause', () => this.onPause());
        this.player.on('ended', () => this.onEnded());
      }
    }
  }

  loadVideo(video) {
    this.setState({ video });

    if (this.player) {
      this.player.src({
        src: this.getVideoSource(),
        type: 'application/x-mpegURL',
      });
      this.player.load();
    }

    // Incrementar vistas
    videoService.incrementViews(video.id);
  }

  getVideoSource() {
    if (!this.state.video) return '';
    return videoService.getStreamUrl(this.state.video.id, this.state.quality);
  }

  changeQuality(quality) {
    const currentTime = this.player ? this.player.currentTime() : 0;
    const wasPlaying = this.state.playing;

    this.setState({ quality });

    if (this.player) {
      this.player.src({
        src: this.getVideoSource(),
        type: 'application/x-mpegURL',
      });
      this.player.load();
      this.player.currentTime(currentTime);

      if (wasPlaying) {
        this.player.play();
      }
    }
  }

  async handleLike() {
    try {
      if (this.state.video.liked) {
        await videoService.unlikeVideo(this.state.video.id);
      } else {
        await videoService.likeVideo(this.state.video.id);
      }
    } catch (error) {
      console.error('Error toggling like:', error);
    }
  }

  handleShare() {
    const url = window.location.href;
    if (navigator.share) {
      navigator.share({
        title: this.state.video.title,
        text: this.state.video.description,
        url: url,
      });
    } else {
      // Fallback: copiar al portapapeles
      navigator.clipboard.writeText(url);
      this.emit('video:share', { url });
    }
  }

  onPlay() {
    this.setState({ playing: true });
    this.emit('video:play', this.state.video);
  }

  onPause() {
    this.setState({ playing: false });
    this.emit('video:pause', this.state.video);
  }

  onEnded() {
    this.setState({ playing: false });
    this.emit('video:ended', this.state.video);
  }

  onUnmount() {
    if (this.player) {
      this.player.dispose();
      this.player = null;
    }
  }

  emit(event, data) {
    this.container.dispatchEvent(new CustomEvent(event, {
      detail: data,
      bubbles: true
    }));
  }
}
