package service

import (
	"context"
	"suscord/internal/config"
	"suscord/internal/domain/entity"
	domainErrors "suscord/internal/domain/errors"
	"suscord/internal/domain/storage"
	"suscord/pkg/hash"

	pkgerr "github.com/pkg/errors"
	"go.uber.org/zap"
)

type AuthService interface {
	Login(ctx context.Context, input *entity.LoginOrCreateInput) (string, error)
}

type authService struct {
	config  *config.Config
	storage storage.Storage
	logger  *zap.SugaredLogger
}

func NewAuthService(config *config.Config, storage storage.Storage, logger *zap.SugaredLogger) *authService {
	return &authService{
		config:  config,
		storage: storage,
		logger:  logger,
	}
}

func (s *authService) Login(ctx context.Context, input *entity.LoginOrCreateInput) (string, error) {
	log := s.logger.With(
		"username", input.Username,
	)

	hasher := hash.NewSHA1Hasher(s.config.Secure.Hash.Salt)

	user, err := s.storage.User().GetByUsername(ctx, input.Username)
	if err != nil {
		log.Infow("user not found, creating")

		hash, err := hasher.Hash(input.Password)
		if err != nil {
			return "", pkgerr.WithStack(err)
		}

		err = s.storage.User().Create(ctx, entity.UnsafeUser{
			Username: input.Username,
			Password: hash,
		})
		if err != nil {
			return "", pkgerr.WithStack(err)
		}

		user, err = s.storage.User().GetByUsername(ctx, input.Username)
		if err != nil {
			if pkgerr.Is(err, domainErrors.ErrRecordNotFound) {
				log.Debugw("created user not found after create")
				return "", domainErrors.ErrInvalidLoginOrPassword
			}
			return "", pkgerr.WithStack(err)
		}

		log.Infow(
			"user created",
			"user_id", user.ID,
		)
	} else {
		hash, err := hasher.Hash(input.Password)
		if err != nil {
			return "", pkgerr.WithStack(err)
		}

		if user.Password != hash {
			log.Debugw(
				"invalid password",
				"user_id", user.ID,
			)
			return "", domainErrors.ErrInvalidLoginOrPassword
		}
	}

	log.Infow(
		"user authenticated",
		"user_id", user.ID,
	)

	return s.createSession(ctx, user.ID)
}

func (s *authService) createSession(ctx context.Context, userID uint) (string, error) {
	var uuid string

	log := s.logger.With(
		"user_id", userID,
	)

	_, err := s.storage.Session().GetByUserID(ctx, userID)
	if err != nil {
		if pkgerr.Is(err, domainErrors.ErrRecordNotFound) {
			uuid, err = s.storage.Session().Create(ctx, userID)
			if err != nil {
				return "", err
			}
			log.Infow("session created")
			return uuid, nil
		}
		return "", err
	} else {
		uuid, err = s.storage.Session().Update(ctx, userID)
		if err != nil {
			return "", err
		}
		log.Infow("session updated")
	}

	return uuid, nil
}
