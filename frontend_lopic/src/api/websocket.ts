import type { WebSocketMessage, UploadStartMessage, UploadProgressMessage, 
  UploadErrorMessage, UploadCompleteMessage, UploadProcessingStartMessage,
  UploadProcessingErrorMessage, UploadProcessingCompleteMessage, DeleteSuccessMessage, DeleteErrorMessage, DeleteUserSuccessMessage, DeleteUserErrorMessage
 } from '../types/api';
import { serverUrl, isSeparation } from './axios';


// WebSocket 服务类
export class UploadWebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private token: string;
  private listeners: {
    start: ((data: UploadStartMessage['payload']) => void)[];
    progress: ((data: UploadProgressMessage['payload']) => void)[];
    error: ((data: UploadErrorMessage['payload']) => void)[];
    complete: ((data: UploadCompleteMessage['payload']) => void)[];
    processingStart: ((data: UploadProcessingStartMessage['payload']) => void)[];
    processingError: ((data: UploadProcessingErrorMessage['payload']) => void)[];
    processingComplete: ((data: UploadProcessingCompleteMessage['payload']) => void)[];
    deleteSuccess: ((data: DeleteSuccessMessage['payload']) => void)[];
    deleteError: ((data: DeleteErrorMessage['payload']) => void)[];
    deleteUserSuccess: ((data: DeleteUserSuccessMessage['payload']) => void)[];
    deleteUserError: ((data: DeleteUserErrorMessage['payload']) => void)[];
    open: (() => void)[];
    close: (() => void)[];
    wsError: ((error: Event) => void)[];
  };

  constructor() {
    this.token = localStorage.getItem('access_token') || '';
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.url = `${protocol}//${isSeparation ? serverUrl.split('://')[1] : window.location.host}/ws/upload?token=Bearer ${this.token}`;
    this.listeners = {
      start: [],
      progress: [],
      complete: [],
      error: [],
      processingStart: [],
      processingError: [],
      processingComplete: [],
      deleteSuccess: [],
      deleteError: [],
      deleteUserSuccess: [],
      deleteUserError: [],
      open: [],
      close: [],
      wsError: [],
    };
  }

  // 连接 WebSocket
  connect(): void {
    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.listeners.open.forEach(callback => callback());
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onclose = () => {
        this.listeners.close.forEach(callback => callback());
      };

      this.ws.onerror = (error) => {
        this.listeners.wsError.forEach(callback => callback(error));
      };
    } catch (error) {
      console.error('Error connecting to WebSocket:', error);
    }
  }

  // 关闭 WebSocket 连接
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  // 处理 WebSocket 消息
  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'upload_start':
        this.listeners.start.forEach(callback => callback(message.payload));
        break;
      case 'upload_progress':
        // 区分总进度和文件进度
        // const progressPayload = message.payload as any;
        // if (progressPayload.upload_id === 'total') {
        //   console.log('upload_progress (total):', progressPayload.progress + '%');
        // } else {
        //   console.log('upload_progress (file):', message.payload);
        // }
        this.listeners.progress.forEach(callback => callback(message.payload));
        break;
      case 'upload_error':
        this.listeners.error.forEach(callback => callback(message.payload));
        break;
      case 'upload_complete':
        this.listeners.complete.forEach(callback => callback(message.payload));
        break;
      case 'upload_processing_start':
        this.listeners.processingStart.forEach(callback => callback(message.payload));
        break;
      case 'upload_processing_error':
        this.listeners.processingError.forEach(callback => callback(message.payload));
        break;
      case 'upload_processing_complete':
        this.listeners.processingComplete.forEach(callback => callback(message.payload));
        break;
      case 'delete_success':
        this.listeners.deleteSuccess.forEach(callback => callback(message.payload));
        break;
      case 'delete_exist_error':
        this.listeners.deleteError.forEach(callback => callback(message.payload));
        break;
      case 'delete_user_success':
        this.listeners.deleteUserSuccess.forEach(callback => callback(message.payload));
        break;
      case 'delete_user_error':
        this.listeners.deleteUserError.forEach(callback => callback(message.payload));
        break;
      default:
        console.log('Unknown message type:', message);
    }
  }

  // 注册监听器
  on(event: 'start', callback: (data: UploadStartMessage['payload']) => void): void;
  on(event: 'progress', callback: (data: UploadProgressMessage['payload']) => void): void;
  on(event: 'complete', callback: (data: UploadCompleteMessage['payload']) => void): void;
  on(event: 'error', callback: (data: UploadErrorMessage['payload']) => void): void;
  on(event: 'processingStart', callback: (data: UploadProcessingStartMessage['payload']) => void): void;
  on(event: 'processingError', callback: (data: UploadProcessingErrorMessage['payload']) => void): void;
  on(event: 'processingComplete', callback: (data: UploadProcessingCompleteMessage['payload']) => void): void;
  on(event: 'deleteSuccess', callback: (data: DeleteSuccessMessage['payload']) => void): void;
  on(event: 'deleteError', callback: (data: DeleteErrorMessage['payload']) => void): void;
  on(event: 'deleteUserSuccess', callback: (data: DeleteUserSuccessMessage['payload']) => void): void;
  on(event: 'deleteUserError', callback: (data: DeleteUserErrorMessage['payload']) => void): void;
  on(event: 'wsError', callback: (error: Event) => void): void;
  on(event: 'open', callback: () => void): void;
  on(event: 'close', callback: () => void): void;
  on(event: string, callback: Function): void {
    if (event in this.listeners) {
      this.listeners[event as keyof typeof this.listeners].push(callback as any);
    }
  }

  // 移除监听器
  off(event: string, callback: Function): void {
    if (event in this.listeners) {
      const index = this.listeners[event as keyof typeof this.listeners].indexOf(callback as any);
      if (index > -1) {
        this.listeners[event as keyof typeof this.listeners].splice(index, 1);
      }
    }
  }

  // 检查连接状态
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// 导出单例实例
export const uploadWebSocketService = new UploadWebSocketService();
