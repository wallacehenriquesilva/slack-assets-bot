package fileutil

import (
	"archive/zip"
	"fmt"
	"github.com/gofrs/uuid"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	LocalPath  string
	RemotePath string
}

func NewFile(extension string) (*os.File, error) {
	fileName, err := generateFileName(extension)
	if err != nil {
		return nil, err
	}

	csvFile, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return csvFile, nil
}

// generateFileName creates an uuid4 file name in the temp folder.
func generateFileName(extension string) (string, error) {
	u4, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	tempDir := os.TempDir()
	return fmt.Sprintf("%s/%s.%s", tempDir, u4, extension), nil
}

func UnzipFiles(path string, ignoreFileFunction func(string) bool) ([]File, error) {
	var files []File

	u4, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	destination := u4.String()
	archive, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer archive.Close()

	for _, f := range archive.File {
		if ignoreFileFunction(f.Name) {
			continue
		}

		filePath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return nil, nil
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return nil, err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return nil, err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return nil, err
		}

		files = append(files, File{
			LocalPath:  filePath,
			RemotePath: f.Name,
		})

		_ = dstFile.Close()
		_ = fileInArchive.Close()
	}
	return files, nil
}

func DeleteFiles(files ...string) error {
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}
	return nil
}
