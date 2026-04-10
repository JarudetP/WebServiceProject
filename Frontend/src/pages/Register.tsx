import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { authService } from '../services/auth.service';
import { useNavigate, Link } from 'react-router-dom';
import toast from 'react-hot-toast';

export const Register: React.FC = () => {
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    password: '',
    full_name: '',
    company: ''
  });
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await authService.register(formData);
      toast.success('Registration successful! Please login.');
      navigate('/login');
    } catch (err) {
      toast.error('Registration failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData(prev => ({ ...prev, [e.target.id]: e.target.value }));
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-background text-foreground py-12 px-4 sm:px-6 lg:px-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
        className="w-full max-w-md p-8 bg-white border border-border shadow-sm rounded-2xl"
      >
        <div className="mb-8 text-center">
          <h1 className="text-2xl font-semibold tracking-tight">Create an account</h1>
          <p className="text-sm text-accent mt-2">Sign up to get started</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1.5" htmlFor="username">Username</label>
            <input id="username" required className="w-full px-4 py-2.5 bg-primary border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all" value={formData.username} onChange={handleChange} />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1.5" htmlFor="email">Email</label>
            <input id="email" type="email" required className="w-full px-4 py-2.5 bg-primary border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all" value={formData.email} onChange={handleChange} />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1.5" htmlFor="password">Password</label>
            <input id="password" type="password" required minLength={6} className="w-full px-4 py-2.5 bg-primary border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all" value={formData.password} onChange={handleChange} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-1.5" htmlFor="full_name">Full Name</label>
              <input id="full_name" className="w-full px-4 py-2.5 bg-primary border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all" value={formData.full_name} onChange={handleChange} />
            </div>
            <div>
              <label className="block text-sm font-medium mb-1.5" htmlFor="company">Company</label>
              <input id="company" className="w-full px-4 py-2.5 bg-primary border border-border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-foreground transition-all" value={formData.company} onChange={handleChange} />
            </div>
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full py-2.5 px-4 mt-4 bg-foreground text-background hover:bg-foreground/90 disabled:opacity-50 rounded-lg text-sm font-medium transition-colors"
          >
            {loading ? 'Creating...' : 'Register'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-accent">
          Already have an account?{' '}
          <Link to="/login" className="text-foreground font-medium hover:underline">
            Sign In
          </Link>
        </p>
      </motion.div>
    </div>
  );
};
