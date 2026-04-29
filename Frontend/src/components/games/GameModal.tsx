import React, { useState, useEffect } from 'react';
import type { Game } from '../../types';
import { X, Upload, Plus, Save } from 'lucide-react';

interface GameModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (formData: FormData) => Promise<void>;
  game?: Game | null;
}

export const GameModal: React.FC<GameModalProps> = ({ isOpen, onClose, onSave, game }) => {
  const [formData, setFormData] = useState({
    name: '',
    genre: '',
    platform: '',
    total_players: 0,
    current_players: 0,
    revenue: 0,
    developer: '',
    publisher: '',
    region: 'Global'
  });
  const [image, setImage] = useState<File | null>(null);
  const [preview, setPreview] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (game) {
      setFormData({
        name: game.name,
        genre: game.genre,
        platform: game.platform,
        total_players: game.total_players,
        current_players: game.current_players,
        revenue: game.revenue,
        developer: game.developer,
        publisher: game.publisher,
        region: game.region
      });
      setPreview(game.image_url.startsWith('http') ? game.image_url : `http://localhost:8083${game.image_url.startsWith('/') ? '' : '/'}${game.image_url}`);
    } else {
      setFormData({
        name: '',
        genre: '',
        platform: '',
        total_players: 0,
        current_players: 0,
        revenue: 0,
        developer: '',
        publisher: '',
        region: 'Global'
      });
      setPreview('');
      setImage(null);
    }
  }, [game, isOpen]);

  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setImage(file);
      setPreview(URL.createObjectURL(file));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    const data = new FormData();
    Object.entries(formData).forEach(([key, value]) => {
      data.append(key, value.toString());
    });
    if (image) {
      data.append('image', image);
    }
    
    try {
      await onSave(data);
      onClose();
    } catch (error) {
      console.error('Save failed', error);
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-white w-full max-w-2xl rounded-3xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200">
        <div className="flex items-center justify-between px-8 py-6 border-b border-border">
          <h3 className="text-xl font-bold text-foreground">
            {game ? 'Edit Game' : 'Add New Game'}
          </h3>
          <button onClick={onClose} className="p-2 hover:bg-secondary rounded-full transition-colors">
            <X className="w-5 h-5 text-accent" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
          {/* Image Upload */}
          <div className="flex flex-col items-center justify-center border-2 border-dashed border-border rounded-2xl p-8 bg-gray-50/50 hover:bg-gray-50 transition-colors group relative cursor-pointer">
            <input 
              type="file" 
              className="absolute inset-0 opacity-0 cursor-pointer" 
              onChange={handleImageChange}
              accept="image/*"
            />
            {preview ? (
              <img src={preview} alt="Preview" className="w-32 h-32 rounded-xl object-cover shadow-md" />
            ) : (
              <div className="flex flex-col items-center gap-2">
                <div className="p-3 bg-white rounded-full shadow-sm group-hover:scale-110 transition-transform">
                  <Upload className="w-6 h-6 text-primary" />
                </div>
                <p className="text-sm font-medium text-accent">Upload Game Image</p>
                <p className="text-xs text-accent/60">JPG, PNG or WEBP (Max 5MB)</p>
              </div>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Game Name</label>
              <input 
                required
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.name}
                onChange={e => setFormData({...formData, name: e.target.value})}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Genre</label>
              <input 
                required
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.genre}
                onChange={e => setFormData({...formData, genre: e.target.value})}
                placeholder="e.g. RPG, FPS"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Platform</label>
              <input 
                required
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.platform}
                onChange={e => setFormData({...formData, platform: e.target.value})}
                placeholder="e.g. PC, PS5, Mobile"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Region</label>
              <select 
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all appearance-none"
                value={formData.region}
                onChange={e => setFormData({...formData, region: e.target.value})}
              >
                <option value="Global">Global</option>
                <option value="North America">North America</option>
                <option value="Europe">Europe</option>
                <option value="Asia">Asia</option>
              </select>
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Total Players</label>
              <input 
                type="number"
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.total_players}
                onChange={e => setFormData({...formData, total_players: parseInt(e.target.value) || 0})}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Current Players</label>
              <input 
                type="number"
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.current_players}
                onChange={e => setFormData({...formData, current_players: parseInt(e.target.value) || 0})}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Revenue ($)</label>
              <input 
                type="number"
                step="0.01"
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.revenue}
                onChange={e => setFormData({...formData, revenue: parseFloat(e.target.value) || 0})}
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-bold text-accent uppercase tracking-wider">Developer</label>
              <input 
                required
                className="w-full px-4 py-3 bg-secondary/50 border border-border rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary transition-all"
                value={formData.developer}
                onChange={e => setFormData({...formData, developer: e.target.value})}
              />
            </div>
          </div>
        </form>

        <div className="px-8 py-6 bg-gray-50/50 border-t border-border flex items-center justify-end gap-4">
          <button 
            type="button"
            onClick={onClose}
            className="px-6 py-2.5 text-sm font-semibold text-accent hover:text-foreground transition-colors"
          >
            Cancel
          </button>
          <button 
            onClick={handleSubmit}
            disabled={loading}
            className="flex items-center gap-2 px-8 py-2.5 bg-foreground text-background rounded-xl font-bold hover:bg-foreground/90 transition-all disabled:opacity-50 active:scale-95"
          >
            {loading ? (
              <div className="w-5 h-5 border-2 border-background/20 border-t-background rounded-full animate-spin"></div>
            ) : game ? (
              <>
                <Save className="w-4 h-4" />
                Save Changes
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Create Game
              </>
            )}
          </button>
        </div>
      </div>
      
      <style>{`
        .custom-scrollbar::-webkit-scrollbar { width: 6px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: #e5e7eb; border-radius: 10px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: #d1d5db; }
      `}</style>
    </div>
  );
};
