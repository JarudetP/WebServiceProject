import React, { useEffect, useState, useRef } from 'react';
import { Link } from 'react-router-dom';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { gameService } from '../services/game.service';
import { GAME_API_BASE, getApiKey } from '../services/api';
import type { GenreAnalytic, RegionAnalytic, RevenueEntry, Game } from '../types';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from 'recharts';
import { Lock, Download, Radio, Users, DollarSign, Globe, BarChart2 } from 'lucide-react';
import toast from 'react-hot-toast';

const LockedCard: React.FC<{ plan: string }> = ({ plan }) => (
  <div className="flex flex-col items-center justify-center py-14 gap-3 text-center">
    <div className="p-3 bg-secondary rounded-2xl">
      <Lock className="w-6 h-6 text-accent" />
    </div>
    <p className="text-sm font-medium text-foreground">Requires {plan}</p>
    <p className="text-xs text-accent">Upgrade your plan to unlock this feature</p>
    <Link
      to="/profile"
      className="mt-1 px-4 py-2 bg-foreground text-background text-xs font-semibold rounded-lg hover:bg-foreground/80 transition-colors"
    >
      Upgrade Plan
    </Link>
  </div>
);

const SectionShell: React.FC<{ title: string; icon: React.ReactNode; children: React.ReactNode }> = ({ title, icon, children }) => (
  <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
    <div className="flex items-center gap-2 mb-6">
      <div className="text-accent">{icon}</div>
      <h3 className="text-lg font-medium tracking-tight">{title}</h3>
    </div>
    {children}
  </div>
);

