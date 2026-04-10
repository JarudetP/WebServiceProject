import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { gameService } from '../services/game.service';
import type { Game } from '../types';
import toast from 'react-hot-toast';
import { Search } from 'lucide-react';

export const Games: React.FC = () => {
  const navigate = useNavigate();
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchGames();
  }, []);

  const fetchGames = async () => {
    try {
      const data = await gameService.getGames();
      setGames(data);
    } catch (error) {
      toast.error('Failed to load games');
    } finally {
      setLoading(false);
    }
  };

  return (
    <DashboardLayout title="Games Library">
      <div className="bg-white rounded-2xl border border-border shadow-sm overflow-hidden flex flex-col min-h-[500px]">
        {/* Table Toolbar */}
        <div className="p-4 border-b border-border flex items-center justify-between bg-primary/30">
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-accent" />
            <input 
              type="text" 
              placeholder="Search games..." 
              className="w-full pl-9 pr-4 py-2 bg-white border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all"
            />
          </div>
        </div>

        {/* Table Content */}
        <div className="flex-1 overflow-x-auto">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-border bg-gray-50/50">
                <th className="px-6 py-4 text-xs font-semibold text-accent uppercase tracking-wider">Game</th>
                <th className="px-6 py-4 text-xs font-semibold text-accent uppercase tracking-wider">Genre</th>
                <th className="px-6 py-4 text-xs font-semibold text-accent uppercase tracking-wider">Platform</th>
                <th className="px-6 py-4 text-xs font-semibold text-accent uppercase tracking-wider text-right">Players</th>
                <th className="px-6 py-4 text-xs font-semibold text-accent uppercase tracking-wider text-right">Revenue</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {loading ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-accent text-sm">
                    Loading games...
                  </td>
                </tr>
              ) : games.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center text-accent text-sm">
                    No games found.
                  </td>
                </tr>
              ) : (
                games.map((game) => (
                  <tr 
                    key={game.id} 
                    className="hover:bg-primary/50 transition-colors cursor-pointer group"
                    onClick={() => navigate(`/games/${game.id}`)}
                  >
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        {game.image_url ? (
                          <img 
                            src={game.image_url.startsWith('http') ? game.image_url : `http://localhost:8080${game.image_url.startsWith('/') ? '' : '/'}${game.image_url}`} 
                            alt={game.name} 
                            className="w-10 h-10 rounded-lg object-cover bg-secondary border border-border"
                            onError={(e) => {
                              // Fallback if image fails to load
                              (e.target as HTMLImageElement).style.display = 'none';
                              (e.target as HTMLImageElement).nextElementSibling?.classList.remove('hidden');
                            }}
                          />
                        ) : null}
                        
                        {/* Fallback character box */}
                        <div className={`w-10 h-10 rounded-lg bg-secondary flex items-center justify-center font-bold text-foreground ${game.image_url ? 'hidden' : ''}`}>
                          {game.name.charAt(0)}
                        </div>

                        <div>
                          <p className="text-sm font-medium text-foreground">{game.name}</p>
                          <p className="text-xs text-accent">{game.developer}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-foreground">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full bg-secondary text-xs font-medium">
                        {game.genre}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-foreground">{game.platform}</td>
                    <td className="px-6 py-4 text-sm text-foreground text-right">{game.total_players.toLocaleString()}</td>
                    <td className="px-6 py-4 text-sm font-medium text-foreground text-right">${game.revenue.toLocaleString()}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>
    </DashboardLayout>
  );
};
