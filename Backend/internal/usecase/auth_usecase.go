package usecase

import (
	"context"
	"errors"
	
	"golang.org/x/crypto/bcrypt"
	
	"server/internal/entity"
	"server/internal/repository"
	"server/pkg/auth"
)

type authUseCase struct {
	userRepo  repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewAuthUseCase(userRepo repository.UserRepository, jwtManager *auth.JWTManager) AuthUseCase {
	return &authUseCase{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (uc *authUseCase) Register(ctx context.Context, email, password string) error {
	// Проверяем, существует ли пользователь
	existingUser, _ := uc.userRepo.GetUserByEmail(ctx, email)
	if existingUser != nil {
		return errors.New("user already exists")
	}
	
	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user := &entity.User{
		Email:    email,
		Password: string(hashedPassword),
	}
	
	return uc.userRepo.CreateUser(ctx, user)
}

func (uc *authUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	
	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	
	// Генерируем JWT токен
	token, err := uc.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}
	
	return token, nil
}

func (uc *authUseCase) ValidateToken(ctx context.Context, token string) (uint, error) {
	claims, err := uc.jwtManager.ValidateToken(token)
	if err != nil {
		return 0, err
	}
	
	// Проверяем, существует ли пользователь
	_, err = uc.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return 0, errors.New("user not found")
	}
	
	return claims.UserID, nil
}