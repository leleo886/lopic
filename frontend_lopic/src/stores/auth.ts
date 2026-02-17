import { defineStore } from 'pinia';
import type { User } from '../types/api';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null as User | null,
    token: localStorage.getItem('access_token') || null,
    isAuthenticated: !!localStorage.getItem('access_token'),
  }),
  
  getters: {
    currentUser: (state) => state.user,
    isLoggedIn: (state) => state.isAuthenticated,
    isAdmin: (state) => state.user?.role.name === 'admin',
  },
  
  actions: {
    setAuth(user: User, token: string) {
      this.user = user;
      this.token = token;
      this.isAuthenticated = true;
      localStorage.setItem('access_token', token);
      localStorage.setItem('user', JSON.stringify(user));
    },
    
    logout() {
      this.user = null;
      this.token = null;
      this.isAuthenticated = false;
      localStorage.removeItem('access_token');
      localStorage.removeItem('user');
    },
    
    // 从localStorage恢复用户状态
    restoreState() {
      const storedUser = localStorage.getItem('user');
      if (storedUser) {
        try {
          this.user = JSON.parse(storedUser);
        } catch (error) {
          console.error('Failed to parse stored user:', error);
          this.logout();
        }
      }
    },
  },
});
