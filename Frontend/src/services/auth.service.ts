import api from './api';
import type { User, AuthResponse, ApiKey } from '../types';

export const authService = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post('/users/login', { email, password });
    if (response.data.access_token) {
      localStorage.setItem('access_token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }
    return response.data;
  },

  register: async (data: Partial<User & { password: string }>) => {
    const response = await api.post('/users/register', data);
    return response.data;
  },

  logout: () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('api_key');
  },

  getCurrentUser: (): { user_id: number; username: string; role: string } | null => {
    const token = localStorage.getItem('access_token');
    if (!token) return null;
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return { 
        user_id: payload.user_id, 
        username: payload.username, 
        role: payload.role 
      };
    } catch {
      return null;
    }
  },

  getProfile: async (id: number): Promise<User> => {
    const response = await api.get(`/users/${id}`);
    return response.data;
  },

  getKeys: async (userId: number): Promise<ApiKey[]> => {
    const response = await api.get(`/users/${userId}/keys`);
    return response.data.api_keys || [];
  },
  
  generateKey: async (userId: number): Promise<{ apiKey: string }> => {
    const response = await api.post(`/users/${userId}/keys`);
    // Will return plain api_key we can use. The backend actually might return it differently, adjust later.
    return response.data;
  },

  deleteKey: async (userId: number, key: string): Promise<void> => {
    await api.delete(`/users/${userId}/keys/${key}`);
  }
};
