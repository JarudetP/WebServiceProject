export interface User {
  id: number;
  username: string;
  email: string;
  full_name: string;
  company: string;
  role: string;
  balance: number;
  is_active: boolean;
}

export interface Game {
  id: number;
  name: string;
  total_players: number;
  current_players: number;
  revenue: number;
  genre: string;
  region: string;
  platform: string;
  publisher: string;
  developer: string;
  image_url: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User; // Backend profile endpoint might need to be called
}

export interface ApiKey {
  id: number;
  key_hash: string;
  user_id: number;
  created_at: string;
  is_active: boolean;
}
