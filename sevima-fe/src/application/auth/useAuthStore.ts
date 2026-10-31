import { create } from 'zustand';
import { User, UserRole } from '@/domain/models/types';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  institutionCode: string | null;
  institutionName: string | null;
  isAuthenticated: boolean;
  setAuth: (data: {
    user: User;
    accessToken: string;
    refreshToken: string;
    institutionCode?: string;
    institutionName?: string;
  }) => void;
  logout: () => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  institutionCode: null,
  institutionName: null,
  isAuthenticated: false,

  setAuth: ({ user, accessToken, refreshToken, institutionCode, institutionName }) => {
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('user_session', JSON.stringify({ user, institutionCode, institutionName }));
    set({
      user,
      accessToken,
      institutionCode: institutionCode || null,
      institutionName: institutionName || null,
      isAuthenticated: true,
    });
  },

  logout: () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_session');
    set({
      user: null,
      accessToken: null,
      institutionCode: null,
      institutionName: null,
      isAuthenticated: false,
    });
  },

  initialize: () => {
    const token = localStorage.getItem('access_token');
    const sessionStr = localStorage.getItem('user_session');
    if (token && sessionStr) {
      try {
        const { user, institutionCode, institutionName } = JSON.parse(sessionStr);
        set({
          user,
          accessToken: token,
          institutionCode: institutionCode || null,
          institutionName: institutionName || null,
          isAuthenticated: true,
        });
      } catch {
        localStorage.removeItem('access_token');
        localStorage.removeItem('user_session');
      }
    }
  },
}));
