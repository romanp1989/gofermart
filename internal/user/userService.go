package user

import (
	"context"
	"github.com/romanp1989/gophermart/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userStorage interface {
	CreateUser(ctx context.Context, user domain.User) (*domain.User, error)
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
}

type Service struct {
	storage userStorage
	log     *zap.Logger
}

func NewService(userStore userStorage, log *zap.Logger) *Service {
	return &Service{
		storage: userStore,
		log:     log,
	}
}

func (s *Service) CreateUser(ctx context.Context, userReg domain.User) (*domain.User, error) {
	var user *domain.User

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userReg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	userReg.Password = string(hashedPassword)
	user, err = s.storage.CreateUser(ctx, userReg)
	if err != nil {
		return nil, err

	}

	return user, nil
}

func (s *Service) Authorization(ctx context.Context, userReg *domain.User) (*domain.User, error) {
	user, err := s.storage.FindByLogin(ctx, userReg.Login)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userReg.Password))
	if err != nil {
		return nil, err
	}

	return user, nil
}
