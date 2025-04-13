package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/repository"
	"github.com/spanwalla/pvz/pkg/hasher"
)

var (
	ErrCannotGenerateToken = errors.New("cannot generate token")
	ErrCannotAcceptToken   = errors.New("cannot accept this token")
	ErrTokenExpired        = errors.New("token is expired")
	ErrUserNotFound        = errors.New("user not found")
	ErrCannotGetUser       = errors.New("cannot get user")
	ErrWrongPassword       = errors.New("wrong password")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrCannotRegisterUser  = errors.New("cannot register user")
)

type AuthService struct {
	userRepo       repository.User
	passwordHasher hasher.PasswordHasher
	clock          clockwork.Clock
	secretKey      string
	tokenTTL       time.Duration
}

func NewAuthService(userRepo repository.User, passwordHasher hasher.PasswordHasher,
	clock clockwork.Clock, secretKey string, ttl time.Duration) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		clock:          clock,
		secretKey:      secretKey,
		tokenTTL:       ttl,
	}
}

func (s *AuthService) DummyLogin(_ context.Context, role entity.RoleType) (string, error) {
	token, err := s.generateToken(uuid.New(), role)
	if err != nil {
		log.Errorf("AuthService.DummyLogin - s.generateToken: %v", err)
		return "", ErrCannotGenerateToken
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrUserNotFound
		}

		log.Errorf("AuthService.Login - s.usersRepo.GetByEmail: %v", err)
		return "", ErrCannotGetUser
	}

	if !s.passwordHasher.Match(password, user.Password) {
		return "", ErrWrongPassword
	}

	token, err := s.generateToken(user.ID, user.Role)
	if err != nil {
		log.Errorf("AuthService.Login - s.generateToken: %v", err)
		return "", ErrCannotGenerateToken
	}

	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password string, role entity.RoleType) (RegisterOutput, error) {
	hashedPassword, err := s.passwordHasher.Hash(password)
	if err != nil {
		log.Errorf("AuthService.Register - s.passwordHasher.Hash: %v", err)
		return RegisterOutput{}, ErrCannotRegisterUser
	}

	user, err := s.userRepo.Create(ctx, email, hashedPassword, role)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return RegisterOutput{}, ErrUserAlreadyExists
		}

		log.Errorf("AuthService.Register - s.userRepo.Create: %v", err)
		return RegisterOutput{}, ErrCannotRegisterUser
	}

	return RegisterOutput{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *AuthService) ParseToken(token string) (*entity.TokenClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &entity.TokenClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(s.secretKey), nil
	}, jwt.WithExpirationRequired(), jwt.WithLeeway(time.Second*3))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}

		log.Errorf("AuthService.ParseToken - jwt.ParseWithClaims: %v", err)
		return nil, ErrCannotAcceptToken
	}

	claims, ok := jwtToken.Claims.(*entity.TokenClaims)
	if !ok {
		log.Error("AuthService.ParseToken: unsuccessful cast to custom claims")
		return nil, ErrCannotAcceptToken
	}

	return claims, nil
}

func (s *AuthService) generateToken(userID uuid.UUID, role entity.RoleType) (string, error) {
	now := s.clock.Now()

	return jwt.NewWithClaims(jwt.SigningMethodHS256, &entity.TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString([]byte(s.secretKey))
}
