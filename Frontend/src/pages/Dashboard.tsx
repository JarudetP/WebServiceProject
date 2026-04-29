import React, { useEffect, useState } from 'react';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { gameService } from '../services/game.service';
import type { Game } from '../types';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { Users, Banknote, Gamepad2, Activity } from 'lucide-react';
import toast from 'react-hot-toast';

export const Dashboard: React.FC = () => {
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchGames();
    const interval = setInterval(fetchGames, 1800000); // Poll every 30m
    return () => clearInterval(interval);
  }, []);

  const fetchGames = async () => {
    try {
      const data = await gameService.getGames();
      setGames(data);
    } catch (error) {
      toast.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const totalPlayers = games.reduce((sum, g) => sum + g.total_players, 0);
  const totalRevenue = games.reduce((sum, g) => sum + g.revenue, 0);
  const activePlayers = games.reduce((sum, g) => sum + g.current_players, 0);

  // Group games by genre for the pie chart
  const genreData = games.reduce((acc: any[], game) => {
    const existing = acc.find(item => item.name === game.genre);
    if (existing) {
      existing.value += game.total_players;
    } else {
      acc.push({ name: game.genre, value: game.total_players });
    }
    return acc;
  }, []);

  const COLORS = ['#111827', '#4B5563', '#9CA3AF', '#D1D5DB', '#E5E7EB'];

  const metrics = [
    { title: 'Total Games', value: games.length, icon: Gamepad2 },
    { title: 'Global Players', value: totalPlayers.toLocaleString(), icon: Users },
    { title: 'Active Right Now', value: activePlayers.toLocaleString(), icon: Activity },
    { title: 'Total Revenue', value: `฿${totalRevenue.toLocaleString()}`, icon: Banknote },
  ];

  return (
    <DashboardLayout title="Overview">
      <div className="space-y-6">
        {/* State Cards */}
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {metrics.map((metric) => {
            const Icon = metric.icon;
            return (
              <div key={metric.title} className="bg-white p-6 rounded-2xl border border-border shadow-sm flex items-start justify-between">
                <div>
                  <p className="text-sm font-medium text-accent">{metric.title}</p>
                  <p className="mt-2 text-3xl font-semibold tracking-tight text-foreground">
                    {loading ? '-' : metric.value}
                  </p>
                </div>
                <div className="p-3 bg-secondary rounded-xl text-foreground">
                  <Icon className="w-5 h-5" strokeWidth={2} />
                </div>
              </div>
            );
          })}
        </div>

        {/* Charts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm lg:col-span-2">
            <h3 className="text-lg font-medium tracking-tight mb-6">Audience Distribution by Genre</h3>
            <div className="h-80 w-full flex items-center justify-center">
              {loading ? (
                <div className="w-8 h-8 rounded-full border-2 border-primary border-t-foreground animate-spin"></div>
              ) : genreData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={genreData}
                      cx="50%"
                      cy="50%"
                      innerRadius={60}
                      outerRadius={100}
                      paddingAngle={2}
                      dataKey="value"
                      stroke="none"
                    >
                      {genreData.map((_entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip 
                      contentStyle={{ borderRadius: '12px', borderColor: '#e5e7eb', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                      itemStyle={{ color: '#111827', fontSize: '14px', fontWeight: 500 }}
                    />
                    <Legend iconType="circle" wrapperStyle={{ fontSize: '13px' }} />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-accent text-sm">No data available to display chart.</p>
              )}
            </div>
          </div>

          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm flex flex-col justify-between">
            <div>
              <h3 className="text-lg font-medium tracking-tight mb-2">Performance Focus</h3>
              <p className="text-sm text-accent mb-6 leading-relaxed">
                Stay updated with the latest trends in the gaming landscape. Utilize this dashboard to adapt your package subscriptions and monitor your API usages in real time.
              </p>
            </div>
            <div className="bg-primary p-4 rounded-xl border border-border">
              <p className="text-xs font-semibold text-accent uppercase tracking-wider mb-1">System Status</p>
              <div className="flex items-center gap-2">
                <span className="relative flex h-3 w-3">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
                </span>
                <span className="text-sm font-medium text-foreground">API Connection Stable</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
};
