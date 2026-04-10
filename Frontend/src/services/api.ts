import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const getAuthToken = () => localStorage.getItem('access_token');
export const getApiKey = () => localStorage.getItem('api_key');

api.interceptors.request.use(
  (config) => {
    // If it's a game endpoint, attach api key
    if (config.url?.includes('/games')) {
      const apiKey = getApiKey();
      if (apiKey) {
        config.headers['X-API-Key'] = apiKey;
      }
    } else {
      // Otherwise, standard bearer auth (for profile, packages, etc)
      const token = getAuthToken();
      if (token) {
        config.headers['Authorization'] = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

export default api;
