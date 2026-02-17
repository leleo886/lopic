import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import type { DataResponse, ErrorResponse } from '../types/api';

export const isSeparation = import.meta.env.VITE_IS_SEPARATION === 'true';
export const serverUrl = import.meta.env.VITE_SERVER_URL;
export const projectName = import.meta.env.VITE_PROJECT_NAME;

// 创建Axios实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: isSeparation ? serverUrl : "",
  headers: {
    'Content-Type': 'application/json',
  },
});

// 检查令牌是否即将过期
const isTokenExpiringSoon = (): boolean => {
  const expiryStr = localStorage.getItem('access_token_expiry');
  if (!expiryStr) return false;
  
  const expiry = parseInt(expiryStr, 10);
  const now = Date.now();
  const bufferTime = 5 * 60 * 1000; // 5分钟缓冲时间
  
  return expiry - now < bufferTime;
};

// 提前刷新令牌
const proactiveRefreshToken = async () => {
  if (!isRefreshing && isTokenExpiringSoon()) {
    try {
      await refreshToken();
      console.log('Token refreshed proactively');
    } catch (error) {
      console.error('Proactive token refresh failed:', error);
    }
  }
};

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 从localStorage获取token
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
      // 提前检查并刷新即将过期的令牌
      proactiveRefreshToken();
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 令牌刷新状态
let isRefreshing = false;
// 存储等待刷新的请求队列
let refreshSubscribers: ((token: string) => void)[] = [];

// 刷新令牌
const refreshToken = async () => {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }
  
  try {
    const response = await axios.post(`${serverUrl}api/auth/refresh`, {
      refresh_token: refreshToken
    });
    
    if (!response.data || !response.data.data) {
      throw new Error('Invalid refresh token response');
    }
    
    const { access_token, refresh_token, expires_in, refresh_expires_in } = response.data.data;
    
    if (!access_token || !refresh_token) {
      throw new Error('Missing token in response');
    }
    
    // 存储新的令牌和过期时间信息
    localStorage.setItem('access_token', access_token);
    localStorage.setItem('refresh_token', refresh_token);
    
    // 可选：存储令牌过期时间，用于提前刷新
    if (expires_in) {
      const accessTokenExpiry = Date.now() + (expires_in * 1000);
      localStorage.setItem('access_token_expiry', accessTokenExpiry.toString());
    }
    
    if (refresh_expires_in) {
      const refreshTokenExpiry = Date.now() + (refresh_expires_in * 1000);
      localStorage.setItem('refresh_token_expiry', refreshTokenExpiry.toString());
    }
    
    return access_token;
  } catch (error) {
    console.error('Token refresh failed:', error);
    
    // 刷新失败，清除所有令牌和相关信息
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('access_token_expiry');
    localStorage.removeItem('refresh_token_expiry');
    localStorage.removeItem('user');
    
    // 跳转到登录页面
    window.location.href = '/login';
    throw error;
  }
};

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    return response;
  },
  async (error) => {
    // 处理错误响应
    if (error.response) {
      const errorData = error.response.data as ErrorResponse;
      console.error('API Error:', errorData.message);
      
      // 处理401错误（token过期）
      if (error.response.status === 401) {
        const originalRequest = error.config;
        
        // 确保原始请求存在且未被标记为刷新请求，避免无限循环
        if (!originalRequest || originalRequest._retry) {
          return Promise.reject(error);
        }
        
        // 标记请求为已重试，避免无限循环
        originalRequest._retry = true;
        
        if (!isRefreshing) {
          isRefreshing = true;
          
          try {
            // 刷新令牌
            const newToken = await refreshToken();
            
            // 通知所有等待的请求
            refreshSubscribers.forEach(callback => {
              try {
                callback(newToken);
              } catch (callbackError) {
                console.error('Error in refresh subscriber:', callbackError);
              }
            });
            refreshSubscribers = [];
            
            // 重试原始请求
            originalRequest.headers.Authorization = `Bearer ${newToken}`;
            return axiosInstance(originalRequest);
          } catch (refreshError) {
            // 刷新失败，拒绝所有等待的请求
            refreshSubscribers.forEach(callback => {
              try {
                callback('');
              } catch (callbackError) {
                console.error('Error in refresh subscriber during failure:', callbackError);
              }
            });
            refreshSubscribers = [];
            return Promise.reject(refreshError);
          } finally {
            isRefreshing = false;
          }
        } else {
          // 等待令牌刷新完成后重试
          return new Promise((resolve, reject) => {
            refreshSubscribers.push((token: string) => {
              if (token) {
                originalRequest.headers.Authorization = `Bearer ${token}`;
                resolve(axiosInstance(originalRequest));
              } else {
                reject(new Error('Token refresh failed'));
              }
            });
          });
        }
      }
    } else if (error.request) {
      console.error('Network Error:', 'No response received from server');
    } else {
      console.error('Request Error:', error.message);
    }
    return Promise.reject(error);
  }
);

// 封装请求方法
export const api = {
  get: <T>(url: string, config?: AxiosRequestConfig) => {
    return axiosInstance.get<DataResponse<T> | T>(url, config).then((res) => {
      // 检查响应是否有data字段，如果没有，直接返回响应数据
      if (typeof res.data === 'object' && res.data !== null && 'data' in res.data && res.data.data !== undefined) {
        return res.data as DataResponse<T>;
      } else {
        // 如果没有data字段，包装成DataResponse格式
        return {
          message: 'Success',
          data: res.data as T
        } as DataResponse<T>;
      }
    });
  },
  post: <T>(url: string, data?: any, config?: AxiosRequestConfig) => {
    return axiosInstance.post<DataResponse<T> | T>(url, data, config).then((res) => {
      if (typeof res.data === 'object' && res.data !== null && 'data' in res.data && res.data.data !== undefined) {
        return res.data as DataResponse<T>;
      } else {
        return {
          message: 'Success',
          data: res.data as T
        } as DataResponse<T>;
      }
    });
  },
  put: <T>(url: string, data?: any, config?: AxiosRequestConfig) => {
    return axiosInstance.put<DataResponse<T> | T>(url, data, config).then((res) => {
      if (typeof res.data === 'object' && res.data !== null && 'data' in res.data && res.data.data !== undefined) {
        return res.data as DataResponse<T>;
      } else {
        return {
          message: 'Success',
          data: res.data as T
        } as DataResponse<T>;
      }
    });
  },
  delete: <T>(url: string, config?: AxiosRequestConfig) => {
    return axiosInstance.delete<DataResponse<T> | T>(url, config).then((res) => {
      if (typeof res.data === 'object' && res.data !== null && 'data' in res.data && res.data.data !== undefined) {
        return res.data as DataResponse<T>;
      } else {
        return {
          message: 'Success',
          data: res.data as T
        } as DataResponse<T>;
      }
    });
  },
};

export default axiosInstance;
