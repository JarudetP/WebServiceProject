import api from './api';

export interface Package {
  id: number;
  name: string;
  description: string;
  request_limit: number;
  refresh_interval_minutes: number;
  price: number;
  duration_days: number;
}

export interface Subscription {
  id: number;
  user_id: number;
  package_id: number;
  start_date: string;
  end_date: string;
  is_active: boolean;
}

export const packageService = {
  getPackages: async (): Promise<Package[]> => {
    const response = await api.get('/packages');
    return response.data;
  },

  getActiveSubscription: async (userId: number): Promise<Subscription | null> => {
    try {
      const response = await api.get(`/packages/subscription?user_id=${userId}`);
      return response.data;
    } catch {
      return null;
    }
  },

  purchasePackage: async (userId: number, packageId: number): Promise<{ message: string }> => {
    const response = await api.post(`/packages/purchase?user_id=${userId}`, { package_id: packageId });
    return response.data;
  }
};
