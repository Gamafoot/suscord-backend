package file

import (
	"suscord/internal/domain/entity"
)

type FileManager interface {
	Upload(file *entity.File, uploadTo string) (string, error)
	Delete(filepath string) error
}
