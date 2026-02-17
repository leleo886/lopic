// storage/webdav_storage.go
package storage

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"

	"github.com/studio-b12/gowebdav"
	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/internal/log"
)

// WebDAVStorage 实现 Storage 接口
type WebDAVStorage struct {
	client     *gowebdav.Client
	staticPath string // 对外 URL 路径前缀，如 "static"
	basePath   string // WebDAV 服务器内部存储目录，如 "uploads"
}

// NewWebDAVStorage 创建 WebDAV 存储实例
func NewWebDAVStorage(baseURL, username, password, staticURL, basePath string) *WebDAVStorage {
	client := gowebdav.NewClient(baseURL, username, password)
	return &WebDAVStorage{
		client:     client,
		staticPath: strings.Trim(staticURL, "/"),
		basePath:   strings.Trim(basePath, "/"),
	}
}

// UploadFile 上传文件到 WebDAV
func (s *WebDAVStorage) UploadFile(file *multipart.FileHeader, filePath string, uploadPath string, fileName string) (string, error) {
	// 构建 WebDAV 内部路径：/basePath/uploadPath/fileName
	webdavPath := path.Join(s.basePath, uploadPath, fileName)
	webdavPath = path.Clean("/" + webdavPath) // 防路径穿越

	var reader io.Reader

	if file == nil && filePath != "" {
		// 从本地文件读取
		f, err := os.Open(filePath)
		if err != nil {
			log.Errorf("failed to open local file: path=%s, error=%v", filePath, err)
			return "", cerrors.ErrInternalServer
		}
		defer f.Close()
		reader = f
	} else if file != nil {
		// 从 multipart 读取
		src, err := file.Open()
		if err != nil {
			log.Errorf("failed to open multipart file: error=%v", err)
			return "", cerrors.ErrInternalServer
		}
		defer src.Close()
		reader = src
	} else {
		return "", cerrors.ErrInternalServer
	}

	// 自动创建父目录（递归）
	parentDir := path.Dir(webdavPath)
	if err := s.client.MkdirAll(parentDir, 0755); err != nil {
		log.Errorf("failed to create WebDAV directory %q: error=%v", parentDir, err)
		return "", cerrors.ErrInternalServer
	}

	// 使用 WriteStream 上传
	if err := s.client.WriteStream(webdavPath, reader, 0644); err != nil {
		log.Errorf("WebDAV upload failed for %q: error=%v", webdavPath, err)
		return "", cerrors.ErrInternalServer
	}

	// 返回对外 URL 路径
	urlPath := s.staticPath + "/" + strings.Trim(path.Join(s.basePath, uploadPath, fileName), "/")
	return urlPath, nil
}

// DeleteFile 删除 WebDAV 上的文件
func (s *WebDAVStorage) DeleteFile(fileURL string) error {
	prefix := strings.Trim(s.staticPath, "/")
	// 转换为 WebDAV 内部路径
	relativePath := strings.TrimPrefix(fileURL, prefix)

	if err := s.client.Remove(relativePath); err != nil {
		log.Errorf("WebDAV delete failed for %q: error=%v", relativePath, err)
		return cerrors.ErrInternalServer
	}
	return nil
}

// CreateDirectory 创建 WebDAV 目录（支持多级）
func (s *WebDAVStorage) CreateDirectory(dirPath string) error {
	fullPath := path.Join(s.basePath, dirPath)
	fullPath = path.Clean("/" + fullPath)
	return s.client.MkdirAll(fullPath, 0755)
}

// TestConnection 测试 WebDAV 连接是否成功
func (s *WebDAVStorage) TestConnection() error {
//    client := gowebdav.NewClient(s.baseURL, s.username, s.password)
   _,err := s.client.ReadDir("/")
   if err != nil {
       log.Errorf("WebDAV connection test failed: %v", err)
       return cerrors.ErrTestConnectionFailed
   }
   return nil
}
