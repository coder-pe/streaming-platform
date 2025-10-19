/**
 * Service Worker for StreamLearn PWA
 */

const CACHE_NAME = 'streamlearn-v1';
const STATIC_CACHE = 'streamlearn-static-v1';
const DYNAMIC_CACHE = 'streamlearn-dynamic-v1';

// Recursos estáticos a cachear
const STATIC_ASSETS = [
  '/',
  '/static/css/styles.css',
  '/static/js/app.js',
  '/static/manifest.json'
];

// Instalar Service Worker
self.addEventListener('install', (event) => {
  console.log('[SW] Installing Service Worker...');

  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => {
        console.log('[SW] Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .catch((error) => {
        console.log('[SW] Error caching static assets:', error);
      })
  );

  // Forzar activación inmediata
  self.skipWaiting();
});

// Activar Service Worker
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating Service Worker...');

  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName !== STATIC_CACHE && cacheName !== DYNAMIC_CACHE) {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
  );

  // Tomar control inmediatamente
  return self.clients.claim();
});

// Interceptar peticiones
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Ignorar peticiones a otros dominios
  if (url.origin !== location.origin) {
    return;
  }

  // Estrategia: Network First para API, Cache First para estáticos
  if (url.pathname.startsWith('/api')) {
    // API: Network First (siempre intentar red primero)
    event.respondWith(networkFirst(request));
  } else if (url.pathname.startsWith('/static')) {
    // Estáticos: Cache First (cache primero, red si falla)
    event.respondWith(cacheFirst(request));
  } else {
    // HTML: Network First con fallback
    event.respondWith(networkFirst(request));
  }
});

/**
 * Estrategia Cache First
 */
async function cacheFirst(request) {
  const cache = await caches.open(STATIC_CACHE);
  const cached = await cache.match(request);

  if (cached) {
    return cached;
  }

  try {
    const response = await fetch(request);

    // Solo cachear respuestas exitosas
    if (response && response.status === 200) {
      cache.put(request, response.clone());
    }

    return response;
  } catch (error) {
    console.log('[SW] Fetch failed:', error);
    throw error;
  }
}

/**
 * Estrategia Network First
 */
async function networkFirst(request) {
  const cache = await caches.open(DYNAMIC_CACHE);

  try {
    const response = await fetch(request);

    // Cachear respuesta exitosa
    if (response && response.status === 200) {
      cache.put(request, response.clone());
    }

    return response;
  } catch (error) {
    // Si falla la red, buscar en cache
    const cached = await cache.match(request);

    if (cached) {
      return cached;
    }

    throw error;
  }
}

// Manejar mensajes del cliente
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (event.data && event.data.type === 'CLEAR_CACHE') {
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => caches.delete(cacheName))
      );
    });
  }
});
