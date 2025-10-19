/**
 * @file VideoGrid.js
 * @description Grilla de videos
 */

import Component from '../base/Component.js';
import VideoCard from './VideoCard.js';

export default class VideoGrid extends Component {
  constructor(container) {
    super(container);
    this.state = {
      videos: [],
      loading: false,
      error: null,
    };
    this.videoCards = [];
  }

  template() {
    if (this.state.loading) {
      return '<div class="loading-state">Cargando videos...</div>';
    }

    if (this.state.error) {
      return `<div class="error-state">Error: ${this.state.error}</div>`;
    }

    if (this.state.videos.length === 0) {
      return '<div class="empty-state">No hay videos disponibles</div>';
    }

    return `
      <div class="video-grid">
        ${this.state.videos.map(video => `
          <div class="video-grid-item" data-video-id="${video.id}"></div>
        `).join('')}
      </div>
    `;
  }

  onMount() {
    this.renderVideoCards();
  }

  onStateChange(oldState, newState) {
    if (oldState.videos !== newState.videos) {
      this.renderVideoCards();
    }
  }

  renderVideoCards() {
    // Limpiar cards anteriores
    this.videoCards.forEach(card => card.unmount());
    this.videoCards = [];

    // Crear nuevos cards
    this.state.videos.forEach(video => {
      const cardContainer = this.$(`.video-grid-item[data-video-id="${video.id}"]`);
      if (cardContainer) {
        const card = new VideoCard(cardContainer, video);
        card.mount();
        this.videoCards.push(card);
      }
    });
  }

  setVideos(videos) {
    this.setState({ videos, loading: false, error: null });
  }

  setLoading(loading) {
    this.setState({ loading });
  }

  setError(error) {
    this.setState({ error, loading: false });
  }

  onUnmount() {
    this.videoCards.forEach(card => card.unmount());
    this.videoCards = [];
  }
}
