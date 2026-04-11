import React, { useEffect, useState } from 'react';
import { DashboardLayout } from '../components/layout/DashboardLayout';
import { authService } from '../services/auth.service';
import type { Package, Subscription } from '../services/package.service';
import { packageService } from '../services/package.service';
import type { User, ApiKey } from '../types';
import toast from 'react-hot-toast';
import { Key, Package as PkgIcon, Wallet, Copy, Check, User as UserIcon, Trash2, BarChart as ChartIcon } from 'lucide-react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export const Profile: React.FC = () => {
  const [user, setUser] = useState<User | null>(null);
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [packages, setPackages] = useState<Package[]>([]);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [usageData, setUsageData] = useState<{ date: string, count: number }[]>([]);
  const [loading, setLoading] = useState(true);
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const currentUser = authService.getCurrentUser();
      if (!currentUser) return;
      const userId = currentUser.user_id;
      
      const [usr, apis, pkgs, sub, usage] = await Promise.all([
        authService.getProfile(userId),
        authService.getKeys(userId),
        packageService.getPackages(),
        packageService.getActiveSubscription(userId),
        packageService.getUsageStats()
      ]);
      
      setUser(usr && !(usr as any).error ? usr : null);
      setKeys(Array.isArray(apis) ? apis : []);
      setPackages(Array.isArray(pkgs) ? pkgs : []);
      setSubscription(sub && !(sub as any).error ? sub : null);
      setUsageData(Array.isArray(usage) ? usage : []);
      
      // Auto-set the first active api key as global default for games easily if not exist
      const validApis = Array.isArray(apis) ? apis : [];
      if (validApis.length > 0 && !localStorage.getItem('api_key')) {
         // Custom logic placeholder
      }
    } catch (err) {
      toast.error('Failed to load profile data');
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateKey = async () => {
    if (!user) return;
    try {
      const result = await authService.generateKey(user.id);
      // Backend returns the string. Show it to user.
      toast.success('API Key generated successfully');
      // Save visually
      setKeys([...keys, { id: Date.now(), key_hash: "Newly generated... (refresh to see status)", user_id: user.id, created_at: new Date().toISOString(), is_active: true } as any]);
      
      // For immediate dev ease, since backend gives struct containing apiKey if possible
      const newKey = (result as any).api_key || result.apiKey || (result as any).key;
      if (newKey) {
        localStorage.setItem('api_key', newKey);
        toast('New API Key has been auto-set for your Game requests!', { icon: '🔑' });
      } else {
        // Just reload
        fetchData();
      }
    } catch (err) {
      toast.error('Failed to generate key');
    }
  };

  const handlePurchase = async (pkgId: number) => {
    if (!user) return;
    try {
      await packageService.purchasePackage(user.id, pkgId);
      toast.success('Package purchased successfully!');
      fetchData(); // Reload sub & balance
    } catch (err: any) {
      toast.error(err.response?.data?.error || 'Failed to purchase package');
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedKey(text);
    setTimeout(() => setCopiedKey(null), 2000);
    toast.success('Copied to clipboard');
  };

  const handleDeleteKey = async (keyStr: string) => {
    if (!user) return;
    try {
      await authService.deleteKey(user.id, keyStr);
      toast.success('API Key deleted');
      fetchData(); // Reload keys
      if (localStorage.getItem('api_key') === keyStr) {
        localStorage.removeItem('api_key');
      }
    } catch (err) {
      toast.error('Failed to delete key');
    }
  };

  if (loading) {
    return (
      <DashboardLayout title="Profile & Settings">
        <div className="flex h-64 items-center justify-center">
            <div className="w-8 h-8 rounded-full border-2 border-primary border-t-foreground animate-spin"></div>
        </div>
      </DashboardLayout>
    );
  }

  return (
    <DashboardLayout title="Profile & Settings">
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
        
        {/* Left Column: Profile & API Keys */}
        <div className="xl:col-span-1 space-y-6">
          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
            <div className="flex items-center gap-4 mb-6">
              <div className="w-16 h-16 bg-secondary text-foreground flex items-center justify-center rounded-2xl">
                <UserIcon className="w-8 h-8" />
              </div>
              <div>
                <h3 className="text-xl font-semibold tracking-tight">{user?.full_name || user?.username}</h3>
                <p className="text-sm text-accent">{user?.email}</p>
              </div>
            </div>
            
            <div className="space-y-4">
              <div className="flex justify-between items-center py-2 border-b border-border">
                <span className="text-sm text-accent">Company</span>
                <span className="text-sm font-medium">{user?.company || '-'}</span>
              </div>
              <div className="flex justify-between items-center py-2 border-b border-border">
                <span className="text-sm text-accent">Role</span>
                <span className="text-sm font-medium capitalize">{user?.role}</span>
              </div>
              <div className="flex justify-between items-center py-2 pt-4">
                <div className="flex items-center gap-2">
                  <Wallet className="w-4 h-4 text-accent" />
                  <span className="text-sm text-accent">Balance</span>
                </div>
                <span className="text-lg font-semibold tracking-tight">${user?.balance?.toLocaleString()}</span>
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-2">
                <Key className="w-5 h-5 text-accent" />
                <h3 className="text-lg font-medium tracking-tight">API Keys</h3>
              </div>
              <button 
                onClick={handleGenerateKey}
                className="px-3 py-1.5 bg-foreground text-background text-xs font-medium rounded-lg hover:bg-foreground/90 transition-colors"
              >
                Generate New
              </button>
            </div>
            
            <div className="space-y-3">
              {!keys || keys.length === 0 ? (
                <p className="text-sm text-accent">No API keys created yet.</p>
              ) : (
                keys.map((k, index) => {
                  const keyStr = typeof k === 'string' ? k : (k as any).key_hash || (k as any).key;
                  // If it's freshly generated, we hacked it as an object with key_hash.
                  const isActive = typeof k === 'string' ? true : (k as any).is_active;

                  return (
                    <div key={index} className="p-3 bg-secondary rounded-xl flex items-center justify-between group">
                      <div className="overflow-hidden pr-4 flex-1">
                        <p className="text-sm font-mono text-foreground truncate w-full">
                          {keyStr?.substring(0, 20)}...
                        </p>
                        <p className="text-xs text-accent mt-1">
                          Status: <span className={isActive ? 'text-green-600 font-medium' : 'text-red-500'}>{isActive ? 'Active' : 'Revoked'}</span>
                        </p>
                      </div>
                      <div className="flex items-center gap-1">
                        <button 
                          onClick={() => copyToClipboard(keyStr)}
                          className="p-1.5 hover:bg-white rounded-lg transition-colors text-accent hover:text-foreground"
                          title="Copy Key"
                        >
                          <Copy className="w-4 h-4" />
                        </button>
                        <button 
                          onClick={() => handleDeleteKey(keyStr)}
                          className="p-1.5 hover:bg-red-50 rounded-lg transition-colors text-accent hover:text-red-600"
                          title="Delete Key"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                      {copiedKey === keyStr && <span className="absolute -mt-8 right-16 text-xs bg-black text-white px-2 py-1 rounded">Copied!</span>}
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Packages */}
        <div className="xl:col-span-2 space-y-6">
          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
            <div className="flex items-center gap-2 mb-6">
              <PkgIcon className="w-5 h-5 text-accent" />
              <h3 className="text-lg font-medium tracking-tight">Current Subscription</h3>
            </div>
            
            {subscription ? (
               <div className="p-4 bg-primary rounded-xl border border-border flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-foreground">
                      Package ID: {subscription.package_id}
                    </p>
                    <p className="text-xs text-accent mt-1">
                      Expires: {new Date(subscription.expires_at).toLocaleDateString()}
                    </p>
                  </div>
                  <span className={`px-3 py-1 text-xs font-semibold rounded-full ${subscription.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                    {subscription.status === 'active' ? 'Active' : subscription.status}
                  </span>
               </div>
            ) : (
                <div className="p-4 bg-secondary rounded-xl text-center">
                    <p className="text-sm text-accent">You don't have an active subscription.</p>
                </div>
            )}
          </div>

          {/* Usage Chart */}
          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
            <div className="flex items-center gap-2 mb-6">
              <ChartIcon className="w-5 h-5 text-accent" />
              <h3 className="text-lg font-medium tracking-tight">API Request History (30d)</h3>
            </div>
            <div className="h-64 w-full">
               {usageData.length > 0 ? (
                 <ResponsiveContainer width="100%" height="100%">
                   <BarChart data={usageData}>
                     <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f3f4f6" />
                     <XAxis 
                        dataKey="date" 
                        axisLine={false} 
                        tickLine={false} 
                        tick={{ fill: '#9ca3af', fontSize: 10 }}
                        tickFormatter={(val) => val.split('-').slice(1).join('/')}
                        minTickGap={30}
                     />
                     <YAxis 
                        axisLine={false} 
                        tickLine={false} 
                        tick={{ fill: '#9ca3af', fontSize: 10 }}
                     />
                     <Tooltip 
                        contentStyle={{ borderRadius: '12px', border: '1px solid #e5e7eb', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }}
                        labelStyle={{ color: '#111827', fontWeight: 600 }}
                     />
                     <Bar dataKey="count" fill="#111827" radius={[4, 4, 0, 0]} />
                   </BarChart>
                 </ResponsiveContainer>
               ) : (
                 <div className="h-full flex items-center justify-center border-2 border-dashed border-border rounded-xl">
                   <p className="text-sm text-accent">No usage data found for this period.</p>
                 </div>
               )}
            </div>
          </div>

          <div className="bg-white p-6 rounded-2xl border border-border shadow-sm">
            <h3 className="text-lg font-medium tracking-tight mb-6">Available Packages</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {packages?.map(pkg => (
                <div key={pkg.id} className="p-5 border border-border rounded-xl hover:border-foreground transition-colors flex flex-col justify-between">
                  <div>
                    <h4 className="font-semibold text-foreground text-lg">{pkg.name}</h4>
                    <p className="text-sm text-accent mt-1 mb-4 flex-grow">{pkg.description}</p>
                    
                    <ul className="space-y-2 mb-6">
                       <li className="flex items-center gap-2 text-sm text-foreground">
                         <Check className="w-4 h-4 text-accent" />
                         {pkg.request_limit === -1 ? 'Unlimited requests' : `${pkg.request_limit} reqs / ${pkg.refresh_interval_minutes}m`}
                       </li>
                       <li className="flex items-center gap-2 text-sm text-foreground">
                         <Check className="w-4 h-4 text-accent" />
                         {pkg.historical_data_days} Days historical data
                       </li>
                    </ul>
                  </div>
                  
                  <div className="flex items-center justify-between pt-4 border-t border-border">
                    <span className="text-2xl font-semibold tracking-tight">${pkg.price}</span>
                    <button 
                      onClick={() => handlePurchase(pkg.id)}
                      className="px-4 py-2 bg-secondary text-foreground hover:bg-foreground hover:text-background text-sm font-medium rounded-lg transition-colors"
                    >
                      Purchase
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

      </div>
    </DashboardLayout>
  );
};
