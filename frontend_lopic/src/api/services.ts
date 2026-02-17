import { api } from './axios';
import type {
  LoginRequest,
  LogoutRequest,
  LoginResponse,
  RefreshTokenRequest,
  RegisterRequest,
  ResetPasswordRequest,
  RequestPasswordResetRequest,
  TokenResponse,
  User,
  UserRequest,
  AdminUserRequest,
  GetUsersResponse,
  Role,
  RoleRequest,
  AlbumRequest,
  AlbumResponse,
  GetAlbumsResponse,
  ImageResponse,
  GetImagesResponse,
  GetAlbumImagesResponse,
  AddOrDelImageToAlbumResponse,
  AddOrDelImageToAlbumRequset,
  TagCloudItem,
  StorageUsage,
  Storage,
  StorageRequest,
  UpdateRequest,
  UpdateResponse,
  UpdateStorageRequest,
  SystemSettings,
  GalleryConfig,
  BackupTask,
  RestoreTask,
  DataResponse,
  SuccessResponse,
} from '../types/api';

// 认证相关API
export const authApi = {
  login: (data: LoginRequest) => {
    return api.post<LoginResponse>('/api/auth/login', data);
  },
  logout: (data: LogoutRequest) => {
    return api.post('/api/auth/logout', data);
  },
  refresh: (data: RefreshTokenRequest) => {
    return api.post<TokenResponse>('/api/auth/refresh', data);
  },
  register: (data: RegisterRequest) => {
    return api.post<SuccessResponse>('/api/auth/register', data);
  },
  resetPassword: (data: ResetPasswordRequest) => {
    return api.post('/api/auth/reset-password', data);
  },
  requestPasswordReset: (data: RequestPasswordResetRequest) => {
    return api.post('/api/auth/reset-password/request', data);
  },
  verifyEmail: (token: string) => {
    return api.get('/api/auth/verify-email', { params: { token } });
  },
};

// 用户相关API
export const userApi = {
  // 获取用户列表（管理员）
  getUsers: (params: {
    page?: number;
    page_size?: number;
    searchkey?: string;
    orderby?: string;
    order?: string;
  }) => {
    return api.get<GetUsersResponse>('/api/admin/users', { params });
  },
  // 获取单个用户（管理员）
  getUser: (id: number) => {
    return api.get<User>(`/api/admin/users/${id}`);
  },
  // 创建用户（管理员）
  createUser: (data: AdminUserRequest) => {
    return api.post<User>('/api/admin/users', data);
  },
  // 更新用户（管理员）
  updateUser: (id: number, data: AdminUserRequest) => {
    return api.put<User>(`/api/admin/users/${id}`, data);
  },
  // 删除用户（管理员）
  deleteUser: (id: number) => {
    return api.delete(`/api/admin/users/${id}`);
  },
  // 获取所有用户图片标签云（管理员）
  getUsersTagsCloud: () => {
    return api.get<TagCloudItem[]>('/api/admin/users/tags-cloud');
  },
  // 获取指定用户图片标签云（管理员）
  getUserTagsCloud: (id: number) => {
    return api.get<TagCloudItem[]>(`/api/admin/users/${id}/tags-cloud`);
  },
  // 获取当前用户信息（普通用户）
  getCurrentUser: () => {
    return api.get<User>('/api/users/me');
  },
  // 更新当前用户信息（普通用户）
  updateCurrentUser: (data: UserRequest) => {
    return api.put<User>('/api/users/me', data);
  },
  // 获取当前用户存储空间使用情况（普通用户）
  getCurrentUserStorage: () => {
    return api.get<StorageUsage>('/api/users/me/storage');
  },
  // 获取当前用户图片标签云（普通用户）
  getCurrentUserTagsCloud: () => {
    return api.get<TagCloudItem[]>('/api/users/me/tags-cloud');
  },
};

