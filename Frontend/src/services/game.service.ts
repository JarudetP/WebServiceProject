import { gameApi as api } from './api';
import type { Game } from '../types';

export const gameService = {
  getGames: async (): Promise<Game[]> => {
    const response = await api.get('/games');
    return response.data || [];
  },

  getGame: async (id: number): Promise<Game> => {
    const response = await api.get(`/games/${id}`);
    return response.data;
  },

  getGameHistory: async (id: number): Promise<any[]> => {
    const response = await api.get(`/games/${id}/history`);
    return response.data || [];
  },

  createGame: async (formData: FormData): Promise<Game> => {
    const response = await api.post('/games', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data;
  },

  updateGame: async (id: number, formData: FormData): Promise<void> => {
    await api.put(`/games/${id}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  deleteGame: async (id: number): Promise<void> => {
    await api.delete(`/games/${id}`);
  }
};
