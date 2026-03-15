package entity

import (
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	pkgerr "github.com/pkg/errors"
)

type File struct {
	Name     string
	Size     int64
	MimeType string
	Reader   io.ReadCloser
}

func NewFile(file *multipart.FileHeader) (*File, error) {
	src, err := file.Open()
	if err != nil {
		return nil, pkgerr.WithStack(err)
	}

	ext := filepath.Ext(strings.ToLower(file.Filename))
	return &File{
		Name:     file.Filename,
		Size:     file.Size,
		MimeType: mime.TypeByExtension(ext),
		Reader:   src,
	}, nil
}

func (f *File) Close() error {
	return f.Reader.Close()
}
