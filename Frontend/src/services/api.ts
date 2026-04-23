import axios, { type AxiosInstance } from 'axios';

// Base URLs populated via environment variables
const USER_API_URL = import.meta.env.VITE_USER_API_URL || 'http://localhost:8081/api';
const PACKAGE_API_URL = import.meta.env.VITE_PACKAGE_API_URL || 'http://localhost:8082/api';
const GAME_API_URL = import.meta.env.VITE_GAME_API_URL || 'http://localhost:8083/api';

const createInstance = (baseURL: string): AxiosInstance => {
  return axios.create({
    baseURL: baseURL,
    headers: {
      'Content-Type': 'application/json',
    },
  });
};

export const userApi = createInstance(USER_API_URL);
export const packageApi = createInstance(PACKAGE_API_URL);
export const gameApi = createInstance(GAME_API_URL);

const allInstances = [userApi, packageApi, gameApi];

export const getAuthToken = () => localStorage.getItem('access_token');
export const getApiKey = () => localStorage.getItem('api_key');

// Request Interceptor for all instances
allInstances.forEach(api => {
  api.interceptors.request.use(
    (config) => {
      const token = getAuthToken();
      const apiKey = getApiKey();

      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
      }

      // Attach API Key for game service calls
      if (api === gameApi && apiKey) {
        config.headers['X-API-Key'] = apiKey;
      }

      return config;
    },
    (error) => Promise.reject(error)
  );
});

// Response Interceptor for Token Refresh (on all instances)
let isRefreshing = false;
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = [];

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token!);
    }
  });
  failedQueue = [];
};

allInstances.forEach(api => {
  api.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config;

      if (originalRequest.url?.includes('/users/login') ||
          originalRequest.url?.includes('/users/register') ||
          originalRequest.url?.includes('/users/refresh')) {
        return Promise.reject(error);
      }

      if (error.response?.status === 401 && !originalRequest._retry) {
        if (isRefreshing) {
          return new Promise<string>((resolve, reject) => {
            failedQueue.push({ resolve, reject });
          }).then(token => {
            originalRequest.headers['Authorization'] = `Bearer ${token}`;
            return api(originalRequest);
          });
        }

        originalRequest._retry = true;
        isRefreshing = true;

        const refreshToken = localStorage.getItem('refresh_token');
        if (!refreshToken) {
          handleAuthFailure();
          return Promise.reject(error);
        }

        try {
          // Always use userApi for refresh
          const response = await axios.post(`${USER_API_URL}/users/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token } = response.data;
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);

          processQueue(null, access_token);
          originalRequest.headers['Authorization'] = `Bearer ${access_token}`;
          return api(originalRequest);
        } catch (refreshError) {
          processQueue(refreshError, null);
          handleAuthFailure();
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }

      return Promise.reject(error);
    }
  );
});

const handleAuthFailure = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('api_key');
  if (window.location.pathname !== '/login') {
    window.location.href = '/login';
  }
};

export default userApi; // Default export for compatibility if needed
