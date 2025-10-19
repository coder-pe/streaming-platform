/**
 * @file ApiService.js
 * @description Cliente HTTP base para comunicación con la API
 */

import { API_CONFIG } from '../config/constants.js';
import eventBus from '../core/EventBus.js';

class ApiService {
  constructor() {
    this.baseURL = API_CONFIG.BASE_URL;
    this.timeout = API_CONFIG.TIMEOUT;
    this.maxRetries = API_CONFIG.MAX_RETRIES;
  }

  /**
   * Obtener headers por defecto
   */
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };

    const token = localStorage.getItem('authToken');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    return headers;
  }

  /**
   * Realizar petición HTTP con timeout y reintentos
   */
  async request(url, options = {}, retries = 0) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.baseURL}${url}`, {
        ...options,
        headers: {
          ...this.getHeaders(),
          ...options.headers,
        },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      // Manejar errores HTTP
      if (!response.ok) {
        const error = await this.handleError(response);
        throw error;
      }

      // Intentar parsear JSON
      const contentType = response.headers.get('content-type');
      if (contentType && contentType.includes('application/json')) {
        return await response.json();
      }

      return response;
    } catch (error) {
      clearTimeout(timeoutId);

      // Reintentar en caso de error de red
      if (retries < this.maxRetries && this.shouldRetry(error)) {
        await this.sleep(1000 * (retries + 1)); // Backoff exponencial
        return this.request(url, options, retries + 1);
      }

      throw error;
    }
  }

  /**
   * Manejar errores HTTP
   */
  async handleError(response) {
    let errorData;
    try {
      errorData = await response.json();
    } catch {
      errorData = { message: response.statusText };
    }

    const error = new Error(errorData.message || 'Error en la petición');
    error.status = response.status;
    error.data = errorData;

    // Emitir evento de error
    eventBus.emit('api:error', error);

    // Manejar casos especiales
    if (response.status === 401) {
      eventBus.emit('auth:unauthorized');
    } else if (response.status === 403) {
      eventBus.emit('auth:forbidden');
    }

    return error;
  }

  /**
   * Determinar si se debe reintentar la petición
   */
  shouldRetry(error) {
    // Reintentar solo en errores de red, no en errores HTTP
    return error.name === 'AbortError' || error.message.includes('fetch');
  }

  /**
   * Esperar un tiempo determinado
   */
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * GET request
   */
  async get(url, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const fullUrl = queryString ? `${url}?${queryString}` : url;

    return this.request(fullUrl, {
      method: 'GET',
    });
  }

  /**
   * POST request
   */
  async post(url, data) {
    return this.request(url, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  /**
   * PUT request
   */
  async put(url, data) {
    return this.request(url, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  /**
   * DELETE request
   */
  async delete(url) {
    return this.request(url, {
      method: 'DELETE',
    });
  }

  /**
   * Upload de archivo con seguimiento de progreso
   */
  async upload(url, formData, onProgress) {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();

      // Configurar headers
      const token = localStorage.getItem('authToken');
      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`);
      }

      // Seguimiento de progreso
      if (onProgress) {
        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable) {
            const percentComplete = (e.loaded / e.total) * 100;
            onProgress(percentComplete, e.loaded, e.total);
          }
        });
      }

      // Manejar respuesta
      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText);
            resolve(response);
          } catch {
            resolve(xhr.responseText);
          }
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}`));
        }
      });

      // Manejar error
      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'));
      });

      // Manejar timeout
      xhr.addEventListener('timeout', () => {
        reject(new Error('Upload timeout'));
      });

      // Enviar petición
      xhr.open('POST', `${this.baseURL}${url}`);
      xhr.timeout = this.timeout;
      xhr.send(formData);
    });
  }
}

// Exportar instancia singleton
export default new ApiService();
