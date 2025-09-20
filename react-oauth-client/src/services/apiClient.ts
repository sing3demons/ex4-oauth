import axios, { AxiosInstance, AxiosResponse, AxiosError } from 'axios';
import { apiConfig } from '../config';
import { StorageUtils } from '../utils/oauth';

/**
 * HTTP Client สำหรับ API calls
 * พร้อม auto token injection และ refresh mechanism
 */
class ApiClient {
  private client: AxiosInstance;
  private isRefreshing = false;
  private failedQueue: Array<{
    resolve: (value: any) => void;
    reject: (error: any) => void;
  }> = [];

  constructor() {
    this.client = axios.create({
      baseURL: apiConfig.baseURL,
      timeout: apiConfig.timeout,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  /**
   * Setup request และ response interceptors
   */
  private setupInterceptors(): void {
    // Request interceptor: เพิ่ม access token
    this.client.interceptors.request.use(
      (config) => {
        // Auto-inject Authorization header if token exists
      if (!config.headers?.Authorization) {
        const token = StorageUtils.getAccessToken();
        if (token) {
          config.headers = {
            ...config.headers,
            Authorization: `Bearer ${token}`,
          };
        }
      }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor: จัดการ 401 และ auto refresh
    this.client.interceptors.response.use(
      (response) => response,
      async (error: AxiosError) => {
        const originalRequest = error.config as any;

        if (error.response?.status === 401 && !originalRequest._retry) {
          if (this.isRefreshing) {
            // ถ้ากำลัง refresh อยู่ ให้รอ
            return new Promise((resolve, reject) => {
              this.failedQueue.push({ resolve, reject });
            }).then(token => {
              originalRequest.headers.Authorization = `Bearer ${token}`;
              return this.client(originalRequest);
            }).catch(err => {
              return Promise.reject(err);
            });
          }

          originalRequest._retry = true;
          this.isRefreshing = true;

          try {
            const newToken = await authService.refreshToken();
            if (newToken) {
              // ส่ง token ใหม่ให้ requests ที่รออยู่
              this.processQueue(newToken.access_token, null);
              
              // Retry original request
              originalRequest.headers.Authorization = `Bearer ${newToken.access_token}`;
              return this.client(originalRequest);
            } else {
              throw new Error('Failed to refresh token');
            }
          } catch (refreshError) {
            this.processQueue(null, refreshError);
            authService.logout();
            return Promise.reject(refreshError);
          } finally {
            this.isRefreshing = false;
          }
        }

        return Promise.reject(error);
      }
    );
  }

  /**
   * Process queued requests after token refresh
   */
  private processQueue(token: string | null, error: any): void {
    this.failedQueue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve(token);
      }
    });
    
    this.failedQueue = [];
  }

  /**
   * GET request
   */
  async get<T = any>(url: string, params?: any): Promise<T> {
    const response: AxiosResponse<T> = await this.client.get(url, { params });
    return response.data;
  }

  /**
   * POST request
   */
  async post<T = any>(url: string, data?: any): Promise<T> {
    const response: AxiosResponse<T> = await this.client.post(url, data);
    return response.data;
  }

  /**
   * PUT request
   */
  async put<T = any>(url: string, data?: any): Promise<T> {
    const response: AxiosResponse<T> = await this.client.put(url, data);
    return response.data;
  }

  /**
   * DELETE request
   */
  async delete<T = any>(url: string): Promise<T> {
    const response: AxiosResponse<T> = await this.client.delete(url);
    return response.data;
  }

  /**
   * PATCH request
   */
  async patch<T = any>(url: string, data?: any): Promise<T> {
    const response: AxiosResponse<T> = await this.client.patch(url, data);
    return response.data;
  }

  /**
   * Upload file
   */
  async uploadFile<T = any>(url: string, file: File, fieldName = 'file'): Promise<T> {
    const formData = new FormData();
    formData.append(fieldName, file);

    const response: AxiosResponse<T> = await this.client.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return response.data;
  }

  /**
   * Download file
   */
  async downloadFile(url: string, filename?: string): Promise<void> {
    const response = await this.client.get(url, {
      responseType: 'blob',
    });

    const blob = new Blob([response.data]);
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = downloadUrl;
    link.download = filename || 'download';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(downloadUrl);
  }

  /**
   * Set default headers
   */
  setDefaultHeader(key: string, value: string): void {
    this.client.defaults.headers.common[key] = value;
  }

  /**
   * Remove default header
   */
  removeDefaultHeader(key: string): void {
    delete this.client.defaults.headers.common[key];
  }

  /**
   * Get axios instance สำหรับ advanced usage
   */
  getAxiosInstance(): AxiosInstance {
    return this.client;
  }
}

/**
 * Error Handler Utility
 */
export class ApiErrorHandler {
  /**
   * Extract error message from API response
   */
  static getErrorMessage(error: any): string {
    if (error.response?.data?.error_description) {
      return error.response.data.error_description;
    }
    
    if (error.response?.data?.message) {
      return error.response.data.message;
    }

    if (error.response?.data?.error) {
      return error.response.data.error;
    }

    if (error.message) {
      return error.message;
    }

    return 'An unexpected error occurred';
  }

  /**
   * Check if error is network error
   */
  static isNetworkError(error: any): boolean {
    return !error.response && error.code === 'ECONNABORTED';
  }

  /**
   * Check if error is timeout
   */
  static isTimeoutError(error: any): boolean {
    return error.code === 'ECONNABORTED' && error.message.includes('timeout');
  }

  /**
   * Check if error is authentication error
   */
  static isAuthError(error: any): boolean {
    return error.response?.status === 401;
  }

  /**
   * Check if error is authorization error
   */
  static isForbiddenError(error: any): boolean {
    return error.response?.status === 403;
  }

  /**
   * Check if error is validation error
   */
  static isValidationError(error: any): boolean {
    return error.response?.status === 422 || error.response?.status === 400;
  }

  /**
   * Get validation errors from response
   */
  static getValidationErrors(error: any): Record<string, string[]> {
    if (error.response?.data?.errors) {
      return error.response.data.errors;
    }
    return {};
  }
}

// สร้าง instance เดียวสำหรับทั้งแอป
export const apiClient = new ApiClient();

// Export default เป็น instance
export default apiClient;