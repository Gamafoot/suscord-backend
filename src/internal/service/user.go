package service

import (
	"context"
	"suscord/internal/domain/entity"
	"suscord/internal/domain/event"
	"suscord/internal/domain/eventbus"
	"suscord/internal/domain/storage"
	file "suscord/internal/infra/file_manager"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
)

type UserService interface {
	GetByID(ctx context.Context, userID uint) (entity.User, error)
	SearchUsers(ctx context.Context, userID uint, username string) ([]entity.User, error)
	Update(ctx context.Context, userID uint, input entity.UpdateUserInput) (entity.User, error)
}

type userService struct {
	storage     storage.Storage
	fileManager file.FileManager
	eventbus    eventbus.EventBus
	logger      *zap.SugaredLogger
}

func NewUserService(
	storage storage.Storage,
	fileManager file.FileManager,
	eventbus eventbus.EventBus,
	logger *zap.SugaredLogger,
) *userService {
	return &userService{
		storage:     storage,
		fileManager: fileManager,
		eventbus:    eventbus,
		logger:      logger,
	}
}

func (s *userService) GetByID(ctx context.Context, userID uint) (entity.User, error) {
	log := s.logger.With(
		"user_id", userID,
	)

	user, err := s.storage.User().GetByID(ctx, userID)
	if err != nil {
		return entity.User{}, pkgerr.Wrap(err, "failed to get user")
	}

	log.Debugw(
		"get user by id",
		"user.username", user.Username,
	)

	return user, nil
}

func (s *userService) SearchUsers(ctx context.Context, userID uint, username string) ([]entity.User, error) {
	log := s.logger.With(
		"user_id", userID,
		"username", username,
	)

	users, err := s.storage.User().SearchUsers(ctx, userID, username)
	if err != nil {
		return nil, err
	}

	log.Infow(
		"search users",
		"count", len(users),
	)

	return users, nil
}

func (s *userService) Update(ctx context.Context, userID uint, input entity.UpdateUserInput) (entity.User, error) {
	empty := entity.User{}

	log := s.logger.With(
		"user_id", userID,
	)

	var (
		filepath string
		err      error
	)

	if input.File != nil {
		log.Debugw(
			"uploading new user avatar",
			"filename", input.File.Name,
			"size", input.File.Size,
		)

		filepath, err = s.fileManager.Upload(input.File, "users/avatars")
		if err != nil {
			return empty, pkgerr.WithStack(err)
		}
	}

	data := make(map[string]any)

	if input.Username != nil {
		data["username"] = input.Username
	}
	if len(filepath) > 0 {
		data["avatar_path"] = filepath
	}

	err = s.storage.User().Update(ctx, userID, data)
	if err != nil {
		return empty, err
	}

	user, err := s.storage.User().GetByID(ctx, userID)
	if err != nil {
		return empty, err
	}

	payload := event.NewUserUpdated(user)
	s.eventbus.Publish(payload.EventName(), payload)

	return user, nil
}
