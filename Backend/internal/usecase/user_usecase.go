package usecase

import (
	"context"
	"errors"

	"server/internal/entity"
	"server/internal/repository"
)

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

func (uc *userUseCase) GetUser(ctx context.Context, id uint) (*entity.User, error) {
	user, err := uc.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	
	// Не возвращаем пароль
	user.Password = ""
	return user, nil
}

func (uc *userUseCase) UpdateUser(ctx context.Context, user *entity.User) error {
	// Проверяем, существует ли пользователь
	existingUser, err := uc.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return errors.New("user not found")
	}
	
	// Обновляем только разрешенные поля
	existingUser.Email = user.Email
	// Пароль обновляется через отдельный метод
	// Яндекс токены обновляются через storage use case
	
	return uc.userRepo.UpdateUser(ctx, existingUser)
}