export const Analytics: React.FC = () => {
  const [genreData, setGenreData] = useState<GenreAnalytic[]>([]);
  const [revenueData, setRevenueData] = useState<RevenueEntry[]>([]);
  const [regionData, setRegionData] = useState<RegionAnalytic[]>([]);
  const [streamData, setStreamData] = useState<Game[]>([]);

  const [genreLocked, setGenreLocked] = useState(false);
  const [revenueLocked, setRevenueLocked] = useState(false);
  const [regionLocked, setRegionLocked] = useState(false);
  const [exportLocked, setExportLocked] = useState(false);
  const [streamLocked, setStreamLocked] = useState(false);

  const [loading, setLoading] = useState(true);
  const [streamConnected, setStreamConnected] = useState(false);
  const [exportLoading, setExportLoading] = useState(false);
  const streamAbortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    fetchAll();
    connectStream();
    return () => streamAbortRef.current?.abort();
  }, []);

  const fetchAll = async () => {
    try {
      const [genre, revenue, region] = await Promise.allSettled([
        gameService.getGenreAnalytics(),
        gameService.getRevenueAnalytics(),
        gameService.getRegionBreakdown(),
      ]);

      if (genre.status === 'fulfilled') setGenreData(genre.value);
      else if ((genre.reason as any)?.response?.status === 403) setGenreLocked(true);

      if (revenue.status === 'fulfilled') setRevenueData(revenue.value);
      else if ((revenue.reason as any)?.response?.status === 403) setRevenueLocked(true);

      if (region.status === 'fulfilled') setRegionData(region.value);
      else if ((region.reason as any)?.response?.status === 403) setRegionLocked(true);
    } finally {
      setLoading(false);
    }
  };

  const connectStream = async () => {
    const apiKey = getApiKey();
    if (!apiKey) return;

    const controller = new AbortController();
    streamAbortRef.current = controller;

    try {
      const response = await fetch(`${GAME_API_BASE}/games/stream`, {
        headers: { 'X-API-Key': apiKey },
        signal: controller.signal,
      });

      if (response.status === 403) { setStreamLocked(true); return; }
      if (!response.ok || !response.body) return;

      setStreamConnected(true);
      const reader = response.body.getReader();
      const decoder = new TextDecoder();

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        const text = decoder.decode(value);
        for (const line of text.split('\n')) {
          if (line.startsWith('data: ')) {
            try {
              setStreamData(JSON.parse(line.slice(6)));
            } catch {}
          }
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') setStreamConnected(false);
    }
  };

  const handleExport = async () => {
    const apiKey = getApiKey();
    if (!apiKey) { toast.error('No API key found'); return; }
    setExportLoading(true);
    try {
      const response = await fetch(`${GAME_API_BASE}/games/export`, {
        headers: { 'X-API-Key': apiKey },
      });
      if (response.status === 403) { setExportLocked(true); return; }
      if (!response.ok) { toast.error('Export failed'); return; }
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'games_export.json';
      a.click();
      URL.revokeObjectURL(url);
      toast.success('Export downloaded');
    } catch {
      toast.error('Export failed');
    } finally {
      setExportLoading(false);
    }
  };

  const spinner = (
    <div className="flex h-48 items-center justify-center">
      <div className="w-7 h-7 rounded-full border-2 border-primary border-t-foreground animate-spin" />
    </div>
  );

  return (
    <DashboardLayout title="Analytics">
      <div className="space-y-6">

        {/* Genre Analytics */}
        <SectionShell title="Genre Analytics" icon={<BarChart2 className="w-5 h-5" />}>
          {loading ? spinner : genreLocked ? <LockedCard plan="Platinum or Enterprise" /> : (
            <ResponsiveContainer width="100%" height={260}>
              <BarChart data={genreData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                <XAxis dataKey="genre" axisLine={false} tickLine={false} tick={{ fill: '#9ca3af', fontSize: 12 }} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#9ca3af', fontSize: 12 }}
                  tickFormatter={(v) => v >= 1_000_000 ? `${(v / 1_000_000).toFixed(1)}M` : v.toLocaleString()} />
                <Tooltip
                  contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                  formatter={(value) => [(Number(value) || 0).toLocaleString(), 'Total Players']}
                />
                <Bar dataKey="total_players" fill="#111827" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </SectionShell>

        {/* Revenue Analytics */}
        <SectionShell title="Revenue Analytics" icon={<DollarSign className="w-5 h-5" />}>
          {loading ? spinner : revenueLocked ? <LockedCard plan="Enterprise" /> : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border text-left">
                    <th className="pb-3 font-medium text-accent">#</th>
                    <th className="pb-3 font-medium text-accent">Game</th>
                    <th className="pb-3 font-medium text-accent">Genre</th>
                    <th className="pb-3 font-medium text-accent">Region</th>
                    <th className="pb-3 font-medium text-accent text-right">Revenue</th>
                  </tr>
                </thead>
                <tbody>
                  {revenueData.map((r, i) => (
                    <tr key={r.id} className="border-b border-border last:border-0">
                      <td className="py-3 text-accent">{i + 1}</td>
                      <td className="py-3 font-medium text-foreground">{r.name}</td>
                      <td className="py-3 text-accent">{r.genre}</td>
                      <td className="py-3 text-accent">{r.region}</td>
                      <td className="py-3 font-semibold text-foreground text-right">฿{r.revenue.toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </SectionShell>

        {/* Region Breakdown */}
        <SectionShell title="Region Breakdown" icon={<Globe className="w-5 h-5" />}>
          {loading ? spinner : regionLocked ? <LockedCard plan="Enterprise" /> : (
            <ResponsiveContainer width="100%" height={260}>
              <BarChart data={regionData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                <XAxis dataKey="region" axisLine={false} tickLine={false} tick={{ fill: '#9ca3af', fontSize: 12 }} />
                <YAxis axisLine={false} tickLine={false} tick={{ fill: '#9ca3af', fontSize: 12 }}
                  tickFormatter={(v) => v >= 1_000_000 ? `${(v / 1_000_000).toFixed(1)}M` : v.toLocaleString()} />
                <Tooltip
                  contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                  formatter={(value) => [(Number(value) || 0).toLocaleString(), 'Total Players']}
                />
                <Bar dataKey="total_players" fill="#111827" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          )}
        </SectionShell>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Bulk Export */}
          <SectionShell title="Bulk Export" icon={<Download className="w-5 h-5" />}>
            {exportLocked ? <LockedCard plan="Enterprise" /> : (
              <div className="flex flex-col items-center justify-center py-8 gap-4">
                <p className="text-sm text-accent text-center">
                  Download all game data as a JSON file for offline analysis.
                </p>
                <button
                  onClick={handleExport}
                  disabled={exportLoading}
                  className="px-5 py-2.5 bg-foreground text-background text-sm font-semibold rounded-xl hover:bg-foreground/80 transition-colors disabled:opacity-50"
                >
                  {exportLoading ? 'Preparing...' : 'Download games_export.json'}
                </button>
              </div>
            )}
          </SectionShell>

          {/* Realtime Stream */}
          <SectionShell
            title="Realtime Stream"
            icon={
              <div className="flex items-center gap-1.5">
                <Radio className="w-5 h-5" />
                {streamConnected && (
                  <span className="relative flex h-2 w-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500" />
                  </span>
                )}
              </div>
            }
          >
            {streamLocked ? <LockedCard plan="Enterprise" /> : streamData.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-10 gap-2">
                <div className="w-6 h-6 rounded-full border-2 border-primary border-t-foreground animate-spin" />
                <p className="text-xs text-accent">Connecting to stream...</p>
              </div>
            ) : (
              <div className="space-y-2">
                {streamData.map((g) => (
                  <div key={g.id} className="flex items-center justify-between p-3 bg-secondary rounded-xl">
                    <div>
                      <p className="text-sm font-medium text-foreground">{g.name}</p>
                      <p className="text-xs text-accent">{g.genre} · {g.region}</p>
                    </div>
                    <div className="text-right">
                      <div className="flex items-center gap-1 justify-end">
                        <Users className="w-3 h-3 text-accent" />
                        <span className="text-sm font-semibold">{g.current_players.toLocaleString()}</span>
                      </div>
                      <p className="text-xs text-accent">live</p>
                    </div>
                  </div>
                ))}
                <p className="text-xs text-accent text-center pt-1">Updates every 30 minutes</p>
              </div>
            )}
          </SectionShell>
        </div>

      </div>
    </DashboardLayout>
  );
};
