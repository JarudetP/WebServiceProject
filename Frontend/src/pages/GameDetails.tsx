import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { gameService } from '../services/game.service';
import type { Game } from '../types';
import toast from 'react-hot-toast';
import { ArrowLeft, Users, Globe, Building2, Briefcase, Activity, Calendar, Terminal, Copy, Check } from 'lucide-react';
import { 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts';

export const GameDetails: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [game, setGame] = useState<Game | null>(null);
  const [history, setHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCopying, setIsCopying] = useState(false);

  useEffect(() => {
    fetchGameData();
    const interval = setInterval(fetchGameData, 10000); // Poll every 10s
    return () => clearInterval(interval);
  }, [id]);

  const fetchGameData = async () => {
    try {
      const [gameData, historyData] = await Promise.all([
        gameService.getGame(Number(id)),
        gameService.getGameHistory(Number(id))
      ]);
      
      if (!gameData || (gameData as any).error) throw new Error('Not found');
      setGame(gameData);
      
      // Format history data for the chart
      const formattedHistory = historyData.map(h => ({
        ...h,
        displayTime: new Date(h.recorded_at).toLocaleDateString('en-US', { 
          month: 'short', 
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        }),
        shortDate: new Date(h.recorded_at).toLocaleDateString('en-US', { 
            month: 'short', 
            day: 'numeric'
        })
      }));
      setHistory(formattedHistory);
    } catch (error) {
      toast.error('Failed to load game details');
      navigate('/games');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <DashboardLayout title="Game Details">
        <div className="flex h-64 items-center justify-center">
            <div className="w-8 h-8 rounded-full border-2 border-primary border-t-foreground animate-spin"></div>
        </div>
      </DashboardLayout>
    );
  }

  if (!game) return null;

  const imageUrl = game.image_url 
    ? (game.image_url.startsWith('http') ? game.image_url : `http://localhost:8080${game.image_url.startsWith('/') ? '' : '/'}${game.image_url}`)
    : '';

  return (
    <DashboardLayout title="Game Details">
      <div className="mb-6 flex items-center">
        <button 
          onClick={() => navigate('/games')}
          className="flex items-center gap-2 text-sm font-medium text-accent hover:text-foreground transition-colors group"
        >
          <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
          Back to Games
        </button>
      </div>

      {/* Hero Section */}
      <div className="bg-white p-8 rounded-[2rem] border border-border shadow-sm mb-6 flex flex-col md:flex-row items-center md:items-start justify-between gap-8 relative overflow-hidden">
        {/* Subtle background decoration */}
        <div className="absolute top-0 right-0 w-64 h-64 bg-gray-50 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 opacity-50 pointer-events-none"></div>

        <div className="flex flex-col md:flex-row items-center md:items-start gap-6 relative z-10 w-full">
           <div className="w-32 h-32 md:w-40 md:h-40 shrink-0 rounded-[1.5rem] overflow-hidden bg-secondary border border-border flex items-center justify-center text-5xl font-bold text-foreground shadow-sm">
             {imageUrl ? (
                <img 
                   src={imageUrl} 
                   alt={game.name} 
                   className="w-full h-full object-cover"
                   onError={(e) => {
                     (e.target as HTMLImageElement).style.display = 'none';
                     (e.target as HTMLImageElement).nextElementSibling?.classList.remove('hidden');
                   }}
                />
             ) : null}
             <span className={`${imageUrl ? 'hidden' : ''}`}>{game.name.charAt(0)}</span>
           </div>

           <div className="flex-1 text-center md:text-left flex flex-col justify-center h-full pt-2">
             <h2 className="text-3xl md:text-4xl font-bold tracking-tight text-foreground mb-3">{game.name}</h2>
             
             <div className="flex flex-wrap items-center justify-center md:justify-start gap-3 mb-6">
               <span className="inline-flex items-center px-4 py-1.5 rounded-full bg-foreground text-background text-xs font-semibold tracking-wide uppercase shadow-sm">
                 {game.genre}
               </span>
               <span className="inline-flex items-center px-4 py-1.5 rounded-full bg-secondary text-foreground text-xs font-medium border border-border shadow-sm">
                 {game.platform}
               </span>
               <span className="inline-flex items-center px-4 py-1.5 rounded-full bg-secondary text-foreground text-xs font-medium border border-border shadow-sm">
                 {game.region}
               </span>
             </div>

             <div className="flex items-center justify-center md:justify-start gap-8 mt-auto">
               <div className="flex items-center gap-3">
                 <div className="w-10 h-10 rounded-full bg-gray-50 border border-gray-100 flex items-center justify-center text-accent">
                   <Briefcase className="w-5 h-5" />
                 </div>
                 <div className="text-left">
                   <p className="text-[11px] font-semibold text-accent uppercase tracking-wider mb-0.5">Developer</p>
                   <p className="text-sm font-medium text-foreground">{game.developer}</p>
                 </div>
               </div>

               <div className="w-px h-10 bg-border hidden sm:block"></div>

               <div className="flex items-center gap-3 hidden sm:flex">
                 <div className="w-10 h-10 rounded-full bg-gray-50 border border-gray-100 flex items-center justify-center text-accent">
                   <Building2 className="w-5 h-5" />
                 </div>
                 <div className="text-left">
                   <p className="text-[11px] font-semibold text-accent uppercase tracking-wider mb-0.5">Publisher</p>
                   <p className="text-sm font-medium text-foreground">{game.publisher}</p>
                 </div>
               </div>
             </div>
           </div>
        </div>

        {/* Live Pulse Indicator (Current Players) */}
        <div className="shrink-0 bg-secondary/50 rounded-3xl p-6 border border-border min-w-[200px] text-center md:text-right relative z-10 w-full md:w-auto">
           <div className="flex items-center justify-center md:justify-end gap-2 mb-2">
             <span className="relative flex h-3 w-3">
               <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
               <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
             </span>
             <h3 className="text-xs font-bold uppercase tracking-wider text-accent ml-1">Live Players</h3>
           </div>
           <span className="text-4xl font-bold tracking-tighter text-foreground">
             {game.current_players?.toLocaleString() || 0}
           </span>
        </div>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div className="bg-white p-8 rounded-[2rem] border border-border shadow-sm flex items-center justify-between group hover:border-gray-300 transition-all">
             <div>
               <div className="flex items-center gap-2 mb-3">
                 <Users className="w-5 h-5 text-accent" />
                 <h3 className="text-xs font-bold uppercase tracking-wider text-accent">Audience Size</h3>
               </div>
               <div className="flex items-baseline gap-2">
                 <span className="text-5xl font-bold tracking-tighter text-foreground">
                   {game.total_players?.toLocaleString() || 0}
                 </span>
               </div>
               <p className="text-sm text-accent mt-3">Total registered accounts worldwide.</p>
             </div>
             <div className="w-16 h-16 rounded-full bg-secondary flex items-center justify-center opacity-50 group-hover:opacity-100 transition-opacity">
               <Globe className="w-8 h-8 text-foreground" />
             </div>
          </div>

          <div className="bg-white p-8 rounded-[2rem] border border-border shadow-sm flex items-center justify-between group hover:border-gray-300 transition-all">
             <div>
               <div className="flex items-center gap-2 mb-3">
                 <Activity className="w-5 h-5 text-accent" />
                 <h3 className="text-xs font-bold uppercase tracking-wider text-accent">Lifetime Revenue</h3>
               </div>
               <div className="flex items-end gap-1.5">
                  <span className="text-5xl font-bold tracking-tighter text-foreground">${(game.revenue || 0).toLocaleString()}</span>
               </div>
               <p className="text-sm text-accent mt-3">Gross revenue across all channels.</p>
             </div>
             <div className="text-right flex flex-col items-end">
               <span className="px-3 py-1.5 mb-4 bg-green-50 text-green-700 text-xs font-bold uppercase tracking-wider rounded-full border border-green-200 shadow-sm">
                 Profitable
               </span>
               <span className="text-sm font-semibold text-accent">USD</span>
             </div>
          </div>
      </div>

      {/* 7-Day Player History Chart */}
      <div className="bg-white p-8 rounded-[2rem] border border-border shadow-sm">
        <div className="flex items-center justify-between mb-8">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <Calendar className="w-5 h-5 text-accent" />
              <h3 className="text-lg font-bold tracking-tight text-foreground">7-Day Player Trend</h3>
            </div>
            <p className="text-sm text-accent">Historical concurrent player data for the last week.</p>
          </div>
          <div className="flex items-center gap-4">
             <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-foreground"></div>
                <span className="text-xs font-medium text-accent">Concurrent Players</span>
             </div>
          </div>
        </div>

        <div className="h-[350px] w-full">
          {history.length > 0 ? (
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={history}>
                <defs>
                  <linearGradient id="colorPlayers" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#111827" stopOpacity={0.1}/>
                    <stop offset="95%" stopColor="#111827" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                <XAxis 
                  dataKey="shortDate" 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  interval={Math.floor(history.length / 7)}
                  minTickGap={30}
                />
                <YAxis 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: '#9ca3af', fontSize: 12 }}
                  tickFormatter={(val) => `${(val / 1000).toFixed(0)}k`}
                />
                <Tooltip 
                  contentStyle={{ 
                    borderRadius: '16px', 
                    border: '1px solid #e5e7eb', 
                    boxShadow: '0 10px 15px -3px rgb(0 0 0 / 0.1)',
                    padding: '12px'
                  }}
                  labelStyle={{ fontWeight: 600, color: '#111827', marginBottom: '4px' }}
                  labelFormatter={(label, payload) => {
                    if (payload && payload[0]) {
                        return payload[0].payload.displayTime;
                    }
                    return label;
                  }}
                />
                <Area 
                  type="monotone" 
                  dataKey="current_players" 
                  name="Players"
                  stroke="#111827" 
                  strokeWidth={3}
                  fillOpacity={1}
                  fill="url(#colorPlayers)" 
                  animationDuration={1500}
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-full flex items-center justify-center text-accent">
              No historical data available.
            </div>
          )}
        </div>
      </div>

      {/* API Integration Sandbox */}
      <div className="bg-white p-8 rounded-[2rem] border border-border shadow-sm mt-6">
        <div className="flex items-center justify-between mb-8">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <Terminal className="w-5 h-5 text-accent" />
              <h3 className="text-lg font-bold tracking-tight text-foreground">API Integration Sandbox</h3>
            </div>
            <p className="text-sm text-accent">Preview and copy the live JSON response for this game.</p>
          </div>
          <button 
             onClick={() => {
               navigator.clipboard.writeText(JSON.stringify(game, null, 2));
               setIsCopying(true);
               toast.success('JSON copied to clipboard');
               setTimeout(() => setIsCopying(false), 2000);
             }}
             className="flex items-center gap-2 px-4 py-2 bg-secondary text-foreground hover:bg-foreground hover:text-background text-sm font-semibold rounded-xl transition-all active:scale-95"
          >
             {isCopying ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
             {isCopying ? 'Copied' : 'Copy JSON'}
          </button>
        </div>

        <div className="space-y-4">
           {/* Endpoint Preview */}
           <div className="p-4 bg-gray-50 border border-border rounded-2xl flex items-center justify-between group">
              <div className="flex items-center gap-3 overflow-hidden">
                 <span className="px-2 py-1 bg-green-100 text-green-700 text-[10px] font-bold uppercase rounded-md">GET</span>
                 <code className="text-xs font-mono text-accent truncate">/api/games/{id}</code>
              </div>
              <span className="text-[10px] font-bold text-accent uppercase tracking-widest opacity-0 group-hover:opacity-100 transition-opacity">Production Endpoint</span>
           </div>

           {/* JSON Code Block */}
           <div className="relative group">
              <div className="absolute top-4 right-4 text-[10px] font-bold text-accent uppercase tracking-widest opacity-40">application/json</div>
              <pre className="p-6 bg-[#1e1e1e] text-blue-100 rounded-2xl overflow-x-auto text-xs font-mono custom-scrollbar max-h-80 leading-relaxed shadow-inner">
                {JSON.stringify(game, null, 2)}
              </pre>
           </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
