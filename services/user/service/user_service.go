package service

import (
	"context"
	"errors"
	"time"

	userpb "github.com/SabinGhost19/go-micro-payment/proto/user"
	"github.com/SabinGhost19/go-micro-payment/services/user/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      repository.UserRepository
	jwtSecret string
	userpb.UnimplementedUserServiceServer
}

func NewUserService(repo repository.UserRepository, secret string) *UserService {
	return &UserService{repo: repo, jwtSecret: secret}
}

func (s *UserService) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.UserResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, errors.New("missing required field")
	}
	u := &repository.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return &userpb.UserResponse{
		UserId: u.ID, Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *UserService) AuthenticateUser(ctx context.Context, req *userpb.AuthenticateUserRequest) (*userpb.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	if s.jwtSecret == "" {
		return nil, errors.New("invalid JWT secret")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	sToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}
	return &userpb.AuthResponse{AccessToken: sToken, UserId: user.ID}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &userpb.UserResponse{
		UserId: user.ID, Email: user.Email, Name: user.Name, CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}
