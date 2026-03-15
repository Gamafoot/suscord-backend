package storage

import (
	"context"
	"strings"
	"suscord/internal/domain/entity"
	domainErrors "suscord/internal/domain/errors"
	"suscord/internal/infra/database/relational/model"

	pkgerr "github.com/pkg/errors"
	"gorm.io/gorm"
)

type userStorage struct {
	db *gorm.DB
}

func NewUserStorage(db *gorm.DB) *userStorage {
	return &userStorage{db: db}
}

func (s *userStorage) GetByID(ctx context.Context, userID uint) (entity.User, error) {
	empty := entity.User{}
	user := new(model.User)
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return empty, domainErrors.ErrRecordNotFound
		}
		return empty, pkgerr.WithStack(err)
	}
	return userModelToEntity(*user), nil
}

func (s *userStorage) GetByUsername(ctx context.Context, username string) (entity.UnsafeUser, error) {
	empty := entity.UnsafeUser{}
	user := new(model.User)
	if err := s.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return empty, domainErrors.ErrRecordNotFound
		}
		return empty, pkgerr.WithStack(err)
	}
	return unsafeUserModelToEntity(*user), nil
}

func (s *userStorage) SearchUsers(ctx context.Context, exceptUserID uint, username string) ([]entity.User, error) {
	users := make([]*model.User, 0)
	if err := s.db.WithContext(ctx).
		Order("username ASC").
		Find(&users, "id != ? AND LOWER(username) LIKE LOWER(?) ESCAPE '\\'", exceptUserID, "%"+escapeUserSearchLike(username)+"%").
		Error; err != nil {
		return nil, pkgerr.WithStack(err)
	}

	result := make([]entity.User, len(users))

	for i, user := range users {
		result[i] = userModelToEntity(*user)
	}

	return result, nil
}

func (s *userStorage) Create(ctx context.Context, user entity.UnsafeUser) error {
	userModel := unsafeUserEntityToModel(user)
	if err := s.db.WithContext(ctx).Create(&userModel).Error; err != nil {
		return pkgerr.WithStack(err)
	}
	return nil
}

func (s *userStorage) Update(ctx context.Context, userID uint, data map[string]interface{}) error {
	err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(data).Error
	if err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return domainErrors.ErrRecordNotFound
		}
		return pkgerr.WithStack(err)
	}
	return nil
}

func (s *userStorage) Delete(ctx context.Context, userID uint) error {
	if err := s.db.WithContext(ctx).Delete(&entity.User{ID: userID}).Error; err != nil {
		if pkgerr.Is(err, gorm.ErrRecordNotFound) {
			return domainErrors.ErrRecordNotFound
		}
		return pkgerr.WithStack(err)
	}
	return nil
}

func userModelToEntity(user model.User) entity.User {
	return entity.User{
		ID:         user.ID,
		Username:   user.Username,
		AvatarPath: user.AvatarPath,
	}
}

func unsafeUserModelToEntity(user model.User) entity.UnsafeUser {
	return entity.UnsafeUser{
		ID:         user.ID,
		Username:   user.Username,
		Password:   user.Password,
		AvatarPath: user.AvatarPath,
	}
}

func unsafeUserEntityToModel(user entity.UnsafeUser) model.User {
	return model.User{
		ID:         user.ID,
		Username:   user.Username,
		Password:   user.Password,
		AvatarPath: user.AvatarPath,
	}
}

func escapeUserSearchLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}
