package service_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/repository"
	repomocks "github.com/spanwalla/pvz/internal/repository/mocks"
	"github.com/spanwalla/pvz/internal/service"
	"github.com/spanwalla/pvz/pkg/hasher"
)

func TestAuthService_DummyLogin(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		startTime = lo.Must(time.Parse(time.RFC3339, "2025-04-13T10:00:00Z"))
		ctx       = context.Background()
		secretKey = "secret"
		role      = entity.RoleTypeEmployee
		tokenTTL  = time.Minute
	)

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)

		mockUserRepo := repomocks.NewMockUser(ctrl)
		mockPasswordHasher := hasher.NewMockPasswordHasher(ctrl)
		mockClock := clockwork.NewFakeClockAt(startTime)

		s := service.NewAuthService(mockUserRepo, mockPasswordHasher, mockClock, secretKey, tokenTTL)

		got, err := s.DummyLogin(ctx, role)

		assert.ErrorIs(t, err, nil)
		assert.NotEmpty(t, got)
	})
}

func TestAuthService_Login(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		startTime    = lo.Must(time.Parse(time.RFC3339, "2025-04-13T10:00:00Z"))
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		email        = "test@mail.ru"
		password     = "12TestMark"
		secretKey    = "secret"
		tokenTTL     = time.Minute
	)

	user := entity.User{
		ID:       uuid.New(),
		Email:    email,
		Password: password,
		Role:     entity.RoleTypeEmployee,
	}

	token := lo.Must(jwt.NewWithClaims(jwt.SigningMethodHS256, &entity.TokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(startTime.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(startTime),
		},
	}).SignedString([]byte(secretKey)))

	type MockBehavior func(u *repomocks.MockUser, h *hasher.MockPasswordHasher)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         string
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				u.EXPECT().GetByEmail(ctx, email).Return(user, nil)
				h.EXPECT().Match(password, user.Password).Return(true)
			},
			want: token,
		},
		{
			name: "user not found",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				u.EXPECT().GetByEmail(ctx, email).Return(entity.User{}, repository.ErrNotFound)
			},
			wantErr: service.ErrUserNotFound,
		},
		{
			name: "cannot get user",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				u.EXPECT().GetByEmail(ctx, email).Return(entity.User{}, arbitraryErr)
			},
			wantErr: service.ErrCannotGetUser,
		},
		{
			name: "wrong password",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				u.EXPECT().GetByEmail(ctx, email).Return(user, nil)
				h.EXPECT().Match(password, user.Password).Return(false)
			},
			wantErr: service.ErrWrongPassword,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockUserRepo := repomocks.NewMockUser(ctrl)
			mockPasswordHasher := hasher.NewMockPasswordHasher(ctrl)
			mockClock := clockwork.NewFakeClockAt(startTime)

			tc.mockBehavior(mockUserRepo, mockPasswordHasher)

			s := service.NewAuthService(mockUserRepo, mockPasswordHasher, mockClock, secretKey, tokenTTL)

			got, err := s.Login(ctx, email, password)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		startTime    = lo.Must(time.Parse(time.RFC3339, "2025-04-13T10:00:00Z"))
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		email        = "test@mail.ru"
		password     = "12TestMark"
		role         = entity.RoleTypeModerator
		secretKey    = "secret"
		tokenTTL     = time.Minute
	)

	user := entity.User{
		ID:       uuid.New(),
		Email:    email,
		Password: "hashed_password",
		Role:     role,
	}

	type MockBehavior func(u *repomocks.MockUser, h *hasher.MockPasswordHasher)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         service.RegisterOutput
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				h.EXPECT().Hash(password).Return(user.Password, nil)
				u.EXPECT().Create(ctx, email, user.Password, role).Return(user, nil)
			},
			want: service.RegisterOutput{
				ID:    user.ID,
				Email: user.Email,
				Role:  user.Role,
			},
		},
		{
			name: "cannot hash password",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				h.EXPECT().Hash(password).Return("", arbitraryErr)
			},
			wantErr: service.ErrCannotRegisterUser,
		},
		{
			name: "user already exists",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				h.EXPECT().Hash(password).Return(user.Password, nil)
				u.EXPECT().Create(ctx, user.Email, user.Password, user.Role).Return(entity.User{}, repository.ErrAlreadyExists)
			},
			wantErr: service.ErrUserAlreadyExists,
		},
		{
			name: "cannot create user",
			mockBehavior: func(u *repomocks.MockUser, h *hasher.MockPasswordHasher) {
				h.EXPECT().Hash(password).Return(user.Password, nil)
				u.EXPECT().Create(ctx, user.Email, user.Password, user.Role).Return(entity.User{}, arbitraryErr)
			},
			wantErr: service.ErrCannotRegisterUser,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockUserRepo := repomocks.NewMockUser(ctrl)
			mockPasswordHasher := hasher.NewMockPasswordHasher(ctrl)
			mockClock := clockwork.NewFakeClockAt(startTime)

			tc.mockBehavior(mockUserRepo, mockPasswordHasher)

			s := service.NewAuthService(mockUserRepo, mockPasswordHasher, mockClock, secretKey, tokenTTL)

			got, err := s.Register(ctx, email, password, role)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAuthService_ParseToken(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		issuedTimeValid   = time.Now()
		issuedTimeExpired = issuedTimeValid.Add(-time.Hour)
		userID            = lo.Must(uuid.Parse("2864e043-95b7-42e1-8201-9b0fc2f7e0c1"))
		role              = entity.RoleTypeModerator
		secretKey         = "secret"
		tokenTTL          = time.Minute * 10
	)

	validClaims := entity.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(issuedTimeValid.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(issuedTimeValid),
		},
		UserID: userID,
		Role:   role,
	}

	expiredClaims := entity.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(issuedTimeExpired.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(issuedTimeExpired.Add(tokenTTL)),
		},
		UserID: userID,
		Role:   role,
	}

	validToken := lo.Must(jwt.NewWithClaims(jwt.SigningMethodHS256, &validClaims).SignedString([]byte(secretKey)))
	diffSecretKeyToken := lo.Must(jwt.NewWithClaims(jwt.SigningMethodHS256, &validClaims).
		SignedString([]byte(secretKey + "a")))
	expiredToken := lo.Must(jwt.NewWithClaims(jwt.SigningMethodHS256, &expiredClaims).
		SignedString([]byte(secretKey)))
	diffSigningMethodToken := lo.Must(jwt.NewWithClaims(jwt.SigningMethodES256, &validClaims).
		SignedString(lo.Must(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))))

	for _, tc := range []struct {
		name    string
		token   string
		want    *entity.TokenClaims
		wantErr error
	}{
		{
			name:  "success",
			token: validToken,
			want:  &validClaims,
		},
		{
			name:    "diff secret key",
			token:   diffSecretKeyToken,
			wantErr: service.ErrCannotAcceptToken,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			wantErr: service.ErrTokenExpired,
		},
		{
			name:    "diff signing method",
			token:   diffSigningMethodToken,
			wantErr: service.ErrCannotAcceptToken,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockUserRepo := repomocks.NewMockUser(ctrl)
			mockPasswordHasher := hasher.NewMockPasswordHasher(ctrl)
			mockClock := clockwork.NewFakeClock()

			s := service.NewAuthService(mockUserRepo, mockPasswordHasher, mockClock, secretKey, tokenTTL)

			got, err := s.ParseToken(tc.token)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}
