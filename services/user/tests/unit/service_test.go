package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SabinGhost19/go-micro-payment/proto/user"
	"github.com/SabinGhost19/go-micro-payment/services/user/repository"
	user "github.com/SabinGhost19/go-micro-payment/services/user/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	users map[string]*repository.User
}

func (r *fakeUserRepository) Create(ctx context.Context, user *repository.User) error {
	if _, exists := r.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	user.ID = uuid.New().String()
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)
	user.CreatedAt = time.Now()
	r.users[user.Email] = user
	return nil
}

func (r *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *fakeUserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func TestRegisterUser(t *testing.T) {
	repo := &fakeUserRepository{users: make(map[string]*repository.User)}
	secret := "test-secret"
	service := user.NewUserService(repo, secret)

	t.Run("successful registration", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.RegisterUserRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		resp, err := service.RegisterUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.UserId)
		assert.Equal(t, req.Email, resp.Email)
		assert.Equal(t, req.Name, resp.Name)
		assert.NotEmpty(t, resp.CreatedAt)

		// verify user was stored with hashed password
		storedUser, err := repo.GetByEmail(ctx, req.Email)
		require.NoError(t, err)
		assert.Equal(t, resp.UserId, storedUser.ID)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(req.Password)))
	})

	t.Run("missing required field", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.RegisterUserRequest{
			Email:    "",
			Password: "password123",
			Name:     "Test User",
		}

		resp, err := service.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "missing required field", err.Error())
	})

	t.Run("duplicate email", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.RegisterUserRequest{
			Email:    "test@example.com",
			Password: "password123",
			Name:     "Test User",
		}

		// pre-populate user
		repo.users[req.Email] = &repository.User{Email: req.Email, ID: uuid.New().String()}

		resp, err := service.RegisterUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "user already exists", err.Error())
	})
}

func TestAuthenticateUser(t *testing.T) {
	repo := &fakeUserRepository{users: make(map[string]*repository.User)}
	secret := "test-secret"
	service := user.NewUserService(repo, secret)

	t.Run("successful authentication", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.AuthenticateUserRequest{
			Email:    "mock@ghosty.com",
			Password: "password1222222",
		}

		// pre-populate user with hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		require.NoError(t, err)
		userID := uuid.New().String()
		repo.users[req.Email] = &repository.User{
			ID:       userID,
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		resp, err := service.AuthenticateUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID, resp.UserId)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("invalid email", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.AuthenticateUserRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		resp, err := service.AuthenticateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("invalid password", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.AuthenticateUserRequest{
			Email:    "test@example.com",
			Password: "wrong---password",
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		require.NoError(t, err)
		repo.users[req.Email] = &repository.User{
			ID:       uuid.New().String(),
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		resp, err := service.AuthenticateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("jwt signing error", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.AuthenticateUserRequest{
			Email:    "gabyy@oiiok.com",
			Password: "password123",
		}

		// use an invalid secret to force signing error
		service := user.NewUserService(repo, "")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		require.NoError(t, err)
		repo.users[req.Email] = &repository.User{
			ID:       uuid.New().String(),
			Email:    req.Email,
			Password: string(hashedPassword),
		}

		resp, err := service.AuthenticateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGetUser(t *testing.T) {
	repo := &fakeUserRepository{users: make(map[string]*repository.User)}
	secret := "test-secret"
	service := user.NewUserService(repo, secret)

	t.Run("successful get user", func(t *testing.T) {
		ctx := context.Background()
		userID := uuid.New().String()
		email := "gofy@sikhdlopaaa.com"
		req := &userpb.GetUserRequest{UserId: userID}

		repo.users[email] = &repository.User{
			ID:        userID,
			Email:     email,
			Name:      "Test User",
			CreatedAt: time.Now(),
		}

		resp, err := service.GetUser(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID, resp.UserId)
		assert.Equal(t, email, resp.Email)
		assert.Equal(t, "Test User", resp.Name)
		assert.NotEmpty(t, resp.CreatedAt)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx := context.Background()
		req := &userpb.GetUserRequest{UserId: uuid.New().String()}

		resp, err := service.GetUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "user not found", err.Error())
	})
}
