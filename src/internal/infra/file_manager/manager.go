package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"suscord/internal/domain/entity"
	"time"

	pkgerr "github.com/pkg/errors"
)

type FileManager interface {
	Upload(file *entity.File, uploadTo string) (string, error)
	Delete(filepath string) error
}

type fileManager struct {
	mediaPath string
}

func NewFileManager(mediaPath string) fileManager {
	return fileManager{mediaPath: mediaPath}
}

func (m fileManager) Upload(file *entity.File, uploadTo string) (string, error) {
	filename := strings.ReplaceAll(filepath.Base(file.Name), " ", "")
	filename = fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename)

	rootpath := filepath.Join(m.mediaPath, uploadTo)
	filePath := filepath.Join(rootpath, filename)

	if err := os.MkdirAll(rootpath, os.ModePerm); err != nil {
		return "", pkgerr.WithStack(err)
	}

	dst, err := os.Create(filePath)
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	successWrite := false
	defer func() {
		dst.Close()
		if !successWrite {
			os.Remove(filePath)
		}
	}()

	if _, err := io.Copy(dst, file.Reader); err != nil {
		return "", pkgerr.WithStack(err)
	}
	successWrite = true

	if err := dst.Close(); err != nil {
		return "", pkgerr.WithStack(err)
	}

	relPath, err := filepath.Rel(m.mediaPath, filePath)
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	return relPath, nil
}

func (m fileManager) Delete(filepath string) error {
	if err := os.Remove(filepath); err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}
