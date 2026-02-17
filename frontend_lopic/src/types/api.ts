// API响应类型
export interface SuccessResponse {
  message: string;
}

export interface DataResponse<T> {
  message: string;
  data: T;
}

export interface ErrorResponse {
  code: string;
  message: string;
}

// 认证相关
export interface LoginRequest {
  username: string;
  password: string;
}

export interface LogoutRequest {
  refresh_token: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  locale: string;
}

export interface ResetPasswordRequest {
  email: string;
  code: string;
  new_password: string;
}

export interface RequestPasswordResetRequest {
  email: string;
  locale: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  refresh_expires_in: number;
}

export interface LoginResponse {
  token_response: TokenResponse;
  user: User;
}

// 用户相关
export interface User {
  id: number;
  username: string;
  email: string;
  role: Role;
  role_id: number;
  active: boolean;
  created_at: string;
  updated_at: string;
  image_count: number;
  total_size: number;
}

export interface AdminUserRequest {
  username: string;
  email: string;
  password?: string;
  role: string;
  active?: boolean;
}

export interface GetUsersResponse {
  users: User[];
  total: number;
  page: number;
  page_size: number;
}


export interface StorageConfig {
  base_path: string;
  static_url: string;
  base_url?: string;
  username?: string;
  password?: string;
}

// 角色相关
export interface Role {
  id: number;
  name: string;
  description: string;
  allowed_extensions: string[];
  gallery_open: boolean;
  max_albums_per_user: number;
  max_file_size_mb: number;
  max_files_per_upload: number;
  max_storage_size_mb: number;
  storage_name: string;
  created_at: string;
  updated_at: string;
}

export interface RoleRequest {
  name: string;
  description: string;
  allowed_extensions: string[];
  gallery_open: boolean;
  max_albums_per_user: number;
  max_file_size_mb: number;
  max_files_per_upload: number;
  max_storage_size_mb: number;
  storage_name: string;
}

// 相册相关
export interface AlbumRequest {
  name: string;
  description?: string;
  gallery_enabled: boolean;
  serial_number?: number;
}

export interface AlbumResponse {
  id: number;
  name: string;
  description?: string;
  user_id: number;
  created_at: string;
  updated_at: string;
  cover_image?: string;
  image_count?: number;
  gallery_enabled: boolean;
  serial_number?: number;
}

export interface GetAlbumsResponse {
  albums: AlbumResponse[];
  total: number;
  page: number;
  page_size: number;
}

// 图片相册关联相关
export interface AddOrDelImageToAlbumRequset {
  album_id: number;
  ids: number[];
}

// 图片相关
export interface ImageResponse {
  id: number;
  file_name: string;
  original_name: string;
  path: string;
  file_url: string;
  file_size: number;
  height: number;
  width: number;
  thumbnail_url: string;
  thumbnail_size: number;
  thumbnail_height: number;
  thumbnail_width: number;
  mime_type: string;
  user_id: number;
  storage_name: string;
  created_at: string;
  updated_at: string;
  tags?: string[];
  albums?: AlbumResponse[];
  loaded?: boolean;
}

export interface GetImagesResponse {
  images: ImageResponse[];
  total: number;
  page: number;
  page_size: number;
}

export interface GetAlbumImagesResponse {
  images: ImageResponse[];
  total: number;
  page: number;
  page_size: number;
}

export interface AddOrDelImageToAlbumResponse {
  success_ids: Record<number, string>;
  error_ids: Record<number, string>;
}

// 标签相关
export interface TagCloudItem {
  tag: string;
  count: number;
}

// 存储配置相关
export interface Storage {
  id: number;
  name: string;
  type: string;
  config: StorageConfig;
  created_at: string;
  updated_at: string;
}

export interface StorageRequest {
  name: string;
  type: string;
  config: StorageConfig;
}

// 存储使用情况相关
export interface StorageUsage {
  total_size: number;
	image_count: number;
}

// 上传响应相关
export interface UploadResponse {
  image_responses: ImageResponse[];
}

// 更新请求相关
export interface UpdateRequest {
  ids: number[];
  tags?: string[];
  title?: string;
}

// 更新存储名称请求相关
export interface UpdateStorageRequest {
  ids: number[];
  storage_name: string;
}

// 更新响应相关
export interface UpdateResponse {
  success_ids: Record<number, ImageResponse>;
  error_ids: Record<number, string>;
}

// 普通用户请求相关
export interface UserRequest {
  username: string;
  password?: string;
}

// 系统设置相关
export interface GalleryConfig {
  Title: string;
  BackgroundImage: string;
  CustomContent: string;
}

export interface SystemSettings {
  General: {
    MaxThumbSize: number;
    RegisterEnabled: boolean;
    MaxTags: number;
  };
  Mail: {
    Enabled: boolean;
    ServerAddress: string;
    SMTP: {
      Host: string;
      Port: number;
      Username: string;
      Password: string;
      From: string;
      FromName: string;
    };
  };
  Gallery: GalleryConfig;
}

// 备份相关
export interface BackupTask {
  id: number;
  status: string;
  size: number;
  created_at: string;
  updated_at: string;
  start_time: string;
  end_time: string;
  error: string;
  storage_path: string;
}

export interface RestoreTask {
  id: number;
  backup_task: BackupTask;
  backup_task_id: number;
  status: string;
  created_at: string;
  updated_at: string;
  start_time: string;
  end_time: string;
  error: string;
}

// WebSocket 相关
export interface UploadStartMessage {
  type: 'upload_start';
  payload: {
    upload_id: string;
    file_name: string;
    file_size: number;
  };
}

export interface UploadProgressMessage {
  type: 'upload_progress';
  payload: {
    upload_id: string;
    file_name: string;
    progress: number;
    read: number;
    total: number;
  };
}

export interface UploadErrorMessage {
  type: 'upload_error';
  payload: {
    upload_id: string;
    file_name: string;
    error: string;
  };
}

export interface UploadCompleteMessage {
  type: 'upload_complete';
  payload: {
    upload_id: string;
    file_name: string;
    file_url: string;
    thumbnail_url: string;
    image_id: number;
  };
}

export interface UploadProcessingStartMessage {
  type: 'upload_processing_start';
  payload: {
    message: string;
    file_count: number;
  };
}

export interface UploadProcessingErrorMessage {
  type: 'upload_processing_error';
  payload: {
    message: string;
    error: string;
    code: string;
  };
}

export interface UploadProcessingCompleteMessage {
  type: 'upload_processing_complete';
  payload: {
    message: string;
    file_count: number;
  };
}

// 删除响应相关
export interface DeleteSuccessMessage {
  type: 'delete_success';
  payload: {
    message: string;
    file_count: number;
  };
}

export interface DeleteErrorMessage {
  type: 'delete_error';
  payload: {
    message: string;
    error: string;
    code: string;
  };
}

export interface DeleteUserSuccessMessage {
  type: 'delete_user_success';
  payload: {
    message: string;
    user_count: number;
  };
}

export interface DeleteUserErrorMessage {
  type: 'delete_user_error';
  payload: {
    message: string;
    error: string;
    code: string;
  };
}

export interface DeleteExistErrorMessage {
  type: 'delete_exist_error';
  payload: {
    message: string;
    error: string;
    code: string;
  };
}



export type WebSocketMessage = UploadStartMessage | UploadProgressMessage | UploadErrorMessage | UploadCompleteMessage | UploadProcessingStartMessage | UploadProcessingErrorMessage | UploadProcessingCompleteMessage | DeleteSuccessMessage | DeleteErrorMessage | DeleteUserSuccessMessage | DeleteUserErrorMessage | DeleteExistErrorMessage;
