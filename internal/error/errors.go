package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

var (
	// Common errors
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "unauthorized access",
		StatusCode: http.StatusUnauthorized,
	}
	ErrBadRequest = &AppError{
		Code:       "BAD_REQUEST",
		Message:    "bad request",
		StatusCode: http.StatusBadRequest,
	}
	ErrInternalServer = &AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
	}
	ErrTooManyRequests = &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    "too many requests",
		StatusCode: http.StatusTooManyRequests,
	}

	// Auth errors
	ErrUserNotFound = &AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "user not found",
		StatusCode: http.StatusNotFound,
	}
	ErrRegisterDisabled = &AppError{
		Code:       "REGISTER_DISABLED",
		Message:    "register is disabled",
		StatusCode: http.StatusForbidden,
	}
	ErrCSRFTokenInvalid = &AppError{
		Code:       "CSRF_TOKEN_INVALID",
		Message:    "CSRF token is invalid or expired",
		StatusCode: http.StatusForbidden,
	}
	ErrUsernameExists = &AppError{
		Code:       "USERNAME_EXISTS",
		Message:    "username already exists",
		StatusCode: http.StatusConflict,
	}
	ErrEmailExists = &AppError{
		Code:       "EMAIL_EXISTS",
		Message:    "email already exists",
		StatusCode: http.StatusConflict,
	}
	ErrUserAlreadyExists = &AppError{
		Code:       "USER_ALREADY_EXISTS",
		Message:    "username or email already exists",
		StatusCode: http.StatusConflict,
	}
	ErrInvalidCredentials = &AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "invalid username or password",
		StatusCode: http.StatusUnauthorized,
	}
	ErrInvalidResetCode = &AppError{
		Code:       "INVALID_RESET_CODE",
		Message:    "invalid or expired reset code",
		StatusCode: http.StatusBadRequest,
	}
	ErrUserNotActive = &AppError{
		Code:       "USER_NOT_ACTIVE",
		Message:    "user not active",
		StatusCode: http.StatusForbidden,
	}
	ErrUserAlreadyActive = &AppError{
		Code:       "USER_ALREADY_ACTIVE",
		Message:    "user already active",
		StatusCode: http.StatusForbidden,
	}
	ErrSendEmailFailed = &AppError{
		Code:       "SEND_EMAIL_FAILED",
		Message:    "send email failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrFailedToSendEmail = &AppError{
		Code:       "FAILED_TO_SEND_EMAIL",
		Message:    "failed to send email",
		StatusCode: http.StatusInternalServerError,
	}
	ErrCreateUserFailed = &AppError{
		Code:       "CREATE_USER_FAILED",
		Message:    "create user failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrUserRoleNotFound = &AppError{
		Code:       "USER_ROLE_NOT_FOUND",
		Message:    "user role not found",
		StatusCode: http.StatusNotFound,
	}
	ErrPwdError = &AppError{
		Code:       "PWD_ERROR",
		Message:    "password error",
		StatusCode: http.StatusUnauthorized,
	}
	ErrPwdEncFailed = &AppError{
		Code:       "Pwd_Enc_Failed",
		Message:    "password encryption failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrGenerateTokenFailed = &AppError{
		Code:       "GENERATE_TOKEN_FAILED",
		Message:    "generate token failed",
		StatusCode: http.StatusInternalServerError,
	}
	ErrResetTokenInvalid = &AppError{
		Code:       "RESET_TOKEN_INVALID",
		Message:    "reset token is invalid or expired",
		StatusCode: http.StatusBadRequest,
	}
	ErrResetTokenExpired = &AppError{
		Code:       "RESET_TOKEN_EXPIRED",
		Message:    "reset token has expired",
		StatusCode: http.StatusBadRequest,
	}
	ErrMailServiceDisabled = &AppError{
		Code:       "MAIL_SERVICE_DISABLED",
		Message:    "mail service is disabled",
		StatusCode: http.StatusServiceUnavailable,
	}
	ErrOneAdmin = &AppError{
		Code:       "ONE_ADMIN",
		Message:    "system only has one admin user",
		StatusCode: http.StatusBadRequest,
	}
	ErrCannotDisableSelf = &AppError{
		Code:       "CANNOT_DISABLE_SELF",
		Message:    "cannot disable your own account",
		StatusCode: http.StatusForbidden,
	}
	ErrCannotDeleteAdminUser = &AppError{
		Code:       "CANNOT_DELETE_ADMIN_USER",
		Message:    "cannot delete admin user",
		StatusCode: http.StatusForbidden,
	}
	ErrFailedToDeleteUser = &AppError{
		Code:       "FAILED_TO_DELETE_USER",
		Message:    "failed to delete user",
		StatusCode: http.StatusInternalServerError,
	}
	ErrInvalidSigningMethod = &AppError{
		Code:       "INVALID_SIGNING_METHOD",
		Message:    "invalid signing method",
		StatusCode: http.StatusBadRequest,
	}
	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "invalid token or expired",
		StatusCode: http.StatusUnauthorized,
	}
	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "forbidden access",
		StatusCode: http.StatusForbidden,
	}

	// Role errors
	ErrRoleNotFound = &AppError{
		Code:       "ROLE_NOT_FOUND",
		Message:    "role not found",
		StatusCode: http.StatusNotFound,
	}
	ErrCannotDeleteDefaultRole = &AppError{
		Code:       "CANNOT_DELETE_DEFAULT_ROLE",
		Message:    "cannot delete default role",
		StatusCode: http.StatusForbidden,
	}
	ErrRoleAssociatedWithUsers = &AppError{
		Code:       "ROLE_ASSOCIATED_WITH_USERS",
		Message:    "role is associated with existing users, cannot be deleted",
		StatusCode: http.StatusForbidden,
	}
	ErrCannotChangeAdminRoleName = &AppError{
		Code:       "CANNOT_CHANGE_ADMIN_ROLE_NAME",
		Message:    "cannot change admin role name",
		StatusCode: http.StatusForbidden,
	}
	ErrCannotChangeUserRoleName = &AppError{
		Code:       "CANNOT_CHANGE_USER_ROLE_NAME",
		Message:    "cannot change user role name",
		StatusCode: http.StatusForbidden,
	}
	ErrRoleNameExists = &AppError{
		Code:       "ROLE_NAME_EXISTS",
		Message:    "role name already exists",
		StatusCode: http.StatusConflict,
	}

	// Album-Image errors
	ErrAlbumNotFound = &AppError{
		Code:       "ALBUM_NOT_FOUND",
		Message:    "album not found",
		StatusCode: http.StatusNotFound,
	}
	ErrImageNotFound = &AppError{
		Code:       "IMAGE_NOT_FOUND",
		Message:    "image not found",
		StatusCode: http.StatusNotFound,
	}
	ErrImageNotInAlbum = &AppError{
		Code:       "IMAGE_NOT_IN_ALBUM",
		Message:    "image not in album",
		StatusCode: http.StatusBadRequest,
	}
	ErrImageAlreadyInAlbum = &AppError{
		Code:       "IMAGE_ALREADY_IN_ALBUM",
		Message:    "image already in album",
		StatusCode: http.StatusConflict,
	}
	ErrUnsupportedMimeType = &AppError{
		Code:       "UNSUPPORTED_MIME_TYPE",
		Message:    "unsupported mime type",
		StatusCode: http.StatusBadRequest,
	}
	ErrMaxFilesPerUpload = &AppError{
		Code:       "MAX_FILES_PER_UPLOAD",
		Message:    "exceeded maximum number of files per upload",
		StatusCode: http.StatusForbidden,
	}
	ErrAllowedExtensions = &AppError{
		Code:       "ALLOWED_EXTENSIONS",
		Message:    "only supported file extensions are allowed",
		StatusCode: http.StatusBadRequest,
	}
	ErrMaxFileSizeMB = &AppError{
		Code:       "MAX_FILE_SIZE_MB",
		Message:    "exceeded maximum file size per single file",
		StatusCode: http.StatusForbidden,
	}
	ErrMaxStorageSizeMB = &AppError{
		Code:       "MAX_STORAGE_SIZE_MB",
		Message:    "exceeded maximum storage size",
		StatusCode: http.StatusForbidden,
	}
	ErrMaxAlbumsPerUser = &AppError{
		Code:       "ALBUM_LIMIT_EXCEEDED",
		Message:    "exceeded maximum number of albums per user",
		StatusCode: http.StatusForbidden,
	}
	ErrDecodeImage = &AppError{
		Code:       "DECODE_IMAGE_ERROR",
		Message:    "error decoding image",
		StatusCode: http.StatusBadRequest,
	}
	ErrEncodeImage = &AppError{
		Code:       "ENCODE_IMAGE_ERROR",
		Message:    "error encoding image",
		StatusCode: http.StatusInternalServerError,
	}
	ErrGalleryPermissionDenied = &AppError{
		Code:       "GALLERY_PERMISSION_DENIED",
		Message:    "gallery is disabled for this role",
		StatusCode: http.StatusForbidden,
	}
	ErrUnknownFileType = &AppError{
		Code:       "UNKNOWN_FILE_TYPE",
		Message:    "unknown file type",
		StatusCode: http.StatusBadRequest,
	}



	// Storage errors
	ErrStorageNotFound = &AppError{
		Code:       "STORAGE_NOT_FOUND",
		Message:    "storage configuration not found",
		StatusCode: http.StatusNotFound,
	}
	ErrStorageNameExists = &AppError{
		Code:       "STORAGE_NAME_EXISTS",
		Message:    "storage name already exists",
		StatusCode: http.StatusConflict,
	}
	ErrCannotDeleteLocalStorage = &AppError{
		Code:       "CANNOT_DELETE_LOCAL_STORAGE",
		Message:    "cannot delete local storage",
		StatusCode: http.StatusForbidden,
	}
	ErrCannotChangeLocalStorageName = &AppError{
		Code:       "CANNOT_CHANGE_LOCAL_STORAGE_NAME",
		Message:    "cannot change local storage name",
		StatusCode: http.StatusForbidden,
	}
	ErrCannotChangeLocalStorageType = &AppError{
		Code:       "CANNOT_CHANGE_LOCAL_STORAGE_TYPE",
		Message:    "cannot change local storage type",
		StatusCode: http.StatusForbidden,
	}
	ErrStorageAssociatedWithRoles = &AppError{
		Code:       "STORAGE_ASSOCIATED_WITH_ROLES",
		Message:    "storage is associated with existing roles, cannot be deleted",
		StatusCode: http.StatusForbidden,
	}
	ErrStorageAssociatedWithImages = &AppError{
		Code:       "STORAGE_ASSOCIATED_WITH_IMAGES",
		Message:    "storage is associated with existing images, cannot be deleted",
		StatusCode: http.StatusForbidden,
	}
	ErrStorageConnectionFailed = &AppError{
		Code:       "STORAGE_CONNECTION_FAILED",
		Message:    "failed to connect to storage",
		StatusCode: http.StatusBadRequest,
	}
	ErrTestConnectionFailed = &AppError{
		Code:       "TEST_CONNECTION_FAILED",
		Message:    "failed to test storage connection",
		StatusCode: http.StatusBadRequest,
	}

	// Backup errors
	ErrBackupNotFound = &AppError{
		Code:       "BACKUP_NOT_FOUND",
		Message:    "backup file not found",
		StatusCode: http.StatusNotFound,
	}
	ErrBackupDatabase = &AppError{
		Code:       "BACKUP_DATABASE_ERROR",
		Message:    "error backing up database",
		StatusCode: http.StatusInternalServerError,
	}
	ErrBackupFiles = &AppError{
		Code:       "BACKUP_FILES_ERROR",
		Message:    "error backing up files",
		StatusCode: http.StatusInternalServerError,
	}
	ErrBackupTaskNotCompleted = &AppError{
		Code:       "BACKUP_TASK_NOT_COMPLETED",
		Message:    "backup task status is not completed",
		StatusCode: http.StatusBadRequest,
	}
	ErrExtractBackup = &AppError{
		Code:       "EXTRACT_BACKUP_ERROR",
		Message:    "error extracting backup file",
		StatusCode: http.StatusInternalServerError,
	}
	ErrNoValidDBBackup = &AppError{
		Code:       "NO_VALID_DB_BACKUP",
		Message:    "no valid database backup file found in archive",
		StatusCode: http.StatusBadRequest,
	}
	ErrDBTypeMismatch = &AppError{
		Code:       "DB_TYPE_MISMATCH",
		Message:    "backup database type does not match current database type",
		StatusCode: http.StatusBadRequest,
	}
	ErrMySQLRestore = &AppError{
		Code:       "MYSQL_RESTORE_ERROR",
		Message:    "error restoring MySQL database",
		StatusCode: http.StatusInternalServerError,
	}
	ErrSQLiteRestore = &AppError{
		Code:       "SQLITE_RESTORE_ERROR",
		Message:    "error restoring SQLite database",
		StatusCode: http.StatusInternalServerError,
	}
	ErrRestoreFiles = &AppError{
		Code:       "RESTORE_FILES_ERROR",
		Message:    "error restoring files",
		StatusCode: http.StatusInternalServerError,
	}

)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(err error) (int, *ErrorResponse) {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode, &ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}
	return http.StatusInternalServerError, &ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: err.Error(),
	}
}
