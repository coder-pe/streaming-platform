/**
 * @file VideoCard.js
 * @description Componente de tarjeta de video
 */

import Component from '../base/Component.js';
import videoService from '../../services/VideoService.js';
import { formatDuration, formatRelativeDate } from '../../utils/helpers.js';

export default class VideoCard extends Component {
  constructor(container, video) {
    super(container);
    this.video = video;
  }

  template() {
    const thumbnailUrl = videoService.getThumbnailUrl(this.video.thumbnail);

    return `
      <div class="video-card" data-video-id="${this.video.id}">
        <div class="video-thumbnail">
          <img src="${thumbnailUrl || '/static/images/placeholder.jpg'}" alt="${this.video.title}">
          ${this.video.duration ? `<span class="video-duration">${formatDuration(this.video.duration)}</span>` : ''}
          <div class="video-overlay">
            <button class="btn-play" data-action="play">
              <i class="icon-play"></i>
            </button>
          </div>
        </div>
        <div class="video-info">
          <h3 class="video-title">${this.video.title}</h3>
          <div class="video-meta">
            <span class="video-author">${this.video.author || 'Autor desconocido'}</span>
            <span class="video-stats">
              <span>${this.video.views || 0} vistas</span>
              ${this.video.created_at ? `<span>• ${formatRelativeDate(this.video.created_at)}</span>` : ''}
            </span>
          </div>
          ${this.video.description ? `
            <p class="video-description">${this.truncateText(this.video.description, 100)}</p>
          ` : ''}
        </div>
      </div>
    `;
  }

  attachEventListeners() {
    const card = this.$('.video-card');
    if (card) {
      card.addEventListener('click', (e) => {
        if (e.target.closest('[data-action="play"]')) {
          this.handlePlay();
        } else {
          this.handleClick();
        }
      });
    }
  }

  handleClick() {
    this.emit('video:select', this.video);
  }

  handlePlay() {
    this.emit('video:play', this.video);
  }

  truncateText(text, maxLength) {
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  }

  emit(event, data) {
    this.container.dispatchEvent(new CustomEvent(event, {
      detail: data,
      bubbles: true
    }));
  }
}
