// 时间格式化函数
export const formatDateTime = (dateTime?: string) => {
  if (!dateTime) return '-';
  
  const date = new Date(dateTime);
  if (isNaN(date.getTime())) return '-';
  
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
};

// 文件大小格式化函数
export const formatFileSize = (size?: number) => {
  if (!size) return '-';
  
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index++;
  }
  
  return `${size.toFixed(2)} ${units[index]}`;
};

// 文件地址返回
import {serverUrl, isSeparation} from '../api/axios';
export const getFileUrl = (url: string) => {
  if (url.startsWith('http') || url.startsWith('https')) {
    return url;
  } else if (url.startsWith('/')) {
    if (isSeparation && serverUrl) {
      return serverUrl + url;
    } else {
      return window.location.origin + url;
    }
  } else {
    return url;
  }
};
