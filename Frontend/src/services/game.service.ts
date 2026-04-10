import api from './api';
import type { Game } from '../types';

export const gameService = {
  getGames: async (): Promise<Game[]> => {
    const response = await api.get('/games');
    return response.data;
  },

  getGame: async (id: number): Promise<Game> => {
    const response = await api.get(`/games/${id}`);
    return response.data;
  }
};
