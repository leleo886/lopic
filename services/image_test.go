package services

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"testing"

	cerrors "github.com/leleo886/lopic/internal/error"
	"github.com/leleo886/lopic/models"
)

func createMultipartFileHeader(content []byte, filename string) (*multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	_, err = part.Write(content)
	if err != nil {
		return nil, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", "/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(32 << 20)
	if err != nil {
		return nil, err
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return req.MultipartForm.File["file"][0], nil
}

func createTestImage(width, height int, format string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	buf := new(bytes.Buffer)

	switch format {
	case "png":
		err := png.Encode(buf, img)
		if err != nil {
			return nil, err
		}
	case "jpeg", "jpg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 85})
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func TestGetFileType(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		filename    string
		expectError error
	}{
		{
			name:        "valid jpeg file",
			filename:    "test.jpg",
			expectError: nil,
		},
		{
			name:        "valid png file",
			filename:    "test.png",
			expectError: nil,
		},
		{
			name:        "invalid file type",
			content:     []byte("this is not an image file"),
			filename:    "test.txt",
			expectError: cerrors.ErrUnknownFileType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var imgBytes []byte
			var err error

			if tt.content != nil {
				imgBytes = tt.content
			} else {
				imgBytes, err = createTestImage(100, 100, "jpeg")
				if err != nil {
					t.Fatalf("failed to create test image: %v", err)
				}
			}

			fileHeader, err := createMultipartFileHeader(imgBytes, tt.filename)
			if err != nil {
				t.Fatalf("failed to create multipart file header: %v", err)
			}

			_, _, err = GetFileType(fileHeader)

			if tt.expectError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGetImageDimensions(t *testing.T) {
	tests := []struct {
		name           string
		width          int
		height         int
		format         string
		filename       string
		expectedWidth  int
		expectedHeight int
		expectError    bool
	}{
		{
			name:           "jpeg image dimensions",
			width:          800,
			height:         600,
			format:         "jpeg",
			filename:       "test.jpg",
			expectedWidth:  800,
			expectedHeight: 600,
			expectError:    false,
		},
		{
			name:           "png image dimensions",
			width:          1024,
			height:         768,
			format:         "png",
			filename:       "test.png",
			expectedWidth:  1024,
			expectedHeight: 768,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imgBytes, err := createTestImage(tt.width, tt.height, tt.format)
			if err != nil {
				t.Fatalf("failed to create test image: %v", err)
			}

			fileHeader, err := createMultipartFileHeader(imgBytes, tt.filename)
			if err != nil {
				t.Fatalf("failed to create multipart file header: %v", err)
			}

			mimeType, width, height, err := GetImageDimensions(fileHeader)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if width != tt.expectedWidth {
				t.Errorf("expected width %d, got %d", tt.expectedWidth, width)
			}

			if height != tt.expectedHeight {
				t.Errorf("expected height %d, got %d", tt.expectedHeight, height)
			}

			if mimeType == "" {
				t.Error("mime type should not be empty")
			}
		})
	}
}

func TestMakeImageWithAlbum(t *testing.T) {
	imageModel := models.Image{
		FileName:        "test-uuid$-$test.jpg",
		OriginalName:    "test",
		Tags:            []string{"tag1", "tag2"},
		FileURL:         "/uploads/test.jpg",
		FileSize:        1024,
		Width:           800,
		Height:          600,
		ThumbnailURL:    "/uploads/test_thumbnail.jpg",
		ThumbnailSize:   512,
		ThumbnailWidth:  400,
		ThumbnailHeight: 300,
		MimeType:        "image/jpeg",
		UserID:          1,
		StorageName:     "local",
	}

	result := MakeImageWithAlbum(imageModel)

	if result.FileName != imageModel.FileName {
		t.Errorf("expected filename %s, got %s", imageModel.FileName, result.FileName)
	}

	if result.OriginalName != imageModel.OriginalName {
		t.Errorf("expected original name %s, got %s", imageModel.OriginalName, result.OriginalName)
	}

	if len(result.Tags) != len(imageModel.Tags) {
		t.Errorf("expected %d tags, got %d", len(imageModel.Tags), len(result.Tags))
	}

	if result.FileURL != imageModel.FileURL {
		t.Errorf("expected file URL %s, got %s", imageModel.FileURL, result.FileURL)
	}

	if result.Width != imageModel.Width {
		t.Errorf("expected width %d, got %d", imageModel.Width, result.Width)
	}

	if result.Height != imageModel.Height {
		t.Errorf("expected height %d, got %d", imageModel.Height, result.Height)
	}
}

func TestMakeImagesWithAlbum(t *testing.T) {
	imageModels := []models.Image{
		{
			FileName:     "test1.jpg",
			OriginalName: "test1",
			Tags:         []string{"tag1"},
			FileURL:      "/uploads/test1.jpg",
			FileSize:     1024,
			Width:        800,
			Height:       600,
			UserID:       1,
		},
		{
			FileName:     "test2.jpg",
			OriginalName: "test2",
			Tags:         []string{"tag2"},
			FileURL:      "/uploads/test2.jpg",
			FileSize:     2048,
			Width:        1024,
			Height:       768,
			UserID:       1,
		},
	}

	results := MakeImagesWithAlbum(imageModels)

	if len(results) != len(imageModels) {
		t.Errorf("expected %d results, got %d", len(imageModels), len(results))
	}

	for i, result := range results {
		if result.FileName != imageModels[i].FileName {
			t.Errorf("expected filename %s, got %s", imageModels[i].FileName, result.FileName)
		}
	}
}

func TestImageResponseFields(t *testing.T) {
	imageModel := models.Image{
		FileName:        "test.jpg",
		OriginalName:    "test",
		Tags:            []string{"nature", "landscape"},
		FileURL:         "/uploads/test.jpg",
		FileSize:        1024000,
		Width:           1920,
		Height:          1080,
		ThumbnailURL:    "/uploads/test_thumbnail.jpg",
		ThumbnailSize:   51200,
		ThumbnailWidth:  640,
		ThumbnailHeight: 360,
		MimeType:        "image/jpeg",
		UserID:          1,
		StorageName:     "local",
	}

	response := MakeImageWithAlbum(imageModel)

	if response.MimeType != "image/jpeg" {
		t.Errorf("expected mime type image/jpeg, got %s", response.MimeType)
	}

	if response.StorageName != "local" {
		t.Errorf("expected storage name local, got %s", response.StorageName)
	}

	if response.ThumbnailWidth != 640 {
		t.Errorf("expected thumbnail width 640, got %d", response.ThumbnailWidth)
	}

	if response.ThumbnailHeight != 360 {
		t.Errorf("expected thumbnail height 360, got %d", response.ThumbnailHeight)
	}
}