// 角色相关API
export const roleApi = {
  // 获取角色列表（管理员）
  getRoles: (params: {
    page?: number;
    page_size?: number;
    searchkey?: string;
    orderby?: string;
    order?: string;
  }) => {
    return api.get<Role[]>('/api/admin/roles', { params });
  },
  // 获取角色详情（管理员）
  getRole: (id: number) => {
    return api.get<Role>(`/api/admin/roles/${id}`);
  },
  // 创建角色（管理员）
  createRole: (data: RoleRequest) => {
    return api.post<Role>('/api/admin/roles', data);
  },
  // 更新角色（管理员）
  updateRole: (id: number, data: RoleRequest) => {
    return api.put<Role>(`/api/admin/roles/${id}`, data);
  },
  // 删除角色（管理员）
  deleteRole: (id: number) => {
    return api.delete(`/api/admin/roles/${id}`);
  },
  // 获取每个角色下的用户数量（管理员）
  getRolesUsersCount: () => {
    return api.get<Record<string, number>>('/api/admin/roles/users-count');
  },
};

// 相册相关API
export const albumApi = {
  // 普通用户相册操作
  // 获取当前用户的相册列表
  getUserAlbums: (params: {
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetAlbumsResponse>('/api/albums', { params });
  },
  // 获取当前用户的相册详情
  getUserAlbum: (id: number) => {
    return api.get<AlbumResponse>(`/api/albums/${id}`);
  },
  // 创建相册
  createAlbum: (data: AlbumRequest) => {
    return api.post<AlbumResponse>('/api/albums', data);
  },
  // 更新当前用户的相册
  updateUserAlbum: (id: number, data: AlbumRequest) => {
    return api.put<AlbumResponse>(`/api/albums/${id}`, data);
  },
  // 删除当前用户的相册
  deleteUserAlbum: (id: number) => {
    return api.delete(`/api/albums/${id}`);
  },
  // 获取当前用户相册中的图片
  getUserAlbumImages: (id: number, params: {
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetAlbumImagesResponse>(`/api/albums/${id}/images`, { params });
  },
  // 获取当前用户不在任何相册中的图片
  getImagesNotInAnyAlbum: (params: {
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetAlbumImagesResponse>('/api/albums/images/not-in-any', { params });
  },
  
  // 管理员相册操作
  // 获取所有相册列表（管理员）
  getAllAlbums: (params: {
    page?: number;
    page_size?: number;
    searchkey?: string;
    orderby?: string;
    order?: string;
  }) => {
    return api.get<GetAlbumsResponse>('/api/admin/albums', { params });
  },
  // 获取指定相册详情（管理员）
  getAdminAlbum: (id: number) => {
    return api.get<AlbumResponse>(`/api/admin/albums/${id}`);
  },
  // 删除指定相册（管理员）
  deleteAdminAlbums: (ids: number[]) => {
    return api.delete<AddOrDelImageToAlbumResponse>('/api/admin/albums', {
      data: ids,
    });
  },
};

// 图片相关API
export const imageApi = {
  // 获取所有图片（管理员）
  getImages: (params: {
    page?: number;
    page_size?: number;
    searchkey?: string;
    orderby?: string;
    order?: string;
    field?: string;
    value?: string;
  }) => {
    return api.get<GetImagesResponse>('/api/admin/images', { params });
  },
  // 获取单个图片（管理员）
  getImage: (id: number) => {
    return api.get<ImageResponse>(`/api/admin/images/${id}`);
  },
  // 删除图片（管理员）
  deleteImages: (ids: number[]) => {
    return api.delete<AddOrDelImageToAlbumResponse>('/api/admin/images', {
      data: ids,
    });
  },
  // 获取当前用户的图片列表（普通用户）
  getUserImages: (params: {
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetImagesResponse>('/api/images', { params });
  },
  // 获取指定图片详情（普通用户）
  getUserImage: (id: number) => {
    return api.get<ImageResponse>(`/api/images/${id}`);
  },
  // 更新指定图片信息（普通用户）
  updateUserImage: (data: UpdateRequest) => {
    return api.put<UpdateResponse>('/api/images', data);
  },
  // 删除指定图片（普通用户）
  deleteUserImages: (ids: number[]) => {
    return api.delete<SuccessResponse>('/api/images', {
      data: ids,
    });
  },
  // 搜索图片（普通用户）
  searchImages: (params: {
    search_key?: string;
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetImagesResponse>('/api/images/search', { params });
  },
  // 上传图片（普通用户）
  uploadImages: (formData: FormData) => {
    return api.post<SuccessResponse>('/api/images/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
  // 添加图片到相册（普通用户）
  addImagesToAlbum: (data: AddOrDelImageToAlbumRequset) => {
    return api.post<AddOrDelImageToAlbumResponse>('/api/images/albums', data);
  },
  // 从相册移除图片（普通用户）
  removeImagesFromAlbum: (data: AddOrDelImageToAlbumRequset) => {
    return api.delete<AddOrDelImageToAlbumResponse>('/api/images/albums', {
      data,
    });
  },
  // 更新图片存储名称（管理员）
  updateImageStorage: (data: UpdateStorageRequest) => {
    return api.put<AddOrDelImageToAlbumResponse>('/api/admin/images/storagename', data);
  },
};

// 系统相关API
export const systemApi = {
  // 获取系统信息（管理员）
  getSystemInfo: () => {
    return api.get<SystemSettings>('/api/admin/system/info');
  },
  // 更新系统信息（管理员）
  updateSystemInfo: (data: SystemSettings) => {
    return api.put('/api/admin/system/info', data);
  },
  // 创建备份（管理员）
  createBackup: () => {
    return api.post<DataResponse<BackupTask>>('/api/admin/backup');
  },
  // 获取备份列表（管理员）
  getBackups: () => {
    return api.get<DataResponse<BackupTask[]>>('/api/admin/backup/list');
  },
  // 恢复备份（管理员）
  restoreBackup: (id: number) => {
    return api.post<DataResponse<RestoreTask>>(`/api/admin/backup/restore/${id}`);
  },
  // 删除备份（管理员）
  deleteBackup: (id: number) => {
    return api.delete<SuccessResponse>(`/api/admin/backup/${id}`);
  },
  // 删除恢复任务（管理员）
  deleteRestoreTask: (id: number) => {
    return api.delete<SuccessResponse>(`/api/admin/backup/restore/${id}`);
  },
  // 获取恢复任务列表（管理员）
  getRestoreTasks: () => {
    return api.get<DataResponse<RestoreTask[]>>('/api/admin/backup/restore/list');
  },
  // 下载备份文件（管理员）
  downloadBackup: (id: number) => {
    return api.get(`/api/admin/backup/download/${id}`, {
      responseType: 'blob',
    });
  },
  // 上传备份文件（管理员）
  uploadBackup: (formData: FormData) => {
    return api.post<DataResponse<BackupTask>>('/api/admin/backup/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },
};

// 画廊相关API
export const galleryApi = {
  // 获取所有画廊相册
  getGalleryAlbums: (userName: string) => {
    return api.get<GetAlbumsResponse>(`/api/gallery/albums/${userName}`);
  },
  // 获取画廊相册图片
  getGalleryAlbumImages: (userName: string, albumId: number, params: {
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetAlbumImagesResponse>(`/api/gallery/images/${userName}/${albumId}`, { params });
  },
  // 搜索画廊相册图片
  getGallerySearch: (userName: string, params: {
    query: string;
    page?: number;
    page_size?: number;
  }) => {
    return api.get<GetAlbumImagesResponse>(`/api/gallery/search/${userName}`, { params });
  },
  // 获取画廊配置
  getGalleryConfig: () => {
    return api.get<GalleryConfig>('/api/gallery/config');
  },
};

// 存储配置相关API
export const storageApi = {
  // 获取存储配置列表（管理员）
  getStorages: (params: {
    page?: number;
    page_size?: number;
    searchkey?: string;
    orderby?: string;
    order?: string;
  }) => {
    return api.get<Storage[]>('/api/admin/storages', { params });
  },
  // 获取单个存储配置（管理员）
  getStorage: (id: number) => {
    return api.get<Storage>(`/api/admin/storages/${id}`);
  },
  // 根据名称获取存储配置（管理员）
  getStorageByName: (name: string) => {
    return api.get<Storage>(`/api/admin/storages/name/${name}`);
  },
  // 创建存储配置（管理员）
  createStorage: (data: StorageRequest) => {
    return api.post<Storage>('/api/admin/storages', data);
  },
  // 更新存储配置（管理员）
  updateStorage: (id: number, data: StorageRequest) => {
    return api.put<Storage>(`/api/admin/storages/${id}`, data);
  },
  // 删除存储配置（管理员）
  deleteStorage: (id: number) => {
    return api.delete<SuccessResponse>(`/api/admin/storages/${id}`);
  },
  // 测试存储连接（管理员）
  testStorageConnection: (data: StorageRequest) => {
    return api.post<SuccessResponse>('/api/admin/storages/test', data);
  },
};
