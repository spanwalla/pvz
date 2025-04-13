package service_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spanwalla/pvz/internal/entity"
	metricmocks "github.com/spanwalla/pvz/internal/metrics/mocks"
	"github.com/spanwalla/pvz/internal/repository"
	repomocks "github.com/spanwalla/pvz/internal/repository/mocks"
	"github.com/spanwalla/pvz/internal/service"
)

func TestReceptionService_Create(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		pointID      = uuid.New()
		timestamp    = time.Now()
	)

	reception := entity.Reception{
		ID:        uuid.New(),
		PointID:   pointID,
		CreatedAt: timestamp,
		Status:    entity.ReceptionStatusInProgress,
	}

	type MockBehavior func(r *repomocks.MockReception, m *metricmocks.MockCounter)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         entity.Reception
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().Create(ctx, pointID).Return(reception, nil)
				m.EXPECT().Inc()
			},
			want: reception,
		},
		{
			name: "reception already opened",
			mockBehavior: func(r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().Create(ctx, pointID).Return(entity.Reception{}, repository.ErrAlreadyExists)
			},
			wantErr: service.ErrReceptionAlreadyOpened,
		},
		{
			name: "cannot create reception",
			mockBehavior: func(r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().Create(ctx, pointID).Return(entity.Reception{}, arbitraryErr)
			},
			wantErr: service.ErrCannotCreateReception,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockReceptionCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockReceptionRepo, mockReceptionCounter)

			s := service.NewReceptionService(mockReceptionRepo, mockReceptionCounter)

			got, err := s.Create(ctx, pointID)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}
