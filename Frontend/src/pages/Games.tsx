import React, { useEffect, useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { gameService } from '../services/game.service';
import { authService } from '../services/auth.service';
import type { Game } from '../types';
import toast from 'react-hot-toast';
import { Search, Plus, Pencil, Trash2, Filter, Copy, Check, Link } from 'lucide-react';
import { GameModal } from '../components/games/GameModal';

export const Games: React.FC = () => {
  const navigate = useNavigate();
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedGame, setSelectedGame] = useState<Game | null>(null);
  const [userRole, setUserRole] = useState<string>('');
  
  // Search and Filter State
  const [searchTerm, setSearchTerm] = useState('');
  const [genreFilter, setGenreFilter] = useState('All Genres');
  const [platformFilter, setPlatformFilter] = useState('All Platforms');
  const [copyingId, setCopyingId] = useState<number | null>(null);
  const [copyingPathId, setCopyingPathId] = useState<number | null>(null);

  useEffect(() => {
    fetchGames();
    const user = authService.getCurrentUser();
    if (user) setUserRole(user.role);
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

  const filteredGames = useMemo(() => {
    return games.filter(game => {
      const matchesSearch = game.name.toLowerCase().includes(searchTerm.toLowerCase()) || 
                           game.developer.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesGenre = genreFilter === 'All Genres' || game.genre === genreFilter;
      const matchesPlatform = platformFilter === 'All Platforms' || game.platform === platformFilter;
      
      return matchesSearch && matchesGenre && matchesPlatform;
    });
  }, [games, searchTerm, genreFilter, platformFilter]);

  // Extract unique genres and platforms for filters
  const genres = useMemo(() => ['All Genres', ...new Set(games.map(g => g.genre))], [games]);
  const platforms = useMemo(() => ['All Platforms', ...new Set(games.map(g => g.platform))], [games]);

  const handleSave = async (formData: FormData) => {
    try {
      if (selectedGame) {
        await gameService.updateGame(selectedGame.id, formData);
        toast.success('Game updated successfully');
      } else {
        await gameService.createGame(formData);
        toast.success('Game created successfully');
      }
      fetchGames();
    } catch (error) {
      toast.error(selectedGame ? 'Failed to update game' : 'Failed to create game');
    }
  };

  const handleDelete = async (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (!window.confirm('Are you sure you want to delete this game?')) return;
    
    try {
      await gameService.deleteGame(id);
      toast.success('Game deleted successfully');
      fetchGames();
    } catch (error) {
      toast.error('Failed to delete game');
    }
  };

  const handleCopyJson = (e: React.MouseEvent, game: Game) => {
    e.stopPropagation();
    navigator.clipboard.writeText(JSON.stringify(game, null, 2));
    setCopyingId(game.id);
    toast.success('JSON copied to clipboard');
    setTimeout(() => setCopyingId(null), 2000);
  };

  const handleCopyPath = (e: React.MouseEvent, game: Game) => {
    e.stopPropagation();
    navigator.clipboard.writeText(`http://localhost:8080/api/games/${game.id}`);
    setCopyingPathId(game.id);
    toast.success('API path copied to clipboard');
    setTimeout(() => setCopyingPathId(null), 2000);
  };

  const openModal = (e?: React.MouseEvent, game?: Game) => {
    if (e) e.stopPropagation();
    setSelectedGame(game || null);
    setIsModalOpen(true);
  };

  const isAdmin = userRole === 'admin';

  return (
    <DashboardLayout title="Games Library">
      <div className="bg-white rounded-3xl border border-border shadow-sm overflow-hidden flex flex-col min-h-[500px]">
        {/* Table Toolbar */}
        <div className="p-6 border-b border-border flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gray-50/30">
          <div className="flex flex-col md:flex-row gap-4 flex-1">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-accent" />
              <input 
                type="text" 
                placeholder="Search games or developers..." 
                className="w-full pl-10 pr-4 py-2.5 bg-white border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all shadow-sm"
                value={searchTerm}
                onChange={e => setSearchTerm(e.target.value)}
              />
            </div>
            
            <div className="flex gap-2">
              <select 
                className="px-4 py-2.5 bg-white border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 appearance-none min-w-[140px] shadow-sm cursor-pointer"
                value={genreFilter}
                onChange={e => setGenreFilter(e.target.value)}
              >
                {genres.map(g => <option key={g} value={g}>{g}</option>)}
              </select>
              
              <select 
                className="px-4 py-2.5 bg-white border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary/20 appearance-none min-w-[140px] shadow-sm cursor-pointer"
                value={platformFilter}
                onChange={e => setPlatformFilter(e.target.value)}
              >
                {platforms.map(p => <option key={p} value={p}>{p}</option>)}
              </select>
            </div>
          </div>
          
          {isAdmin && (
            <button 
              onClick={() => openModal()}
              className="flex items-center gap-2 px-6 py-2.5 bg-foreground text-background rounded-xl text-sm font-bold hover:bg-foreground/90 transition-all active:scale-95 shadow-lg shadow-foreground/10"
            >
              <Plus className="w-4 h-4" />
              Add New Game
            </button>
          )}
        </div>

        {/* Table Content */}
        <div className="flex-1 overflow-x-auto custom-scrollbar">
          <table className="w-full text-left border-collapse">
            <thead>
              <tr className="border-b border-border bg-gray-50/50">
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em]">Game</th>
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em]">Genre</th>
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em]">Platform</th>
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em] text-right">Players</th>
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em] text-right">Revenue</th>
                <th className="px-6 py-4 text-[10px] font-bold text-accent uppercase tracking-[0.1em] text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {loading ? (
                <tr>
                  <td colSpan={6} className="px-6 py-20 text-center">
                    <div className="flex flex-col items-center gap-3">
                       <div className="w-8 h-8 border-3 border-primary/20 border-t-primary rounded-full animate-spin"></div>
                       <p className="text-sm font-medium text-accent">Loading games library...</p>
                    </div>
                  </td>
                </tr>
              ) : filteredGames.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-6 py-20 text-center">
                    <div className="flex flex-col items-center gap-2 opacity-40">
                       <Filter className="w-10 h-10 mb-2" />
                       <p className="text-lg font-bold text-foreground">No games match your filters</p>
                       <p className="text-sm text-accent">Try adjusting your search terms or filters</p>
                    </div>
                  </td>
                </tr>
              ) : (
                filteredGames.map((game) => (
                  <tr 
                    key={game.id} 
                    className="hover:bg-primary/5 transition-colors cursor-pointer group"
                    onClick={() => navigate(`/games/${game.id}`)}
                  >
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-4">
                        <div className="relative">
                          {game.image_url ? (
                            <img 
                              src={game.image_url.startsWith('http') ? game.image_url : `http://localhost:8080${game.image_url.startsWith('/') ? '' : '/'}${game.image_url}`} 
                              alt={game.name} 
                              className="w-12 h-12 rounded-xl object-cover bg-secondary border border-border shadow-sm group-hover:scale-105 transition-transform"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display = 'none';
                                (e.target as HTMLImageElement).parentElement?.querySelector('.fallback')?.classList.remove('hidden');
                              }}
                            />
                          ) : null}
                          <div className={`fallback w-12 h-12 rounded-xl bg-secondary flex items-center justify-center font-bold text-foreground border border-border ${game.image_url ? 'hidden' : ''}`}>
                            {game.name.charAt(0)}
                          </div>
                        </div>

                        <div>
                          <p className="text-sm font-bold text-foreground group-hover:text-primary transition-colors">{game.name}</p>
                          <p className="text-[11px] font-medium text-accent uppercase tracking-wider">{game.developer}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-foreground">
                      <span className="inline-flex items-center px-3 py-1 rounded-full bg-secondary text-[11px] font-bold uppercase tracking-tight text-accent">
                        {game.genre}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm font-medium text-accent">{game.platform}</td>
                    <td className="px-6 py-4 text-sm font-bold text-foreground text-right">{game.total_players.toLocaleString()}</td>
                    <td className="px-6 py-4 text-sm font-bold text-foreground text-right text-emerald-600">${game.revenue.toLocaleString()}</td>
                    <td className="px-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-1 px-2">
                        <button 
                          onClick={(e) => handleCopyJson(e, game)}
                          className="p-2 text-accent hover:text-primary hover:bg-primary/10 rounded-xl transition-all"
                          title="Copy JSON Response"
                        >
                          {copyingId === game.id ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                        </button>
                        
                        {isAdmin && (
                          <>
                            <button 
                              onClick={(e) => openModal(e, game)}
                              className="p-2 text-accent hover:text-primary hover:bg-primary/10 rounded-xl transition-all"
                              title="Edit Game"
                            >
                              <Pencil className="w-4 h-4" />
                            </button>
                            <button 
                              onClick={(e) => handleDelete(e, game.id)}
                              className="p-2 text-accent hover:text-red-600 hover:bg-red-50 rounded-xl transition-all"
                              title="Delete Game"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <GameModal 
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSave={handleSave}
        game={selectedGame}
      />
      
      <style>{`
        .custom-scrollbar::-webkit-scrollbar { height: 8px; width: 8px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: #e5e7eb; border-radius: 10px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #d1d5db; }
      `}</style>
    </DashboardLayout>
  );
};
