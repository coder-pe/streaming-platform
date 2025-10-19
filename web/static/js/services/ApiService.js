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
    this.isRefreshing = false;
    this.refreshSubscribers = [];
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

        // Si es error 401 y tenemos refresh token, intentar refrescar
        if (error.shouldRetryWithRefresh && !url.includes('/auth/refresh')) {
          try {
            // Si ya se está refrescando, esperar
            if (this.isRefreshing) {
              return new Promise((resolve, reject) => {
                this.addRefreshSubscriber((token) => {
                  // Reintentar la petición original con el nuevo token
                  this.request(url, options, 0)
                    .then(resolve)
                    .catch(reject);
                });
              });
            }

            // Intentar refrescar el token
            this.isRefreshing = true;
            const newToken = await this.refreshAuthToken();
            this.isRefreshing = false;
            this.onRefreshed(newToken);

            // Reintentar la petición original con el nuevo token
            return await this.request(url, options, 0);
          } catch (refreshError) {
            this.isRefreshing = false;
            this.refreshSubscribers = [];

            // Solo limpiar sesión si el refresh token es inválido (401/403)
            // No limpiar por errores de red
            if (refreshError.status === 401 || refreshError.status === 403) {
              console.log('Refresh token inválido, limpiando sesión...');
              localStorage.removeItem('authToken');
              localStorage.removeItem('refreshToken');
              localStorage.removeItem('userData');
              eventBus.emit('auth:unauthorized');
            } else {
              console.log('Error de red al refrescar token, manteniendo sesión en caché');
            }

            throw error;
          }
        }

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

      // Si el error tiene shouldRetryWithRefresh, ya fue manejado arriba
      if (error.shouldRetryWithRefresh) {
        throw error;
      }

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
      // No emitir auth:unauthorized inmediatamente, se manejará en request()
      error.shouldRetryWithRefresh = true;
    } else if (response.status === 403) {
      eventBus.emit('auth:forbidden');
    }

    return error;
  }

  /**
   * Intentar refrescar el token de autenticación
   */
  async refreshAuthToken() {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      const error = new Error('No refresh token available');
      error.status = 401;
      throw error;
    }

    // Hacer la petición de refresh sin pasar por el interceptor
    const response = await fetch(`${this.baseURL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (!response.ok) {
      const error = new Error('Failed to refresh token');
      error.status = response.status;
      throw error;
    }

    const data = await response.json();

    if (data.token) {
      localStorage.setItem('authToken', data.token);
      if (data.refresh_token) {
        localStorage.setItem('refreshToken', data.refresh_token);
      }
    }

    return data.token;
  }

  /**
   * Agregar petición a la cola de espera durante el refresh
   */
  onRefreshed(token) {
    this.refreshSubscribers.forEach(callback => callback(token));
    this.refreshSubscribers = [];
  }

  /**
   * Agregar callback a la cola de suscriptores
   */
  addRefreshSubscriber(callback) {
    this.refreshSubscribers.push(callback);
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

      // Enviar petición
      xhr.open('POST', `${this.baseURL}${url}`);
      xhr.timeout = this.timeout;

      // Configurar headers (después de open)
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

      xhr.send(formData);
    });
  }
}

// Exportar instancia singleton
export default new ApiService();
