package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
)

// LocalStorage 本地存储实现
type LocalStorage struct {
	BasePath   string
	StaticPath string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(basePath, staticPath string) *LocalStorage {
	return &LocalStorage{
		BasePath:   basePath,
		StaticPath: staticPath,
	}
}

// UploadFile 上传文件到本地存储
func (s *LocalStorage) UploadFile(file *multipart.FileHeader, filePath string, uploadPath string, fileName string) (string, error) {
	// 安全检查：验证 uploadPath 和 fileName 不包含路径遍历字符
	if containsPathTraversal(uploadPath) {
		log.Errorf("upload path contains traversal characters: %s", uploadPath)
		return "", cerrors.ErrForbidden
	}
	if containsPathTraversal(fileName) {
		log.Errorf("file name contains traversal characters: %s", fileName)
		return "", cerrors.ErrForbidden
	}

	// 创建完整的存储路径
	dstPath := filepath.Join(".", s.BasePath, uploadPath, fileName)

	// 解析绝对路径以消除 .. 等路径遍历元素
	absPath, err := filepath.Abs(dstPath)
	if err != nil {
		log.Errorf("failed to resolve absolute path: %v", err)
		return "", cerrors.ErrInternalServer
	}

	// 计算基础目录的绝对路径
	baseAbsPath, err := filepath.Abs(filepath.Join(".", s.BasePath))
	if err != nil {
		log.Errorf("failed to resolve base path: %v", err)
		return "", cerrors.ErrInternalServer
	}

	// 安全检查：确保目标路径在基础目录内（防止路径遍历攻击）
	if !strings.HasPrefix(absPath, baseAbsPath+string(filepath.Separator)) {
		log.Errorf("path traversal detected: %s is outside of %s", absPath, baseAbsPath)
		return "", cerrors.ErrForbidden
	}

	// 创建目录
	if err := s.CreateDirectory(filepath.Dir(absPath)); err != nil {
		log.Errorf("failed to create directory: path=%s, error=%v", filepath.Dir(absPath), err)
		return "", err
	}

	if file == nil && filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Errorf("failed to read file: path=%s, error=%v", filePath, err)
			return "", cerrors.ErrInternalServer
		}
		err = os.WriteFile(absPath, content, 0644)
		if err != nil {
			log.Errorf("failed to write file: path=%s, error=%v", absPath, err)
			return "", cerrors.ErrInternalServer
		}
	} else if file != nil {
		// 从 multipart.FileHeader 创建 io.Reader
		src, err := file.Open()
		if err != nil {
			log.Error("failed to open file")
			return "", cerrors.ErrInternalServer
		}
		defer src.Close()

		dst, err := os.Create(absPath)
		if err != nil {
			log.Errorf("failed to create file: path=%s, error=%v", absPath, err)
			return "", cerrors.ErrInternalServer
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			log.Errorf("failed to copy file: path=%s, error=%v", absPath, err)
			return "", cerrors.ErrInternalServer
		}
	}

	urlPath := filepath.Join(s.StaticPath, uploadPath, fileName)
	urlPath = filepath.ToSlash(urlPath)
	return urlPath, nil
}

// DeleteFile 从本地存储删除文件
func (s *LocalStorage) DeleteFile(fileURL string) error {
	// 安全检查：验证 fileURL 是否以 StaticPath 开头
	if !strings.HasPrefix(fileURL, s.StaticPath) {
		log.Errorf("invalid file URL: %s does not start with %s", fileURL, s.StaticPath)
		return cerrors.ErrForbidden
	}

	// 将 URL 路径转换为本地路径
	relativePath := strings.TrimPrefix(fileURL, s.StaticPath)
	relativePath = strings.TrimPrefix(relativePath, "/")

	// 构建完整的本地路径
	fullPath := filepath.Join(".", s.BasePath, relativePath)

	// 解析绝对路径以消除 .. 等路径遍历元素
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		log.Errorf("failed to resolve absolute path: %v", err)
		return cerrors.ErrInternalServer
	}

	// 计算基础目录的绝对路径
	baseAbsPath, err := filepath.Abs(filepath.Join(".", s.BasePath))
	if err != nil {
		log.Errorf("failed to resolve base path: %v", err)
		return cerrors.ErrInternalServer
	}

	// 安全检查：确保目标路径在基础目录内（防止路径遍历攻击）
	if !strings.HasPrefix(absPath, baseAbsPath+string(filepath.Separator)) && absPath != baseAbsPath {
		log.Errorf("path traversal detected: %s is outside of %s", absPath, baseAbsPath)
		return cerrors.ErrForbidden
	}

	return os.Remove(absPath)
}

// CreateDirectory 创建本地目录
func (s *LocalStorage) CreateDirectory(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// TestConnection 测试本地存储连接是否成功
func (s *LocalStorage) TestConnection() error {
	// 检查基础路径是否存在，如果不存在则尝试创建
	return os.MkdirAll(filepath.Join(".", s.BasePath), 0755)
}

// containsPathTraversal 检查路径是否包含路径遍历字符
func containsPathTraversal(path string) bool {
	// 检查是否包含 .. 或 . 开头的路径元素
	if path == "" {
		return false
	}

	// 分割路径并检查每个部分
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." || part == "." {
			return true
		}
	}

	// 同时检查正斜杠（URL 风格路径）
	parts = strings.Split(path, "/")
	for _, part := range parts {
		if part == ".." || part == "." {
			return true
		}
	}

	return false
}
