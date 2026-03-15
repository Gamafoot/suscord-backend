package storage

import (
	"context"
	"suscord/internal/domain/entity"
	domainErrors "suscord/internal/domain/errors"
	"suscord/internal/infra/database/relational/model"

	uuidlib "github.com/google/uuid"
	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

type sessionStorage struct {
	db *gorm.DB
}

func NewSessionStorage(db *gorm.DB) *sessionStorage {
	return &sessionStorage{db: db}
}

func (repo *sessionStorage) GetByUUID(ctx context.Context, uuid string) (*entity.Session, error) {
	session := new(model.Session)
	if err := repo.db.WithContext(ctx).First(&session, "uuid = ?", uuid).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrRecordNotFound
		}
		return nil, pkgerr.WithStack(err)
	}

	return sessionModelToDomain(session), nil
}

func (repo *sessionStorage) GetByUserID(ctx context.Context, userID uint) (*entity.Session, error) {
	session := new(model.Session)
	if err := repo.db.WithContext(ctx).First(&session, "user_id = ?", userID).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainErrors.ErrRecordNotFound
		}
		return nil, pkgerr.WithStack(err)
	}

	return sessionModelToDomain(session), nil
}

func (repo *sessionStorage) Create(ctx context.Context, userID uint) (string, error) {
	UUID, err := uuidlib.NewUUID()
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	uuid := UUID.String()

	err = repo.db.WithContext(ctx).Create(&entity.Session{UUID: uuid, UserID: userID}).Error
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	return uuid, nil
}

func (repo *sessionStorage) Update(ctx context.Context, userID uint) (string, error) {
	UUID, err := uuidlib.NewUUID()
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	uuid := UUID.String()

	err = repo.db.WithContext(ctx).Model(&model.Session{}).Where("user_id = ?", userID).Update("uuid", uuid).Error
	if err != nil {
		return "", pkgerr.WithStack(err)
	}

	return uuid, nil
}

func (repo *sessionStorage) Delete(ctx context.Context, uuid string) error {
	if err := repo.db.WithContext(ctx).Delete(&entity.Session{UUID: uuid}).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return domainErrors.ErrRecordNotFound
		}
		return pkgerr.WithStack(err)
	}
	return nil
}

func sessionModelToDomain(session *model.Session) *entity.Session {
	return &entity.Session{
		UUID:   session.UUID,
		UserID: session.UserID,
	}
}
