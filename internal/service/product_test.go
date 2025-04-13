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

func TestProductService_Create(t *testing.T) {
	log.SetOutput(io.Discard)
	var (
		arbitraryErr = errors.New("arbitrary error")
		ctx          = context.Background()
		pointID      = uuid.New()
		receptionID  = uuid.New()
		productType  = entity.ProductTypeClothes
		timestamp    = time.Now()
	)

	product := entity.Product{
		ID:          uuid.New(),
		ReceptionID: receptionID,
		CreatedAt:   timestamp,
		Type:        productType,
	}

	type MockBehavior func(p *repomocks.MockProduct, r *repomocks.MockReception, m *metricmocks.MockCounter)

	for _, tc := range []struct {
		name         string
		mockBehavior MockBehavior
		want         entity.Product
		wantErr      error
	}{
		{
			name: "success",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().Create(ctx, receptionID, productType).Return(product, nil)
				m.EXPECT().Inc()
			},
			want: product,
		},
		{
			name: "active reception not found",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, repository.ErrNotFound)
			},
			wantErr: service.ErrActiveReceptionNotFound,
		},
		{
			name: "cannot get reception id",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(uuid.Nil, arbitraryErr)
			},
			wantErr: service.ErrCannotCreateProduct,
		},
		{
			name: "cannot create product",
			mockBehavior: func(p *repomocks.MockProduct, r *repomocks.MockReception, m *metricmocks.MockCounter) {
				r.EXPECT().GetActiveID(ctx, pointID).Return(receptionID, nil)
				p.EXPECT().Create(ctx, receptionID, productType).Return(entity.Product{}, arbitraryErr)
			},
			wantErr: service.ErrCannotCreateProduct,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mockProductRepo := repomocks.NewMockProduct(ctrl)
			mockReceptionRepo := repomocks.NewMockReception(ctrl)
			mockProductCounter := metricmocks.NewMockCounter(ctrl)

			tc.mockBehavior(mockProductRepo, mockReceptionRepo, mockProductCounter)

			s := service.NewProductService(mockProductRepo, mockReceptionRepo, mockProductCounter)

			got, err := s.Create(ctx, pointID, productType)

			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}
