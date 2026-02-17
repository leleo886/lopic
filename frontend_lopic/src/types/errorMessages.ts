// 错误消息映射 - 对应后端 errors.go 中的错误代码
import i18n from '../i18n';

// 多语言错误消息映射
export const errorMessages: Record<string, Record<string, string>> = {
  zh: {
    // Common errors
    'UNAUTHORIZED': '未授权访问',
    'BAD_REQUEST': '请求参数错误',
    'INTERNAL_SERVER_ERROR': '服务器内部错误',
    
    // Auth errors
    'USER_NOT_FOUND': '用户不存在',
    'REGISTER_DISABLED': '管理员未开启注册功能',
    'CSRF_TOKEN_INVALID': 'CSRF令牌无效或已过期',
    'USERNAME_EXISTS': '用户名已存在',
    'EMAIL_EXISTS': '邮箱已存在',
    'USER_NOT_ACTIVE': '用户账户未激活',
    'USER_ALREADY_ACTIVE': '用户账户已激活',
    'SEND_EMAIL_FAILED': '发送邮件失败',
    'FAILED_TO_SEND_EMAIL': '发送邮件失败',
    'CREATE_USER_FAILED': '创建用户失败',
    'USER_ROLE_NOT_FOUND': '用户角色不存在',
    'PWD_ERROR': '密码错误',
    'Pwd_Enc_Failed': '密码加密失败',
    'GENERATE_TOKEN_FAILED': '生成令牌失败',
    'RESET_TOKEN_INVALID': '重置令牌无效或已过期',
    'RESET_TOKEN_EXPIRED': '重置令牌已过期',
    'MAIL_SERVICE_DISABLED': '邮件服务已禁用',
    'ONE_ADMIN': '系统只有一个管理员用户',
    'CANNOT_DELETE_ADMIN_USER': '不能删除管理员用户',
    'FAILED_TO_DELETE_USER': '删除用户失败',
    'INVALID_SIGNING_METHOD': '无效的签名方法',
    'INVALID_TOKEN': '令牌无效或已过期',
    'FORBIDDEN': '禁止访问',
    'TOO_MANY_REQUESTS': '请求次数过多，请稍后重试',
    'CANNOT_DISABLE_SELF': '不能禁用自己',
    'USER_ALREADY_EXISTS': '用户已存在',
    'INVALID_CREDENTIALS': '无效的用户名或密码',
    'INVALID_RESET_CODE': '无效的重置码',
    
    // Role errors
    'ROLE_NOT_FOUND': '角色不存在',
    'CANNOT_DELETE_DEFAULT_ROLE': '不能删除默认角色',
    'ROLE_ASSOCIATED_WITH_USERS': '角色已关联现有用户，请先删除关联用户后再删除角色',
    'CANNOT_CHANGE_ADMIN_ROLE_NAME': '不能修改管理员角色名称',
    'CANNOT_CHANGE_USER_ROLE_NAME': '不能修改用户角色名称',
    'ROLE_NAME_EXISTS': '角色名称已存在',
    
    // Album-Image errors
    'ALBUM_NOT_FOUND': '相册不存在',
    'IMAGE_NOT_FOUND': '图片不存在',
    'IMAGE_NOT_IN_ALBUM': '图片不在相册中',
    'IMAGE_ALREADY_IN_ALBUM': '图片已在相册中',
    'UNSUPPORTED_MIME_TYPE': '不支持的文件类型',
    'MAX_FILES_PER_UPLOAD': '超出单次上传文件数量限制',
    'ALLOWED_EXTENSIONS': '只允许上传支持的文件格式',
    'MAX_FILE_SIZE_MB': '超出单个文件大小限制',
    'MAX_STORAGE_SIZE_MB': '超出存储空间限制',
    'ALBUM_LIMIT_EXCEEDED': '超出每个用户相册数量限制',
    'DECODE_IMAGE_ERROR': '图片解码错误',
    'ENCODE_IMAGE_ERROR': '图片编码错误',
    'GALLERY_PERMISSION_DENIED': '该角色的画廊功能已禁用',
    'UNKNOWN_FILE_TYPE': '未知文件类型',

    // Backup errors
    'BACKUP_DATABASE_ERROR': '备份数据库失败',
    'BACKUP_FILES_ERROR': '备份文件失败',
    'BACKUP_TASK_NOT_COMPLETED': '备份任务未完成',
    'EXTRACT_BACKUP_ERROR': '解压备份文件失败',
    'NO_VALID_DB_BACKUP': '备份文件中未找到有效的数据库备份',
    'DB_TYPE_MISMATCH': '备份数据库类型与当前数据库类型不匹配',
    'MYSQL_RESTORE_ERROR': '恢复MySQL数据库失败',
    'SQLITE_RESTORE_ERROR': '恢复SQLite数据库失败',
    'RESTORE_FILES_ERROR': '恢复文件失败',
    'BACKUP_NOT_FOUND': '备份文件不存在',
    
    // Storage errors
    'STORAGE_NOT_FOUND': '存储配置不存在',
    'STORAGE_NAME_EXISTS': '存储名称已存在',
    'CANNOT_DELETE_LOCAL_STORAGE': '不能删除本地存储',
    'CANNOT_CHANGE_LOCAL_STORAGE_NAME': '不能修改本地存储名称',
    'CANNOT_CHANGE_LOCAL_STORAGE_TYPE': '不能修改本地存储类型',
    'STORAGE_ASSOCIATED_WITH_ROLES': '存储已关联现有角色，请先删除关联角色后再删除存储',
    'STORAGE_ASSOCIATED_WITH_IMAGES': '存储已关联现有图片，请先删除关联图片后再删除存储',
    'STORAGE_CONNECTION_FAILED': '连接存储失败',
    'TEST_CONNECTION_FAILED': '测试连接存储失败',

  },
  en: {
    // Common errors
    'UNAUTHORIZED': 'Unauthorized access',
    'BAD_REQUEST': 'Invalid request parameters',
    'INTERNAL_SERVER_ERROR': 'Internal server error',
    
    // Auth errors
    'USER_NOT_FOUND': 'User not found',
    'REGISTER_DISABLED': 'Registration is disabled by admin',
    'CSRF_TOKEN_INVALID': 'CSRF token is invalid or expired',
    'USERNAME_EXISTS': 'Username already exists',
    'EMAIL_EXISTS': 'Email already exists',
    'USER_NOT_ACTIVE': 'User account is not active',
    'USER_ALREADY_ACTIVE': 'User account is already active',
    'SEND_EMAIL_FAILED': 'Failed to send email',
    'FAILED_TO_SEND_EMAIL': 'Failed to send email',
    'CREATE_USER_FAILED': 'Failed to create user',
    'USER_ROLE_NOT_FOUND': 'User role not found',
    'PWD_ERROR': 'Password error',
    'Pwd_Enc_Failed': 'Password encryption failed',
    'GENERATE_TOKEN_FAILED': 'Failed to generate token',
    'RESET_TOKEN_INVALID': 'Reset token is invalid or expired',
    'RESET_TOKEN_EXPIRED': 'Reset token has expired',
    'MAIL_SERVICE_DISABLED': 'Email service is disabled',
    'ONE_ADMIN': 'Only one admin user exists in the system',
    'CANNOT_DELETE_ADMIN_USER': 'Cannot delete admin user',
    'FAILED_TO_DELETE_USER': 'Failed to delete user',
    'INVALID_SIGNING_METHOD': 'Invalid signing method',
    'INVALID_TOKEN': 'Token is invalid or expired',
    'FORBIDDEN': 'Access forbidden',
    'TOO_MANY_REQUESTS': 'Too many requests, please try again later',
    'CANNOT_DISABLE_SELF': 'Cannot disable self',
    'USER_ALREADY_EXISTS': 'User already exists',
    'INVALID_CREDENTIALS': 'Invalid username or password',
    'INVALID_RESET_CODE': 'Invalid reset code',
    
    // Role errors
    'ROLE_NOT_FOUND': 'Role not found',
    'CANNOT_DELETE_DEFAULT_ROLE': 'Cannot delete default role',
    'ROLE_ASSOCIATED_WITH_USERS': 'Role is associated with existing users, please delete associated users first before deleting the role',
    'CANNOT_CHANGE_ADMIN_ROLE_NAME': 'Cannot change admin role name',
    'CANNOT_CHANGE_USER_ROLE_NAME': 'Cannot change user role name',
    'ROLE_NAME_EXISTS': 'Role name already exists',
    
    // Album-Image errors
    'ALBUM_NOT_FOUND': 'Album not found',
    'IMAGE_NOT_FOUND': 'Image not found',
    'IMAGE_NOT_IN_ALBUM': 'Image not in album',
    'IMAGE_ALREADY_IN_ALBUM': 'Image already in album',
    'UNSUPPORTED_MIME_TYPE': 'Unsupported file type',
    'MAX_FILES_PER_UPLOAD': 'Exceeded maximum files per upload limit',
    'ALLOWED_EXTENSIONS': 'Only allowed file formats are permitted',
    'MAX_FILE_SIZE_MB': 'Exceeded single file size limit',
    'MAX_STORAGE_SIZE_MB': 'Exceeded storage space limit',
    'ALBUM_LIMIT_EXCEEDED': 'Exceeded maximum albums per user limit',
    'DECODE_IMAGE_ERROR': 'Image decoding error',
    'ENCODE_IMAGE_ERROR': 'Image encoding error',
    'GALLERY_PERMISSION_DENIED': 'Gallery function is disabled for this role',
    'UNKNOWN_FILE_TYPE': 'Unknown file type',

    // Backup errors
    'BACKUP_DATABASE_ERROR': 'Failed to backup database',
    'BACKUP_FILES_ERROR': 'Failed to backup files',
    'BACKUP_TASK_NOT_COMPLETED': 'Backup task is not completed',
    'EXTRACT_BACKUP_ERROR': 'Failed to extract backup file',
    'NO_VALID_DB_BACKUP': 'No valid database backup file found in archive',
    'DB_TYPE_MISMATCH': 'Backup database type does not match current database type',
    'MYSQL_RESTORE_ERROR': 'Failed to restore MySQL database',
    'SQLITE_RESTORE_ERROR': 'Failed to restore SQLite database',
    'RESTORE_FILES_ERROR': 'Failed to restore files',
    'BACKUP_NOT_FOUND': 'Backup file not found',
    
    // Storage errors
    'STORAGE_NOT_FOUND': 'Storage configuration not found',
    'STORAGE_NAME_EXISTS': 'Storage name already exists',
    'CANNOT_DELETE_LOCAL_STORAGE': 'Cannot delete local storage',
    'CANNOT_CHANGE_LOCAL_STORAGE_NAME': 'Cannot change local storage name',
    'CANNOT_CHANGE_LOCAL_STORAGE_TYPE': 'Cannot change local storage type',
    'STORAGE_ASSOCIATED_WITH_ROLES': 'Storage is associated with existing roles, please delete associated roles first before deleting the storage',
    'STORAGE_ASSOCIATED_WITH_IMAGES': 'Storage is associated with existing images, please delete associated images first before deleting the storage',
    'STORAGE_CONNECTION_FAILED': 'Failed to connect to storage',
    'TEST_CONNECTION_FAILED': 'Failed to test storage connection',
  },
};

/**
 * 根据错误代码获取对应语言的错误消息
 * @param code 错误代码
 * @param defaultMessage 默认消息
 * @returns 对应语言的错误消息
 */
export const getErrorMessage = (code: string, defaultMessage: string = '操作失败，请稍后重试'): string => {
  // 获取当前语言
  const currentLanguage = i18n.global.locale.value;
  // 确保语言代码存在于错误消息映射中
  const language = errorMessages[currentLanguage] ? currentLanguage : 'zh';
  // 获取对应语言的错误消息
  return errorMessages[language]?.[code] ?? defaultMessage;
};
